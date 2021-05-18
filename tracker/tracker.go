package tracker

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/timedb/wheatDFS/app"
	"github.com/timedb/wheatDFS/etc"
	"github.com/timedb/wheatDFS/log"
	"github.com/timedb/wheatDFS/serverTorch"
	"math/rand"
	"os"
	"sync"
	"time"
)

var oneTracker sync.Once //单实例
var sysServer *Server
var WriteOnes sync.Once

func test() {
	for {
		time.Sleep(time.Second * 5)
		fmt.Println("----------tracker-------")
		for _, val := range sysServer.AnyTrackerHosts() {
			fmt.Println(val)
		}

		fmt.Println("--------storage----------")
		for _, val := range sysServer.AnyStorageHosts() {
			fmt.Println(val)
		}

		fmt.Println("--------vote-----------")
		fmt.Println(sysServer.getState())
	}

}

// Server 服务端
type Server struct {
	LocalHost *etc.Addr    //本地地址
	localLock sync.RWMutex //local读写锁

	StateSole  int          //服务端角色
	leaderHost *etc.Addr    //leader的地址
	voteLock   sync.RWMutex //选举锁

	service *TraRpcService //服务端服务接口

	TrackerHosts []*etc.Addr //Tracker集群地址
	StorageHosts []*etc.Addr //Storage集群地址

	traLock sync.RWMutex //tracker锁
	stoLock sync.RWMutex //storage锁

	startVoteChan chan bool //开始选举标准

	syncFileChan chan *app.EsoData //storage同步

	//秒传实现
	blot *bolt.DB //分布式事务器

}

// MakeServer 创建Server
func MakeServer() *Server {
	oneTracker.Do(func() {
		ser := new(Server)
		//获取本地ip
		ip, err := serverTorch.GetIPv4s()
		if err != nil {
			log.SysLog.Add(fmt.Sprintf("%v", err), log.Panic)
		}

		//获取本地ip
		port, err := serverTorch.GetAvailablePort()
		if err != nil {
			log.SysLog.Add(fmt.Sprintf("%v", err), log.Panic)
		}

		//判断Tracker角色
		var state int
		if etc.SysConf.TrackerConf.CandidateIp == ip { //初始ip以及本地ip相同
			state = stateCandidate
			port = etc.SysConf.TrackerConf.CandidatePort //修改端口为Candidate端口
		} else {
			state = stateFollower
		}

		var addr *etc.Addr
		if ip == etc.SysConf.TrackerConf.CandidateIp { //本机为初始机
			addr = etc.SysConf.TrackerConf.CandidateHost
		} else {
			addr, err = etc.MakeAddr(ip, port, etc.StateTracker) //定义ip角色为Tracker, 并且自动生成权重
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

		}

		ser.LocalHost = addr //绑定本地地址
		ser.StateSole = state
		ser.service = new(TraRpcService)
		ser.leaderHost = nil

		ser.startVoteChan = make(chan bool, etc.SysConf.StorageConf.MaxStorageCount) //最大接入数

		//新建库
		db, err := bolt.Open(etc.SysConf.TrackerConf.EsotericaPath, 0600, nil) //打开库
		if err != nil {
			log.SysLog.Add(err.Error(), log.Panic) //
			log.SysLog.Exit()
		}

		ser.blot = db //绑定db

		//创建esoterica表
		{
			err = ser.blot.Update(func(tx *bolt.Tx) error {

				//判断要创建的表是否存在
				b := tx.Bucket([]byte("esoterica"))
				if b == nil {
					//创建叫"MyBucket"的表
					_, err := tx.CreateBucket([]byte("esoterica"))
					if err != nil {
						//也可以在这里对表做插入操作
						log.SysLog.Exit() // 退出程序
					}
				}

				return nil
			})

		}

		//创建同步用管道
		ser.syncFileChan = make(chan *app.EsoData, etc.SysConf.TrackerConf.SyncMaxCount) //创建同步用管道

		sysServer = ser
		etc.SysLocalHost = addr
		//启动维护服务
		go sysServer.work()

	})

	return sysServer //返回sysServer
}

//修改localHosts
func (s *Server) putLocalHosts(l *etc.Addr) {
	s.localLock.Lock()
	defer s.localLock.Unlock()
	s.LocalHost = l
}

//读取localHosts
func (s *Server) getLocalHosts() *etc.Addr {
	s.localLock.RLock()
	defer s.localLock.RUnlock()
	return s.LocalHost
}

// PutTrackerHost 添加一个Tra节点
func (s *Server) PutTrackerHost(address ...*etc.Addr) {

	s.stoLock.Lock()
	defer s.stoLock.Unlock()
	//防止重复加入
loop:
	for _, addr := range address {
		for _, host := range s.TrackerHosts {
			if host.GetAddress() == addr.GetAddress() || s.getLeader().GetAddress() == addr.GetAddress() {
				continue loop
			}
		}

		s.TrackerHosts = append(s.TrackerHosts, addr)

		//注册到pool中
	}

}

