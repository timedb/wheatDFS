package storage

import (
	"errors"
	"fmt"
	"github.com/timedb/wheatDFS/app"
	"github.com/timedb/wheatDFS/etc"
	"github.com/timedb/wheatDFS/fileKeyTorch"
	logner "github.com/timedb/wheatDFS/log"
	serverTorch2 "github.com/timedb/wheatDFS/serverTorch"
	"github.com/timedb/wheatDFS/torch/clientTorch"
	"log"
	"sync"
	"time"
)

var ones sync.Once

// Server storage服务
type Server struct {
	groupPath     string         //写入组
	localHost     *etc.Addr      //本地地址
	service       *StoRpcService //rpc通信服务
	traLeader     *etc.Addr      //tracker leader的地址
	traLeaderLock sync.RWMutex   //traLeader的锁

	cacheFile *sync.Map //缓存文件

	counter *app.Counter //计数器
}

// MakeServer 创建server对象
func MakeServer() *Server {
	ones.Do(func() {
		ser := new(Server)
		ip, err := serverTorch2.GetIPv4s()
		if err != nil {
			//获取ip失败
			log.Fatalln(err)
		}

		var (
			localPort string
		)

		//获取端口
		if etc.SysConf.StorageConf.Port != "" {
			localPort = etc.SysConf.StorageConf.Port
		} else {
			//无法获取可用端口
			if localPort, err = serverTorch2.GetAvailablePort(); err != nil {
				log.Fatalln(err)
			}
		}

		ser.groupPath = etc.SysConf.StorageConf.GroupPath //写入path
		ser.traLeader = etc.SysConf.TrackerConf.CandidateHost

		ser.localHost, _ = etc.MakeAddr(ip, localPort, etc.StateStorage) //rpc接口，小文件传输
		ser.service = new(StoRpcService)

		//创建缓存
		ser.cacheFile = new(sync.Map)

		ser.counter = app.MakeCounter()

		etc.SysLocalHost = ser.localHost

		SysServer = ser
	})

	if etc.SysConf.Debug == false {
		serverTorch2.CreteMkdir(etc.SysConf.StorageConf.GroupPath) //自动检查并且创建group文件组
	}

	etc.SysLocalHost = SysServer.localHost //绑定本地地址

	return SysServer
}

// RegisterToTraLeader 初始化到leader
func (s *Server) RegisterToTraLeader() {
	req := app.MakeRegisterTraReq()
	resp := new(app.TraRegisterTraResp)
	err := req.Do(s.traLeader, resp)
	if err != nil {
		logner.SysLog.Add(err.Error(), logner.Panic)
	}
	if resp.State != app.ResponseStateOK {
		logner.SysLog.Add("not register to tracker leader", logner.Panic)

		log.Fatalln("regiser err")

	}

}

// Start 启动服务
func (s *Server) Start() {
	s.RegisterToTraLeader() //注册到leader中

	fmt.Println("bind host", s.localHost.GetAddress())
	go s.work() //启动自动维护

	go s.counter.Work() //启动计数器维护
	s.service.start()   //启动服务
}

// CacheFile 缓存一个文件令牌的状态
func (s *Server) CacheFile(file *fileKeyTorch.FileKey) {
	//保证可以使用
	if file == nil {
		return
	}

	s.cacheFile.Store(file.Hash, file)

}

// CacheFileRead 读取一个缓存令牌
func (s *Server) CacheFileRead(key string) (*fileKeyTorch.FileKey, bool) {
	val, ok := s.cacheFile.Load(key)

	var fk = val.(*fileKeyTorch.FileKey)

	return fk, ok

}

// DelCacheFile 删除一个Key
func (s *Server) DelCacheFile(key string) {
	s.cacheFile.Delete(key)
}

//storage的工作流函数
func (s *Server) work() {
	ticker := time.NewTicker(time.Second * 60)

	//删除过期的对象
	deleteTheOverdue := func() {
		defer logner.ProtectRun() //保护
		s.cacheFile.Range(func(key, value interface{}) bool {
			fk := value.(*fileKeyTorch.FileKey)
			if !fk.PastDue() { //已经过期
				s.DelCacheFile(key.(string))
			}

			return true

		})

	}

	for true {
		select {
		case <-ticker.C:
			deleteTheOverdue()
		}
	}
}

// AddOesFileHash 添加一个秒传的hash到leader上
func (s *Server) AddOesFileHash(token string, status int) {

	req := app.MakeTraPutDataReq(token, status, s.localHost)
	resp := new(app.TraGetEsoDataResp)
	err := req.Do(s.traLeader, resp)
	if err != nil {
		logner.SysLog.Add(err.Error(), logner.Error)
		return
	}

	if !resp.Successful() {
		logner.SysLog.Add(resp.Err, logner.Error)
		return
	}

}

//获取traLeader
func (s *Server) getTraLeader() *etc.Addr {
	s.traLeaderLock.RLock()
	defer s.traLeaderLock.RUnlock()

	return s.traLeader
}

//更新traLeader
func (s *Server) putTraLeader(host *etc.Addr) {
	s.traLeaderLock.Lock()
	defer s.traLeaderLock.Unlock()

	s.traLeader = host
}

//同步一个Eso文件到本地
func (s *Server) syncToLocal(eso *app.EsoData) error {

	//检查数据是否全部存在
	if eso.Token == "" || eso.Hosts == "" {
		return MissDataErr
	}

	host, err := clientTorch.PathHostsToAddr(eso.Hosts)
	if err != nil {
		return err
	}

	fk := fileKeyTorch.MakeFileKeyByToken(eso.Token)
	defer fk.Close() //最后关闭全部文件指针

	if fk == nil {
		return CreatFkErr
	}

	token := fk.GeyToken() //获取token

	if fk.Types() == fileKeyTorch.MinFile { //处理小文件

		req := app.MakeStoGetSmallFileReq(token)
		resp := new(app.StoGetSmallFileResp)
		err = req.Do(host, resp)
		//判断请求
		if err != nil {
			return err
		}
		if !resp.Successful() {
			err = errors.New(resp.Err)
			return err
		}

		//写入数据
		fk.Write(resp.Content)
		logner.SysLog.Add(token+"sync ok", logner.Info)
		return nil

	} else { //处理大文件
		req := app.MakeGetMaxFile(token)
		resp := new(app.StoGetMaxFileResp)

		for {
			err = req.Do(host, resp)
			//处理连接错误
			if err != nil {
				return err
			}

			//处理请求错误
			if !resp.Successful() {
				err = errors.New(resp.Err)
				return err
			}

			//结束
			if resp.Offset == -1 {
				return nil
			}

			fk.Write(resp.Content)
			req.Offset = resp.Offset

		}

	}

}
