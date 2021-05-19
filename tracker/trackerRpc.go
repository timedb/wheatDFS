package tracker

import (
	"fmt"
	"github.com/timedb/wheatDFS/app"
	"github.com/timedb/wheatDFS/etc"
	"github.com/timedb/wheatDFS/log"
	"net"
	"net/rpc"
)

type TraRpcService struct {
}

//开始服务
func (s *TraRpcService) start() {
	_ = rpc.Register(s)
	tcpAddr, err := net.ResolveTCPAddr("tcp", sysServer.getLocalHosts().GetAddress())
	if err != nil {
		log.SysLog.Add(fmt.Sprintf("%v", err), log.Panic)
	}

	//监听端口，开启服务
	listen, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.SysLog.Add(fmt.Sprintf("%v", err), log.Panic)
	}

	concurrence := make(chan int16, 20000) //最大20000个并发处理
	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}

		//服务
		concurrence <- 1 //添加一个负载

		//使用保护模型运行
		go func() {
			rpc.ServeConn(conn)
			//防止负载过多
			defer func() {
				<-concurrence
			}()
		}()

	}

}

// RegisterTra 注册一个的服务
func (s *TraRpcService) RegisterTra(req *app.TraRegisterTraReq, resp *app.TraRegisterTraResp) error {
	//拿到tracker地址

	defer log.ProtectRun() //保护运行

	//无效返回
	if req.Host == nil {
		resp.WriteRespState(app.ResponseStateClientErr, HostErr)
		return nil
	}

	if req.Host.StateType == etc.StateTracker {
		sysServer.PutTrackerHost(req.Host) //注册一个tracker
		resp.WriteRespState(app.ResponseStateOK, nil)
		return nil
	} else if req.Host.StateType == etc.StateStorage {
		sysServer.PutStorageHost(req.Host)
		resp.WriteRespState(app.ResponseStateOK, nil)
		return nil
	}

	resp.WriteRespState(app.ResponseStateClientErr, RegisterErr)

	return nil

}

// GetStoAddrTra 使用负载均衡获取storage
func (s *TraRpcService) GetStoAddrTra(req *app.TraGetStoAddrReq, resp *app.TraGetStoAddrResp) error {

	defer log.ProtectRun() //保护运行
	resp.RespHost = sysServer.GetStorageHost()
	resp.WriteRespState(app.ResponseStateOK, nil)
	return nil
}

// HeartTra 心跳检查包
func (s *TraRpcService) HeartTra(req *app.RequestBase, resp *app.ResponseBase) error {
	defer log.ProtectRun() //保护运行
	resp.WriteRespState(app.ResponseStateOK, nil)
	return nil
}

// SyncLeader 上报当前信息并且对leader进行更新同步
func (s *TraRpcService) SyncLeader(req *app.TraReportTraHostsReq, resp *app.TraReportTraHostsResp) error {

	defer log.ProtectRun() //保护运行

	// trans for leader
	if sysServer.getState() != stateLeader {
		resp.AddTransForHost(sysServer.getLeader())
		return nil
	}

	//给予客户端或者传输为nil时
	if req.Host.StateType == etc.StateDefault || (req.StoHosts == nil && req.TraHosts == nil) ||
		(len(req.TraHosts) == 0 && len(req.StoHosts) == 0) {

		resp.TraHosts = append(sysServer.AnyTrackerHosts(), sysServer.getLeader()) //获取全部的信息写入

		resp.StoHosts = sysServer.AnyStorageHosts()
		resp.WriteRespState(app.ResponseStateOK, nil)
		return nil
	}

	//写入不存在部分
	sysServer.PutTrackerHost(req.TraHosts...) //写入更新体
	sysServer.PutTrackerHost(req.TraHosts...)

	resp.TraHosts = sysServer.AnyTrackerHosts() //获取全部的信息写入
	resp.StoHosts = sysServer.AnyStorageHosts()

	resp.WriteRespState(app.ResponseStateOK, nil)
	log.SysLog.Add(fmt.Sprintf("%s perform sync Tracker and Storage", req.Host.GetAddress()), log.Info)

	return nil

}

