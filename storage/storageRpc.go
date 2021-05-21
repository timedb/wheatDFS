package storage

import (
	"bytes"
	"fmt"
	"github.com/timedb/wheatDFS/app"
	"github.com/timedb/wheatDFS/etc"
	"github.com/timedb/wheatDFS/fileKeyTorch"
	"github.com/timedb/wheatDFS/log"
	"github.com/timedb/wheatDFS/serverTorch"
	"github.com/timedb/wheatDFS/torch/hashTorch"
	"io"
	"net"
	"net/rpc"
)

type StoRpcService struct {
}

func (s *StoRpcService) start() {
	_ = rpc.Register(s)
	tcpAddr, err := net.ResolveTCPAddr("tcp", SysServer.localHost.GetAddress())
	if err != nil {
		log.SysLog.Add(fmt.Sprintf("%v", err), log.Panic)
		return
	}

	//监听端口，开启服务
	listen, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.SysLog.Add(fmt.Sprintf("%v", err), log.Panic)
		return
	}

	concurrence := make(chan int32, 20000) //最大2000个并发处理
	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}

		//服务
		concurrence <- 1 //添加一个负载

		//使用保护模式运行
		go func() {
			rpc.ServeConn(conn)
			//防止负载过多
			defer func() {
				<-concurrence
			}()
		}()

	}
}

// GetServerState 获取cpu，内存、磁盘状态接口
func (s *StoRpcService) GetServerState(req *app.StoGetServerStateReq, resp *app.StoGetServerStateResp) error {

	defer log.ProtectRun() //保护运行

	resp.CPU = serverTorch.GetCpuPercent()
	resp.Disk = serverTorch.GetDiskPercent()
	resp.Mem = serverTorch.GetMemPercent()

	resp.State = app.ResponseStateOK

	return nil
}

// UploadSmallFile 对小文件进行保存接口
func (s *StoRpcService) UploadSmallFile(req *app.StoUploadSmallFileReq, resp *app.StoUploadSmallFileResp) error {

	defer log.ProtectRun() //保护运行
	if req.Content == nil || req.Ext == "" || len(req.Content) == 0 {
		resp.WriteRespState(app.ResponseStateMissingData, nil)
		return nil
	}

	content := bytes.NewBuffer(req.Content)
	hash, err := hashTorch.GetSmallHash(content)
	if hash != req.Hash || err != nil {
		resp.WriteRespState(app.ResponseStateErr, err)
		return nil
	}

	//使用文件令牌进行保存
	fk := fileKeyTorch.MakeFileKeyByHash(hash, req.Ext)

	if fk == nil || fk.Types() != fileKeyTorch.MinFile {
		resp.WriteRespState(app.ResponseStateErr, HashErr)
		return nil
	}

	if _, err := fk.Write(req.Content); err != nil {
		resp.WriteRespState(app.ResponseStateErr, err)
		return nil
	} //保存

	resp.FileKey = fk.GeyToken()
	resp.WriteRespState(app.ResponseStateOK, nil)

	SysServer.AddOesFileHash(fk.GeyToken(), app.Synchronous) // 保存完毕
	SysServer.counter.Add(app.UploadNum)

	return nil

}

// GetSmallFile 获取小文件
func (s *StoRpcService) GetSmallFile(req *app.StoGetSmallFileReq, resp *app.StoGetSmallFileResp) error {

	defer log.ProtectRun() //启动保护模式

	if req.Token == "" {
		resp.WriteRespState(app.ResponseStateMissingData, TokenEmptyErr)
		return nil
	}

	//创建保存令牌
	fk := fileKeyTorch.MakeFileKeyByToken(req.Token)
	if fk == nil {
		resp.WriteRespState(app.ResponseStateErr, InvalidErr)
		return nil
	}

	//只处理小文件
	if fk.Types() != fileKeyTorch.MinFile {
		resp.WriteRespState(app.ResponseStateErr, SmallFileSizeErr)
		return nil
	}

	buf, err := fk.ReadAll()
	if err != nil {
		resp.WriteRespState(app.ResponseStateErr, err)
		return nil
	}

	resp.Content = buf
	resp.WriteRespState(app.ResponseStateOK, nil)
	SysServer.counter.Add(app.DownloadNum)
	return nil

}

