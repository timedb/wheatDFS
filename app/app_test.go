package app

import (
	"fmt"
	"github.com/timedb/wheatDFS/etc"
	"github.com/timedb/wheatDFS/log"
	"testing"
	"time"
)

func TestEsoData_Encode(t *testing.T) {
	etc.LoadConf("D:\\goproject\\wheatDFS\\etc\\wheatDFS.ini")
	log.MakeLogging()
	poll := MakeRpcConnectPool() //创建连接池
	go poll.work()

	fmt.Println(poll)
	hosts, _ := etc.MakeAddr("192.168.31.109", "8080", etc.StateDefault)

	for i := 0; i < 2000; i++ {

		go func() {
			wait := poll.GetWaitConn(hosts)
			conn, err := wait.Get()
			if err != nil {
				fmt.Println(err)
				return
			}

			defer wait.Recycle(nil)

			req := new(StoGetConditionReq)
			resp := new(StoGetConditionResp)

			_ = conn.Call(StoAppGetCondition, req, resp)
			fmt.Println(resp)

		}()

	}

	time.Sleep(15 * time.Second)
	fmt.Println(SysRpcPool.NumCap)

	for {
		fmt.Println(<-SysRpcPool.pool["192.168.31.173:8080"])
	}

}