// UpdateTheRale 投票(选举)
func (s *TraRpcService) UpdateTheRale(req *app.TraUpdateTheRoleReq, resp *app.TraUpdateTheRoleResp) error {
	defer log.ProtectRun() //保护运行

	if !examine(sysServer.getLeader()) {
		resp.WriteRespState(app.ResponseStateOK, nil)
		return nil
	}

	resp.WriteRespState(app.ResponseStateErr, nil)

	return nil
}

// AppointedRole 任命
func (s *TraRpcService) AppointedRole(req *app.TraAppointedRoleReq, resp *app.TraAppointedRoleResp) error {
	defer log.ProtectRun() //保护运行

	//确定请求是否正确
	if !(req.StateRole == stateFollower || req.StateRole == stateCandidate) {
		resp.WriteRespState(app.ResponseStateErr, nil)
		return nil
	}

	sysServer.updateVole(req.StateRole, req.Host)
	resp.WriteRespState(app.ResponseStateOK, nil)

	if req.StateRole == stateCandidate {
		log.SysLog.Add("a appointed as candidate", log.Info)
	} else if req.StateRole == stateFollower {
		log.SysLog.Add("a appointed as follower", log.Info)
	} else {
		log.SysLog.Add("a appointed err", log.Info)
	}

	return nil

}

// GetEsoData 查询一个秒传是否存在
func (s *TraRpcService) GetEsoData(req *app.TraGetEsoDataReq, resp *app.TraGetEsoDataResp) error {
	defer log.ProtectRun()

	//必须给hash
	if req.Hash == "" {
		resp.WriteRespState(app.ResponseStateMissingData, nil)
		return nil
	}

	eso, err := sysServer.viewEsoterica(req.Hash)
	if err != nil {

		data, ok := sysServer.viewLeaderEso(req.Hash)
		if !ok {
			resp.WriteRespState(app.ResponseStateErr, err)
			return nil
		}

		sysServer.updateEsoterica(data)
		eso = data // 更新leader来的指针
	}

	resp.Eso = eso

	resp.WriteRespState(app.ResponseStateOK, nil)
	return nil

}

// PutEsoData 上传一个秒传
func (s *TraRpcService) PutEsoData(req *app.TraPutDataReq, resp *app.TraPutDataResp) error {
	defer log.ProtectRun()

	if sysServer.getState() != stateLeader {
		resp.AddTransForHost(sysServer.getLeader())
		return nil
	}

	//处理无效问题
	if req.Eso == nil || req.Eso.Hash == "" || req.Eso.Hosts == "" {
		resp.WriteRespState(app.ResponseStateMissingData, nil)
		return nil
	}

	//写入，错误无效
	err := sysServer.updateEsoterica(req.Eso)
	if err != nil {
		resp.WriteRespState(app.ResponseStateErr, err)
		return nil
	}

	resp.WriteRespState(app.ResponseStateOK, nil)
	return nil

}

// GetLog 获取日志
func (s *TraRpcService) GetLog(req *app.GetLogReq, resp *app.GetLogResp) error {

	defer log.ProtectRun()

	if etc.SysConf.Debug { //没有开启日志
		resp.WriteRespState(app.ResponseStateErr, DebugErr)
		return nil
	}

	messages := log.SysLog.CheckByTimeAndLevel(req.StartTime, req.EndTime, req.Level)

	resp.LogMsg = messages
	resp.WriteRespState(app.ResponseStateOK, nil)
	return nil

}

// GetRoleName 获取角色状态
func (s *TraRpcService) GetRoleName(req *app.TraGetRoleReq, resp *app.TraGetRoleResp) error {
	defer log.ProtectRun()
	state := sysServer.getState()

	switch state {
	case stateLeader:
		resp.RoleName = "Leader"
	case stateCandidate:
		resp.RoleName = "Candidate"
	case stateFollower:
		resp.RoleName = "Follower"
	default:
		resp.RoleName = "None"
	}

	resp.WriteRespState(app.ResponseStateOK, nil)
	return nil

}