// UploadMaxFile  上传大文件
func (s *StoRpcService) UploadMaxFile(req *app.StoUploadMaxFileReq, resp *app.StoUploadMaxFileResp) error {

	defer log.ProtectRun() //保护运行

	if req.Hash == "" || req.Ext == "" {
		resp.WriteRespState(app.ResponseStateMissingData, nil)
		return nil
	}

	//创建传输
	if req.TransferStatus == app.UpStart { //开始传输
		fk := fileKeyTorch.MakeFileKeyByHash(req.Hash, req.Ext)
		//接口只处理大文件
		if fk.Types() != fileKeyTorch.MaxFile {
			resp.WriteRespState(app.ResponseStateErr, MaxFileSizeErr)
			return nil
		}

		SysServer.CacheFile(fk) //缓存
		resp.TransferStatus = app.Link
		resp.WriteRespState(app.ResponseStateOK, nil)
		resp.Offset = fk.GetOffset() //根据偏移传输

		log.SysLog.Add(fk.GeyToken()+"upload start", log.Info)

		SysServer.AddOesFileHash(fk.GeyToken(), app.Transmitting) // 保存

		return nil

	} else if req.TransferStatus == app.UpSustain { //传输

		//去掉无效情况
		if req.Content == nil || len(req.Content) == 0 {
			resp.WriteRespState(app.ResponseStateMissingData, nil)
			return nil
		}

		fk, ok := SysServer.CacheFileRead(req.Hash)
		if !ok { //不存在
			resp.TransferStatus = app.Due //过期或者无效
			resp.WriteRespState(app.ResponseStateErr, FileExpiredErr)
			return nil
		}

		fk.ResetTime()

		//错误的更新
		if fk.GetOffset() != req.Offset {
			resp.WriteRespState(app.ResponseStateErr, OffsetNotProvideErr)
			return nil
		}

		_, err := fk.Write(req.Content)
		if err != nil { //错误以后，删除令牌
			resp.WriteRespState(app.ResponseStateErr, err)
			SysServer.DelCacheFile(fk.GeyToken())
			return nil
		}

		//更新令牌状态
		resp.WriteRespState(app.ResponseStateOK, nil)

		resp.Offset = fk.GetOffset() //更新下一次的偏移

		return nil

	} else if req.TransferStatus == app.UpEnd { //文件传输完毕

		//判断保存的文件的状态
		fk, ok := SysServer.CacheFileRead(req.Hash)
		if !ok {
			resp.WriteRespState(app.ResponseStateErr, InvalidExpiredErr)
			return nil
		}

		SysServer.DelCacheFile(req.Hash) //传输完毕
		log.SysLog.Add("file transfer successful", log.Info)

		resp.WriteRespState(app.ResponseStateOK, nil)
		resp.Token = fk.GeyToken() //返回文件的token

		SysServer.AddOesFileHash(fk.GeyToken(), app.Synchronous)
		SysServer.counter.Add(app.UploadNum)

		return nil
	}

	resp.WriteRespState(app.ResponseStateClientErr, StateErr)
	return nil

}

// GetMaxFile 下载一个大文件
func (s *StoRpcService) GetMaxFile(req *app.StoGetMaxFileReq, resp *app.StoGetMaxFileResp) error {
	defer log.ProtectRun() //启动保护模式

	//必须给定令牌
	if req.Token == "" {
		resp.WriteRespState(app.ResponseStateMissingData, nil)
		return nil
	}

	fk := fileKeyTorch.MakeFileKeyByToken(req.Token)
	if fk == nil || fk.Types() != fileKeyTorch.MaxFile {
		resp.WriteRespState(app.ResponseStateErr, FileKeyErr)
		return nil
	}

	fk.Seek(req.Offset)

	//获取当前偏移的文件
	buf, err := fk.ReadMaxFileCurrent()
	if err != nil {
		if err == io.EOF {
			resp.Offset = -1 //完成
			resp.WriteRespState(app.ResponseStateOK, nil)
			SysServer.counter.Add(app.DownloadNum)

			return nil
		}

		resp.WriteRespState(app.ResponseStateErr, err)
		return nil
	}

	resp.Content = buf //传输中
	resp.Offset = fk.GetOffset()
	resp.WriteRespState(app.ResponseStateOK, nil)
	return nil
}

// HeartBeat 心跳检测
func (s *StoRpcService) HeartBeat(req *app.RequestBase, resp *app.ResponseBase) error {

	resp.WriteRespState(app.ResponseStateOK, nil)
	return nil
}

// UpdateLeader 更新leader状态
func (s *StoRpcService) UpdateLeader(req *app.RequestBase, resp *app.ResponseBase) error {
	defer log.ProtectRun() //保护函数

	if req.Host == nil {
		resp.WriteRespState(app.ResponseStateMissingData, nil)
		return nil
	}

	SysServer.putTraLeader(req.Host)
	resp.WriteRespState(app.ResponseStateOK, nil)

	log.SysLog.Add("update leader to "+req.Host.GetAddress(), log.Info)

	return nil
}

// SyncFile 同步文件接口
func (s *StoRpcService) SyncFile(req *app.StoSyncFileReq, resp *app.StoSyncFileResp) error {
	if req.Eso.Status != app.Synchronous { //不处理
		resp.WriteRespState(app.ResponseStateClientErr, StateErr)
		return nil
	}

	err := SysServer.syncToLocal(req.Eso) //发起同步
	if err != nil {
		resp.WriteRespState(app.ResponseStateErr, err)
		return nil
	}

	//同步完成
	resp.WriteRespState(app.ResponseStateOK, nil)
	SysServer.counter.Add(app.SyncNum)

	return nil

}

// GetLog 获取日志
func (s *StoRpcService) GetLog(req *app.GetLogReq, resp *app.GetLogResp) error {

	defer log.ProtectRun()

	if etc.SysConf.Debug { //没有开启日志
		resp.WriteRespState(app.ResponseStateErr, UnableUseLogErr)
		return nil
	}

	messages := log.SysLog.CheckByTimeAndLevel(req.StartTime, req.EndTime, req.Level)

	resp.LogMsg = messages
	resp.WriteRespState(app.ResponseStateOK, nil)
	return nil

}

// GetCondition 获取服务器负载情况
func (s *StoRpcService) GetCondition(req *app.StoGetConditionReq, resp *app.StoGetConditionResp) error {
	c := SysServer.counter.GetCounter()

	resp.HistoryLoad = c.HistoryLoad
	resp.Sync = c.SyncMum
	resp.Load = c.LoadNum
	resp.Download = c.DownloadNum
	resp.Upload = c.UploadNum

	resp.WriteRespState(app.ResponseStateOK, nil)

	return nil

}