// DeleteInvalidTraHost 删除无效Tra节点
func (s *Server) DeleteInvalidTraHost(examine func(addr *etc.Addr) bool) {
	hosts := make([]*etc.Addr, 0, len(s.TrackerHosts))
	for _, addr := range s.AnyTrackerHosts() {
		if examine(addr) && addr.GetAddress() != s.getLeader().GetAddress() { //防止Leader在同步器中
			hosts = append(hosts, addr) //有效节点
		}
	}

	s.traLock.Lock()
	defer s.traLock.Unlock()
	s.TrackerHosts = hosts

}

// AnyTrackerHosts 获取全部的tracker hosts
func (s *Server) AnyTrackerHosts() []*etc.Addr {
	s.traLock.RLock()
	defer s.traLock.RUnlock()

	src := make([]*etc.Addr, len(s.TrackerHosts))

	copy(src, s.TrackerHosts)

	return src

}

// PutStorageHost 添加一个sto节点
func (s *Server) PutStorageHost(address ...*etc.Addr) {
	s.stoLock.Lock()
	defer s.stoLock.Unlock()

	addrPut := make([]*etc.Addr, 0, len(address))

loop:
	for _, addr := range address {
		for _, host := range s.StorageHosts {
			if host.GetAddress() == addr.GetAddress() {
				continue loop
			}
		}

		addrPut = append(addrPut, addr)
	}

	s.StorageHosts = append(s.StorageHosts, addrPut...)

	//tracker注册
	if len(address) != 0 {
		WriteOnes.Do(func() {
			s.startVoteChan <- true
		})
	}
}

// UpdateLocalHost 升级成Leader
func (s *Server) UpdateLocalHost() {
	s.traLock.Lock()
	defer s.traLock.Unlock()

	s.updateVole(stateLeader, s.getLocalHosts()) //更新leader

	var index = -1
	for i, val := range s.TrackerHosts {
		if val.GetAddress() == s.getLocalHosts().GetAddress() {
			index = i
			break
		}
	}

	//没有找到自己
	if index == -1 {
		return
	}

	s.TrackerHosts = append(s.TrackerHosts[0:index], s.TrackerHosts[index+1:]...) //删除自己

}

// DeleteInvalidStoHost 删除无效的Sto节点
func (s *Server) DeleteInvalidStoHost(examine func(addr *etc.Addr) bool) {
	hosts := make([]*etc.Addr, 0, 5)
	for _, addr := range s.AnyStorageHosts() {
		if examine(addr) {
			hosts = append(hosts, addr) //有效节点
		}
	}

	s.stoLock.Lock()
	defer s.stoLock.Unlock()
	s.StorageHosts = hosts

}

// GetStorageHost 获取一个storage地址，负载均衡
func (s *Server) GetStorageHost() *etc.Addr {

	s.traLock.RLock()
	defer s.traLock.RUnlock()

	//计算完全权重
	stoWeightSum := 0
	for _, val := range s.StorageHosts {
		stoWeightSum += val.Weight
	}

	if stoWeightSum == 0 {
		return nil
	}

	index := rand.Intn(stoWeightSum) //获取随机权重
	stoWeightSum = 0

	for _, host := range s.StorageHosts {
		stoWeightSum += host.Weight
		if index < stoWeightSum {
			return host
		}
	}

	return nil //响应失败

}

// AnyStorageHosts  获取全部的tracker hosts
func (s *Server) AnyStorageHosts() []*etc.Addr {
	s.stoLock.RLock()
	defer s.stoLock.RUnlock()

	return s.StorageHosts

}

//更新角色状态
func (s *Server) updateVole(state int, leaderHosts *etc.Addr) {
	s.voteLock.Lock()
	defer s.voteLock.Unlock()
	s.StateSole = state
	s.leaderHost = leaderHosts

}

//获取leader
func (s *Server) getLeader() *etc.Addr {
	s.voteLock.RLock()
	defer s.voteLock.RUnlock()

	return s.leaderHost
}

//获取state
func (s *Server) getState() int {
	s.voteLock.RLock()
	defer s.voteLock.RUnlock()

	return s.StateSole
}

//App处理
//注册到leader上
func (s *Server) registerToLeader() {
	req := app.MakeRegisterTraReq()

	resp := new(app.TraRegisterTraResp)

	err := req.Do(s.getLeader(), resp)
	//注册成功
	if err == nil && resp.State == app.ResponseStateOK {

		//验证是否有效
		if !s.checkLeaderTracker() {
			log.SysLog.Add("leader is not return Response", log.Error)
		}

		return
	}

	log.SysLog.Add(resp.Err, log.Info)

}

