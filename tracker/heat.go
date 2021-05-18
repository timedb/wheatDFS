package tracker

import (
	"github.com/timedb/wheatDFS/app"
	"github.com/timedb/wheatDFS/etc"
)

//检查函数
func examine(addr *etc.Addr) bool {

	if addr == nil {
		return false
	}

	if addr.StateType == etc.StateTracker {
		req := app.MakeTraHeartReq()
		resp := new(app.ResponseBase)

		err := req.Do(addr, resp)
		if err == nil && resp.State == app.ResponseStateOK {
			return true
		}

		return false
	} else {
		//处理storage的情况
		req := app.MakeStoHeartBeat()
		resp := new(app.ResponseBase)
		err := req.Do(addr, resp)

		if err == nil && resp.Successful() {
			return true
		}

		return false
	}
}
