package app

import (
	"errors"
	"fmt"
	"github.com/timedb/wheatDFS/etc"
	"regexp"
	"strconv"
	"strings"
)

//主接口方法
//tracker接口以Tra开头, storage接口用Sto

// TraRegisterTraReq Req表示请求结构,注册tracker
type TraRegisterTraReq struct {
	RequestBase
}

// TraRegisterTraResp Resp表示响应结构
type TraRegisterTraResp struct {
	ResponseBase
}

// MakeRegisterTraReq 新建一个注册tracker接口
func MakeRegisterTraReq() *TraRegisterTraReq {
	req := new(TraRegisterTraReq)
	req.BindLocalHost(req, TraAppRegister) //一定要调用绑定自己
	return req
}

// TraGetStoAddrReq 负载均衡的获取storage的地址
type TraGetStoAddrReq struct {
	RequestBase
}

type TraGetStoAddrResp struct {
	ResponseBase
	RespHost *etc.Addr
}

func MakeTraGetStoAddr() *TraGetStoAddrReq {
	req := new(TraGetStoAddrReq)

	req.BindLocalHost(req, TraAppGetStoAddr)
	return req
}

func MakeTraHeartReq() *RequestBase {
	req := new(RequestBase)
	req.BindLocalHost(req, TraAppHeartBet)
	return req
}

// TraReportTraHostsReq TraReportTraHosts 上报，并且同步leader
type TraReportTraHostsReq struct {
	RequestBase
	TraHosts []*etc.Addr
	StoHosts []*etc.Addr
}

type TraReportTraHostsResp struct {
	ResponseBase
	TraHosts []*etc.Addr
	StoHosts []*etc.Addr
}

func MakeTraReportTraHosts(tra []*etc.Addr, sto []*etc.Addr) *TraReportTraHostsReq {
	req := new(TraReportTraHostsReq)
	req.TraHosts = tra
	req.StoHosts = sto
	req.BindLocalHost(req, TraAppReportTraHosts)
	return req
}

type TraUpdateTheRoleReq struct {
	RequestBase
}

type TraUpdateTheRoleResp struct {
	ResponseBase
}

func MakeTraUpdateTheRoleReq() *TraUpdateTheRoleReq {
	req := new(TraUpdateTheRoleReq)

	req.BindLocalHost(req, TraAppUpdateTheRale)
	return req
}

// TraAppointedRoleReq 任命角色
type TraAppointedRoleReq struct {
	StateRole int //角色
	RequestBase
}

type TraAppointedRoleResp struct {
	ResponseBase
}

func MakeTraAppointedRoleReq(state int) *TraAppointedRoleReq {
	req := new(TraAppointedRoleReq)
	req.StateRole = state
	req.BindLocalHost(req, TraAppAppointedRole)
	return req
}

// EsoData 秒传结构体
type EsoData struct {
	Status int    //状态
	Hash   string //文件的hash
	Hosts  string //初始保存的地址
	Ext    string //文件后缀
}

// Encode 编码
func (e *EsoData) Encode() []byte {
	ext := strings.Replace(e.Ext, ".", "", 1)
	str := fmt.Sprintf("%d++%s++%s++%s", e.Status, e.Hash, e.Hosts, ext)
	return []byte(str)
}

// Decode 解码
func (e *EsoData) Decode(b []byte) error {
	str := string(b)
	reg := regexp.MustCompile(`^(\d)\+\+(.*?)\+\+(.*?)\+\+(.+)$`)
	buffer := reg.FindStringSubmatch(str)

	if len(buffer) != 5 {
		return errors.New("decode err")
	}

	l, err := strconv.Atoi(buffer[1])
	if err != nil {
		return errors.New("decode err")
	}

	e.Status = l
	e.Hash = buffer[2]
	e.Hosts = buffer[3]
	e.Ext = buffer[4]

	return nil

}

// TraGetEsoDataReq 获取一个秒传
type TraGetEsoDataReq struct {
	RequestBase
	Hash string //hash
}

type TraGetEsoDataResp struct {
	ResponseBase
	Eso *EsoData //EsoData
}

func MakeTraGetEsoDataReq(hash string) *TraGetEsoDataReq {
	req := new(TraGetEsoDataReq)
	req.Hash = hash

	req.BindLocalHost(req, TraAppGetEsoData)
	return req

}

// TraPutDataReq 上传一个秒传接口
type TraPutDataReq struct {
	RequestBase
	Eso *EsoData
}

type TraPutDataResp struct {
	ResponseBase
}

func MakeTraPutDataReq(hash string, status int, hosts *etc.Addr, ext string) *TraPutDataReq {
	req := new(TraPutDataReq)

	//创建eso
	eso := new(EsoData)
	eso.Status = status
	eso.Hash = hash
	eso.Hosts = hosts.GetAddress()
	eso.Ext = ext

	req.Eso = eso

	req.BindLocalHost(req, TraAppPutEsoData)
	return req

}

// TraGetRoleReq 获取tracker的角色
type TraGetRoleReq struct {
	RequestBase
}

type TraGetRoleResp struct {
	ResponseBase
	RoleName string //角色名称
}

func MakeTraGetRoleReq() *TraGetRoleReq {
	r := new(TraGetRoleReq)
	r.BindLocalHost(r, TraAppGetRoleName)

	return r
}
