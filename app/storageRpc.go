package app

// StoGetServerStateReq StoGetServerState 获取服务器状态
type StoGetServerStateReq struct {
	RequestBase
}

type StoGetServerStateResp struct {
	ResponseBase
	CPU  float64 //cpu占用
	Disk float64 //磁盘占用
	Mem  float64 //内存占用
}

func MakeStoGetServerStateReq() *StoGetServerStateReq {
	req := new(StoGetServerStateReq)
	req.BindLocalHost(req, StoAppGetState)
	return req
}

// StoUploadSmallFileReq 上传小文件
type StoUploadSmallFileReq struct {
	RequestBase
	Content []byte //文件内容
	Hash    string
	Ext     string
}

type StoUploadSmallFileResp struct {
	ResponseBase
	FileKey string //文件令牌
}

func MakeStoUploadSmallFileReq(content []byte, hash string, ext string) *StoUploadSmallFileReq {
	req := new(StoUploadSmallFileReq)
	req.BindLocalHost(req, StoAppUploadSmallFile)
	req.Content = content
	req.Ext = ext
	req.Hash = hash
	return req
}

// StoGetSmallFileReq Sto 获取小文件
type StoGetSmallFileReq struct {
	RequestBase
	Token string //令牌
}

type StoGetSmallFileResp struct {
	ResponseBase
	Content []byte //数据
}

func MakeStoGetSmallFileReq(token string) *StoGetSmallFileReq {
	req := new(StoGetSmallFileReq)
	req.Token = token
	req.BindLocalHost(req, StoAppGetSmallFile)
	return req

}

// StoUploadMaxFileReq 上传大文件
type StoUploadMaxFileReq struct {
	RequestBase
	TransferStatus int //传输状态
	Content        []byte
	Hash           string
	Ext            string //后缀
	Offset         int64  //传输进度
}

type StoUploadMaxFileResp struct {
	ResponseBase
	TransferStatus int    //传输状态
	Offset         int64  //传输参数
	Token          string //传输令牌
}

func MakeStoUploadMaxFileReq() *StoUploadMaxFileReq {
	req := new(StoUploadMaxFileReq)
	req.BindLocalHost(req, StoAppUploadMaxFile)

	return req
}

// MakeStoHeartBeat 创建心跳检测包
func MakeStoHeartBeat() *RequestBase {
	r := new(RequestBase)
	r.BindLocalHost(r, StoAppHeartBeat)
	return r
}

// MakeStoUpdateLeader 更新Leader状态
func MakeStoUpdateLeader() *RequestBase {

	r := new(RequestBase)
	r.BindLocalHost(r, StoAppUpdateLeader)

	return r
}

// StoGetMaxFileReq 获取大文件接口
type StoGetMaxFileReq struct {
	RequestBase
	Token  string //获取文件的Token
	Offset int64
}

type StoGetMaxFileResp struct {
	ResponseBase
	Offset  int64
	Content []byte
}

// MakeGetMaxFile 使用token来获取文件
func MakeGetMaxFile(token string) *StoGetMaxFileReq {
	r := new(StoGetMaxFileReq)
	r.BindLocalHost(r, StoAppGetMaxFile)
	r.Token = token

	return r
}

//StoSyncFileReq 进行文件同步接口
type StoSyncFileReq struct {
	RequestBase
	Eso *EsoData
}

type StoSyncFileResp struct {
	ResponseBase
}

func MakeStoSyncFileReq(eso *EsoData) *StoSyncFileReq {
	req := new(StoSyncFileReq)
	req.Eso = eso
	req.BindLocalHost(req, StoAppSyncFile)

	return req

}

// GetLogReq 获取日志接口
type GetLogReq struct {
	RequestBase
	StartTime string //开始时间
	EndTime   string //结束时间
	Level     string //类型
}

type GetLogResp struct {
	ResponseBase
	LogMsg []string
}

// MakeTraGetReq 创建Tracker获取日志接口
func MakeTraGetReq(startTime string, endTime string, level string) *GetLogReq {
	req := new(GetLogReq)
	req.StartTime = startTime
	req.EndTime = endTime
	req.Level = level
	req.BindLocalHost(req, TraAppGetLog)
	return req

}

// MakeStoGetReq 创建Storage获取日志接口
func MakeStoGetReq(startTime string, endTime string, level string) *GetLogReq {
	req := new(GetLogReq)
	req.StartTime = startTime
	req.EndTime = endTime
	req.Level = level
	req.BindLocalHost(req, StoAppGetLog)
	return req

}

// StoGetConditionReq 获取负载接口
type StoGetConditionReq struct {
	RequestBase
}

type StoGetConditionResp struct {
	ResponseBase
	Sync        int        //同步数
	Upload      int        //上传数
	Download    int        //下载数
	Load        int        //负载
	HistoryLoad []LoadDate //历史负载
}

func MakeStoGetConditionReq() *StoGetConditionReq {
	req := new(StoGetConditionReq)
	req.BindLocalHost(req, StoAppGetCondition)
	return req
}