// StartServer 启动服务
func (s *Server) StartServer() {

	//根据debug启动
	if etc.SysConf.Debug == true {
		go test()
	}

	//先进行注册
	if s.getLeader() == nil {
		s.updateVole(s.StateSole, etc.SysConf.TrackerConf.CandidateHost) //注册给初始创建者
	}

	//进行注册
	if s.getLocalHosts().GetAddress() != s.getLeader().GetAddress() {
		s.registerToLeader()
	}

	fmt.Println("bind host ", s.getLocalHosts().GetAddress())

	defer s.blot.Close() //程序崩溃后退出blot

	go s.syncFileWork() //启动文件同步

	s.service.start()
}

// checkLeaderTracker 检查leader状态
func (s *Server) checkLeaderTracker() bool {
	req := app.MakeTraHeartReq()
	resp := new(app.ResponseBase)

	if req.Do(s.getLeader(), resp); resp.State == app.ResponseStateOK {
		return true
	}

	return false

}

// 根据角色分配工作
func (s *Server) work() {
	ticker := time.NewTicker(WorkTime)
	leaderTicker := time.NewTicker(LeaderTime)
	deltHosts := time.NewTicker(WorkTime)

	//leader
	leader := func() {
		//选择最好的机子任命

		go s.updateStorageLeader() // 更新storage的leader

		cands := 0
		maxWeight := 0

		hosts := s.AnyTrackerHosts()
		for i, host := range hosts {
			if host.Weight > maxWeight {
				cands = i
				maxWeight = host.Weight
			}
		}

		//发送任命报告
		for i := 0; i < len(hosts); i++ {
			req := app.MakeTraAppointedRoleReq(stateFollower)
			resp := new(app.TraAppointedRoleResp)
			if cands == i {
				req.StateRole = stateCandidate
			}
			//发送任命报告
			err := req.Do(hosts[i], resp)
			if err == nil && resp.Successful() {
				log.SysLog.Add("send appointed task successful", log.Info)
			} else {
				log.SysLog.Add(resp.Err, log.Info)
			}
		}

	}

	follower := func() {

		//取消同步
		leaderHosts := s.getLeader()
		if leaderHosts == nil || leaderHosts.GetAddress() == s.getLocalHosts().GetAddress() {
			return
		}

		req := app.MakeTraReportTraHosts(s.AnyTrackerHosts(), s.AnyStorageHosts())
		resp := new(app.TraReportTraHostsResp)

		err := req.Do(s.getLeader(), resp)

		if err != nil {
			log.SysLog.Add(err.Error(), log.Error)
			return
		}

		//根据leader进行更新
		if resp.State != app.ResponseStateOK {
			log.SysLog.Add(resp.Err, log.Error)
			return
		}

		s.PutTrackerHost(resp.TraHosts...)
		s.PutStorageHost(resp.StoHosts...)

	}

	candidate := func() {

		//自动升级
		if s.getLeader().GetAddress() == s.getLocalHosts().GetAddress() {
			<-s.startVoteChan //第一个storage注册后
			s.updateVole(stateLeader, s.getLeader())
			log.SysLog.Add(fmt.Sprintf("%s upgrade is leader", s.getLocalHosts().GetAddress()), log.Info)
			leader() //调用一次leader
			return

		} else {
			//leader失效
			if !examine(s.getLeader()) {
				hosts := s.AnyTrackerHosts()

				//请求其他主机的同意
				for _, host := range hosts {
					//去掉leader
					if host.GetAddress() != s.getLeader().GetAddress() {
						req := app.MakeTraUpdateTheRoleReq()
						resp := new(app.TraUpdateTheRoleResp)
						req.Do(host, resp)
						if resp.State != app.ResponseStateOK {
							log.SysLog.Add("upgrade err", log.Info)
							return
						}

					}
				}

				//主机全部同意后，更改状态选择新继承者
				s.UpdateLocalHost() // 升级为leader
				log.SysLog.Add("update success to leader", log.Info)
				leader()
			}
		}

	}

	for {

		select {
		case <-ticker.C:

			if s.getState() == stateFollower {
				follower()
			} else if s.getState() == stateCandidate {
				candidate()
				follower()
			}
		case <-leaderTicker.C:
			if s.getState() == stateLeader {
				leader()
			}
		case <-deltHosts.C:
			//删除无效节点
			s.DeleteInvalidTraHost(examine)
			s.DeleteInvalidStoHost(examine)
		}

	}

}

