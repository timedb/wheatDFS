package tracker

import (
	"fmt"
	"github.com/timedb/wheatDFS/app"
	"github.com/timedb/wheatDFS/etc"
	"github.com/timedb/wheatDFS/log"
	"github.com/timedb/wheatDFS/torch/hashTorch"
	"io/ioutil"
	"os"
	"testing"
)

func TestTraServer_StartService(t *testing.T) {
	etc.LoadConf("D:/gotest/wheatDFS.ini")
	log.MakeLogging()
	app.MakeRpcConnectPool()

	tracker := MakeServer()
	tracker.StartServer()
}

func TestServer_GetStorageHost(t *testing.T) {
	etc.LoadConf("D:\\goproject\\github/timedb/wheatDFS\\etc\\github/timedb/wheatDFS.ini")
	log.MakeLogging()

	addr, _ := etc.MakeAddr("127.0.0.1", "8080", etc.StateDefault)

	f, _ := os.Open("D:\\goproject\\github/timedb/wheatDFS\\开发日志.md")
	hash, _ := hashTorch.Integrate(f)
	buf, _ := ioutil.ReadAll(f)
	req1 := app.MakeStoUploadSmallFileReq(buf, hash, ".md")
	resq1 := new(app.StoUploadSmallFileResp)
	req1.Do(addr, resq1)
	key := resq1.FileKey

	req2 := app.MakeStoGetSmallFileReq(key)
	resp2 := new(app.StoGetSmallFileResp)
	req2.Do(addr, resp2)
	fmt.Println(resp2.Content)

}

func TestServer_GetStorageHost2(t *testing.T) {

	addr, _ := etc.MakeAddr("192.168.31.110", "41939", etc.StateTracker)
	req := app.MakeTraGetStoAddr()

	k := make(chan int16, 2000)
	for i := 0; i < 2000; i++ {
		go func() {

			resp := new(app.TraGetStoAddrResp)
			req.Do(addr, resp)

			fmt.Println(resp.RespHost.GetAddress())
			k <- 1

		}()
	}
	for i := 0; i < 2000; i++ {
		<-k
	}

}

func TestTraRpcService_GetEsoData(t *testing.T) {
	host, _ := etc.MakeAddr("192.168.31.173", "5590", etc.StateDefault)
	req := app.MakeTraGetEsoDataReq("aadde71bf8a43ddcc220b59e730f5adeb4649714a42776a13c5c0826a06451c91")

	resp := new(app.TraGetEsoDataResp)

	req.Do(host, resp)

	fmt.Println(resp.Eso)
}
