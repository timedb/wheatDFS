package app

const (
	ResponseStateOK          = 200 //成功
	ResponseStateMissingData = 403 //缺少数据
	ResponseStateClientErr   = 400 //客户端错误
	ResponseStateErr         = 500 //处理错误
	ResponseTransPond        = 300 //转发请求
)

// PoolSize 连接池
const PoolSize = 5

// UpStart 大文件上传状态
const (
	UpStart   = 1 //传输开始
	UpSustain = 2 //传输中段
	UpEnd     = 3 //传输结束
)

// Due UpStart断点续传
const (
	Due  = 1 //过期
	Link = 2 //继续传输
)

//秒传机制
const (
	Transmitting = 1 //传输中
	Synchronous  = 2 //同步中
	Ok           = 3 //全部完成
)

// TrackerApp Tracker连接接口
const TrackerApp = "TraRpcService."
const (
	TraAppRegister       = TrackerApp + "RegisterTra"   //注册服务接口
	TraAppGetStoAddr     = TrackerApp + "GetStoAddrTra" // 返回一个storage的负载地址
	TraAppHeartBet       = TrackerApp + "HeartTra"      // 心跳
	TraAppReportTraHosts = TrackerApp + "SyncLeader"    // SyncLeader同步tracker
	TraAppUpdateTheRale  = TrackerApp + "UpdateTheRale" //确定是否升级
	TraAppAppointedRole  = TrackerApp + "AppointedRole" //任命
	TraAppGetEsoData     = TrackerApp + "GetEsoData"    //添加一个秒传
	TraAppPutEsoData     = TrackerApp + "PutEsoData"    //添加一个秒传
	TraAppGetLog         = TrackerApp + "GetLog"        //获取日志
	TraAppGetRoleName    = TrackerApp + "GetRoleName"   //获取状态
)

// StorageApp Storage接口
const StorageApp = "StoRpcService."
const (
	StoAppGetState        = StorageApp + "GetServerState"
	StoAppUploadSmallFile = StorageApp + "UploadSmallFile"
	StoAppGetSmallFile    = StorageApp + "GetSmallFile"
	StoAppUploadMaxFile   = StorageApp + "UploadMaxFile"
	StoAppHeartBeat       = StorageApp + "HeartBeat"
	StoAppUpdateLeader    = StorageApp + "UpdateLeader"
	StoAppGetMaxFile      = StorageApp + "GetMaxFile"   //大文件上传
	StoAppSyncFile        = StorageApp + "SyncFile"     //同步文件接口
	StoAppGetLog          = StorageApp + "GetLog"       //获取日志
	StoAppGetCondition    = StorageApp + "GetCondition" //获取信息
)