// 添加一个秒传
func (s *Server) updateEsoterica(oes *app.EsoData) error {

	//防止重复写入
	newOes, err := s.viewEsoterica(oes.Hash)
	if err == nil {
		if oes.Status <= newOes.Status {
			return nil
		}
	}

	err = s.blot.Update(func(tx *bolt.Tx) error {

		//取出叫esoterica的表
		b := tx.Bucket([]byte("esoterica"))

		if b == nil {
			log.SysLog.Add("bucket nonexistent", log.Panic)
			return BucketErr
		}

		//写入
		b.Put([]byte(oes.Hash), oes.Encode()) //更新数据

		log.SysLog.Add(fmt.Sprintf("%s eso update state is %d", oes.Hash, oes.Status), log.Info)

		//同步到其他tracker上
		go s.pushAnyTracker(oes) //同步上去
		s.putSyncFileChan(oes)   //开始storage的同步

		return nil
	})

	return err

}

//读取一个秒传Hash
func (s *Server) viewEsoterica(hash string) (*app.EsoData, error) {

	//秒传结构体
	eso := new(app.EsoData)

	err := s.blot.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("esoterica"))
		if b == nil {
			log.SysLog.Add("bucket nonexistent", log.Panic)
			return BucketErr
		}

		content := b.Get([]byte(hash))
		if content == nil {
			return HashErr
		}

		err := eso.Decode(content)
		if err != nil {
			return err
		}

		return nil
	})

	return eso, err
}

//请求Leader的情况
func (s *Server) viewLeaderEso(hash string) (*app.EsoData, bool) {

	//对于leader不做检测
	if s.getState() == stateLeader || s.getLeader() == s.getLocalHosts() {
		return nil, false
	}

	req := app.MakeTraGetEsoDataReq(hash)
	resp := new(app.TraGetEsoDataResp)
	err := req.Do(s.getLeader(), resp)
	if err != nil {
		return nil, false
	}

	//失败
	if !resp.Successful() {
		return nil, false
	}

	return resp.Eso, true

}

//对其他的tracker发送秒传同步
func (s *Server) pushAnyTracker(eso *app.EsoData) {
	if s.getLeader().GetAddress() != s.getLocalHosts().GetAddress() {
		return
	}

	hosts := s.AnyTrackerHosts()

	// 推送秒传更新
	for _, addr := range hosts {
		req := app.MakeTraPutDataReq(eso.Hash, eso.Status, addr, eso.Ext)
		req.Eso = eso //修改
		resp := new(app.TraGetEsoDataResp)
		err := req.Do(addr, resp)

		//验证，写入日志
		if err != nil {
			log.SysLog.Add(err.Error(), log.Error)
		}

		if !resp.Successful() {
			log.SysLog.Add(resp.Err, log.Error)
		}

	}

}

//更新其他Storage的leader地址
func (s *Server) updateStorageLeader() {
	hosts := s.AnyStorageHosts()

	req := app.MakeStoUpdateLeader()
	resp := new(app.ResponseBase)

	for _, host := range hosts {
		err := req.Do(host, resp)
		if err != nil {
			log.SysLog.Add(err.Error(), log.Error)
			continue
		}
		if !resp.Successful() {
			log.SysLog.Add(resp.Err, log.Error)
			continue
		}
	}

	log.SysLog.Add("update leader to storages ok", log.Info)

}

func (s *Server) syncEsoToStorage(host *etc.Addr, eso *app.EsoData) error {
	req := app.MakeStoSyncFileReq(eso)
	resp := new(app.StoSyncFileResp)

	err := req.Do(host, resp)

	if err != nil {
		return err
	}

	if !resp.Successful() {
		return errors.New(resp.Err)
	}

	return nil

}

//同步管理器
func (s *Server) syncFileWork() {

	//最多可以同时同步30个文件
	maxSyncCount := make(chan int16, 30)

	for {
		select {
		case eso := <-s.syncFileChan:

			maxSyncCount <- 1 //写入同步数据
			go func() {
				if eso.Status == app.Synchronous { //同步中
					log.SysLog.Add(fmt.Sprintf("%s start sync to storage", eso.Hash), log.Info)

					//
					hosts := s.AnyStorageHosts()
					for _, host := range hosts {

						if host.GetAddress() == eso.Hosts {
							continue
						}
						err := s.syncEsoToStorage(host, eso) //发起同步
						//不成功
						if err != nil {
							//同步失败
							log.SysLog.Add(err.Error(), log.Error)
							s.putSyncFileChan(eso) //加入到尾部
							return
						}

					}

					//全部同步成功
					eso.Status = app.Ok //表示完成同步
					s.updateEsoterica(eso)

				}
				<-maxSyncCount
			}()

		}
	}
}

//写入一个正在同步的ESO
func (s *Server) putSyncFileChan(eso *app.EsoData) {
	//只有主机可以进行同步
	if s.getLeader().GetAddress() != s.getLocalHosts().GetAddress() {
		return
	}

	if eso.Status == app.Synchronous {
		s.syncFileChan <- eso
	}

}
