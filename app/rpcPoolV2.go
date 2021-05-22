package app

import (
	"github.com/timedb/wheatDFS/etc"
	"net/rpc"
	"runtime"
	"sync"
	"time"
)

//实现第二个版本的连接池

var (
	oneRpcPool sync.Once
	SysRpcPool *RpcConnectPool
)

type WaitConn struct {
	Conn    chan *rpc.Client //等待
	Hosts   *etc.Addr        //请求连接
	Err     error            //错误
	client  *rpc.Client
	timeOut time.Duration //超时时间
}

// Get 获取连接
func (w *WaitConn) Get() (*rpc.Client, error) {
	timer := time.NewTimer(w.timeOut)

	select {
	case conn := <-w.Conn:

		w.client = conn
		//获取到连接
		return conn, w.Err

	case <-timer.C:
		close(w.Conn) //关闭管道
		return nil, TimeOutErr
	}
}

// Recycle 关闭并且回收WaitConn
func (w *WaitConn) Recycle(err error) {
	SysRpcPool.reCycleConn(w, err) //回收
}

type RpcConnectPool struct {
	InitCap int            //初始连接数
	MaxCap  int            //最大连接数
	NumCap  map[string]int //当前连接数

	lockNumCap sync.RWMutex //读写🔒

	pool map[string]chan *rpc.Client //连接池

	lockMakePool sync.Mutex //创建连接池锁

	waitChan chan *WaitConn //等待
	TimeOut  time.Duration  //超时时间
}

// MakeRpcConnectPool 创建连接池
func MakeRpcConnectPool() *RpcConnectPool {
	oneRpcPool.Do(func() {
		poll := new(RpcConnectPool)
		poll.MaxCap = etc.SysConf.Pool.MaxConnNum
		poll.InitCap = etc.SysConf.Pool.InitConnNum
		poll.NumCap = make(map[string]int) //初始化连接池数量
		poll.TimeOut = etc.SysConf.Pool.TimeOut * time.Second
		poll.waitChan = make(chan *WaitConn, 2000)    //最大排队人数
		poll.pool = make(map[string]chan *rpc.Client) //创建pool数

		SysRpcPool = poll

		go SysRpcPool.work() //创建工作

	})

	return SysRpcPool

}

//获取Num数据
func (r *RpcConnectPool) getNum(host *etc.Addr) int {

	r.lockNumCap.RLock()
	defer r.lockNumCap.RUnlock()

	var num int
	num = r.NumCap[host.GetAddress()]
	return num
}

//增加num的数据
func (r *RpcConnectPool) addNum(host *etc.Addr) bool {
	r.lockNumCap.Lock()
	defer r.lockNumCap.Unlock()

	num, ok := r.NumCap[host.GetAddress()]
	if ok {

		if num >= r.MaxCap {
			return false
		}

		r.NumCap[host.GetAddress()] = num + 1 //添加一个记入
		return true
	}

	r.NumCap[host.GetAddress()] = 1 //创建
	return true

}

//减少num数据
func (r *RpcConnectPool) reduceNum(host *etc.Addr) {
	r.lockNumCap.Lock()
	defer r.lockNumCap.Unlock()

	num, ok := r.NumCap[host.GetAddress()]
	if ok {
		if num > 0 {
			r.NumCap[host.GetAddress()] = num - 1 //添加一个记入
		}

	}

}

//创建初始化pool
func (r *RpcConnectPool) makePollConn(host *etc.Addr) {
	r.lockMakePool.Lock()
	defer r.lockMakePool.Unlock()

	if _, ok := r.pool[host.GetAddress()]; ok {
		return
	}
	r.pool[host.GetAddress()] = make(chan *rpc.Client, r.MaxCap)

}

// GetWaitConn 排队等待连接
func (r *RpcConnectPool) GetWaitConn(host *etc.Addr) *WaitConn {
	waitConn := new(WaitConn)
	waitConn.Conn = make(chan *rpc.Client)
	waitConn.Hosts = host
	waitConn.timeOut = r.TimeOut
	r.waitChan <- waitConn //排队等待连接

	return waitConn
}

//获取或者创建一个连接
func (r *RpcConnectPool) getConn(wait *WaitConn) {

	var connRpc *rpc.Client
	var err error

	//防止关闭管道报错
	defer func() {
		err := recover()
		switch err.(type) {
		case runtime.Error: //运行错误，回调连接
			wait.client = connRpc //回收
			r.reCycleConn(wait, nil)
		}
	}()

	//尝试获取连接
	connects, ok := r.pool[wait.Hosts.GetAddress()] //获取连接队列
	timer := time.NewTimer(time.Millisecond * 100)

	//无连接管道的时候, 直接创建连接
	if !ok {
		r.addNum(wait.Hosts)

		connRpc, err = rpc.Dial("tcp", wait.Hosts.GetAddress())
		wait.Conn <- connRpc
		wait.Err = err
		//创建管道空间
		r.makePollConn(wait.Hosts)
		return
	}

	if ok {
		select {
		//尝试获取连接
		case connRpc = <-connects:
			wait.Conn <- connRpc //写入连接
			wait.Err = nil
			return

		//尝试创建连接
		case <-timer.C:
			flag := r.addNum(wait.Hosts)
			if flag {
				//创建连接返回
				connRpc, err = rpc.Dial("tcp", wait.Hosts.GetAddress())
				wait.Conn <- connRpc
				wait.Err = err
				return
			}

		}
	}

	//都不成功，等待获取连接
	wait.Err = nil
	connRpc = <-r.pool[wait.Hosts.GetAddress()] //等待连接写入
	wait.Conn <- connRpc

}

//回收一个连接
func (r *RpcConnectPool) reCycleConn(wait *WaitConn, err error) {
	if wait.client == nil || wait.Hosts == nil {
		return
	}

	if err != nil {
		//删除连接
		wait.client.Close() //关闭连接
		r.reduceNum(wait.Hosts)

	}

	r.pool[wait.Hosts.GetAddress()] <- wait.client //回收
}

//创建工作进程，分配等待连接
func (r *RpcConnectPool) work() {
	for true {
		select {
		//获取排队情况
		case wait := <-r.waitChan:
			go r.getConn(wait) //获取连接
		}
	}
}
