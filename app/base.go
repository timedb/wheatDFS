package app

import (
	"encoding/gob"
	"github.com/timedb/wheatDFS/etc"
	"github.com/timedb/wheatDFS/log"
)

//进行接口注册
func init() {

	gob.Register(RequestBase{})
	gob.Register(ResponseBase{})

}

//基础请求类

type Request interface {
	BindLocalHost(request Request, name string)
}

type RequestBase struct {
	Host       *etc.Addr //自身信息
	This       Request
	methodName string //绑定请求名称
	maxNum     int    //最大请求次数
}

// BindLocalHost 绑定自己的信息
func (r *RequestBase) BindLocalHost(request Request, name string) {
	r.Host = etc.SysLocalHost
	r.This = request
	r.methodName = name
	r.maxNum = etc.SysConf.Pool.MaxReConnNum
}

// Do 发起连接获取信息
func (r *RequestBase) Do(host *etc.Addr, resp Response) error {

	//请求用方法
	reqFunc := func() error {
		//获取连接
		waitConn := SysRpcPool.GetWaitConn(host) //获取排队对象
		conn, err := waitConn.Get()

		if err != nil {
			log.SysLog.Add(err.Error(), log.Error) //请求连接失败
			waitConn.Recycle(err)
			return err
		}

		defer waitConn.Recycle(nil) //返回使用

		req := r.This
		r.This = nil

		//发起连接
		err = conn.Call(r.methodName, req, resp)

		if err != nil {
			return err
		}

		r.This = req //回复

		return nil
	}

	// resend
	for i := 0; i < r.maxNum; i++ {
		err := reqFunc()
		//转发请求
		if juderOk, addr := resp.JudgeTransFor(); juderOk {
			host = addr //转发
			return reqFunc()
		}

		if err == nil {
			return nil
		}

	}

	return OnLinkErr

}

// Response 基础响应类
type Response interface {
	JudgeTransFor() (bool, *etc.Addr)
	AddTransForHost(addr *etc.Addr)
}

type ResponseBase struct {
	State      int
	TransAddr  *etc.Addr //转发请求
	Err        string    //错误
	RemoteAddr *etc.Addr //远程地址
}

// JudgeTransFor 是否转发
func (r *ResponseBase) JudgeTransFor() (bool, *etc.Addr) {
	if r.State == ResponseTransPond && r.TransAddr != nil {
		return true, r.TransAddr
	}

	return false, nil
}

// AddTransForHost 转发
func (r *ResponseBase) AddTransForHost(addr *etc.Addr) {
	r.TransAddr = addr
	r.WriteRespState(ResponseTransPond, nil)
}

// Successful 检查请求是否成功
func (r *ResponseBase) Successful() bool {
	return r.State == ResponseStateOK
}

// WriteRespState 写入状态，无错误可以为空
func (r *ResponseBase) WriteRespState(state int, err error) {

	switch state {
	case ResponseStateOK:
		r.State = ResponseStateOK
	case ResponseStateErr:
		r.State = ResponseStateErr
		r.Err = "The server has experienced an unexpected error"
	case ResponseStateMissingData:
		r.State = ResponseStateMissingData
		r.Err = "missing data"
	case ResponseTransPond:
		r.State = ResponseTransPond
		r.Err = "trans pond"
	default:
		r.State = ResponseStateClientErr
		r.Err = "client has experienced an unexpected error"
	}

	if err != nil {
		r.Err = err.Error()
	}

	//写入远端IP
	r.RemoteAddr = etc.SysLocalHost

}
