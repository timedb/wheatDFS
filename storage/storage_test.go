package storage

import (
	"fmt"
	"github.com/timedb/wheatDFS/app"
	"github.com/timedb/wheatDFS/etc"
	logner "github.com/timedb/wheatDFS/log"
	"github.com/timedb/wheatDFS/torch/hashTorch"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestMakeServer(t *testing.T) {
	etc.LoadConf("D:\\goproject\\github.com/timedb/wheatDFS\\etc\\github.com/timedb/wheatDFS.ini")
	server := MakeServer()
	fmt.Println(server.traLeader)

}

func TestServer_Start(t *testing.T) {
	etc.LoadConf("D:\\goproject\\github.com/timedb/wheatDFS\\etc\\github.com/timedb/wheatDFS.ini")
	logner.MakeLogging()
	server := MakeServer()
	server.Start()
}

func TestStoRpcService_GetServerState(t *testing.T) {
	host, _ := etc.MakeAddr("100.84.96.129", "58053", etc.StateStorage)
	resp := new(app.StoGetServerStateResp)
	req := app.MakeStoGetServerStateReq()
	req.Do(host, resp)
	fmt.Println(resp.CPU, resp.Disk, resp.Mem, resp.Successful())

}

func TestStoRpcService_UploadSmallFile(t *testing.T) {
	file, _ := os.Open("D:\\goproject\\github.com/timedb/wheatDFS\\etc\\github.com/timedb/wheatDFS.ini")
	hash, _ := hashTorch.GetSmallHash(file)
	file.Seek(io.SeekStart, 0)
	buf, _ := ioutil.ReadAll(file)
	req := app.MakeStoUploadSmallFileReq(buf, hash, ".ini")
	resp := new(app.StoUploadSmallFileResp)
	host, _ := etc.MakeAddr("10.15.98.163", "8080", etc.StateDefault)

	req.Do(host, resp)

	fmt.Println(resp)

}

func TestStoRpcService_GetSmallFile(t *testing.T) {
	token := "group/93/51/ad822e4d4762dac91f65864c35c0cf3c0.ini"
	req := app.MakeStoGetSmallFileReq(token)
	host, _ := etc.MakeAddr("10.15.98.62", "55942", etc.StateStorage)
	resp := new(app.StoGetSmallFileResp)
	req.Do(host, resp)
	path := "D:\\goproject\\gotest\\ad8.ini"
	f, _ := os.Create(path)
	defer f.Close()

	if resp.Successful() {
		f.Write(resp.Content)
		fmt.Println("保存成功")
	}

}

//大文件传输部分代码
func TestStoRpcService_GetMaxFile(t *testing.T) {
	etc.LoadConf("D:\\goproject\\github.com/timedb/wheatDFS\\etc\\github.com/timedb/wheatDFS.ini")
	host, _ := etc.MakeAddr("127.0.0.1", "8080", etc.StateDefault)
	req := app.MakeStoUploadMaxFileReq()
	resp := new(app.StoUploadMaxFileResp)

	path := "F:\\视频\\MyTv\\fodfs\\使用.mp4"
	f, _ := os.Open(path)
	defer f.Close()

	hash, _ := hashTorch.Integrate(f)

	//开始请求
	req.Hash = hash
	req.TransferStatus = app.UpStart
	req.Ext = ".mp4"
	err := req.Do(host, resp)
	if resp.State != app.ResponseStateOK || err != nil {
		fmt.Println(resp.Err, err)
		return
	}

	req.TransferStatus = app.UpSustain //传输数据

	//传输数据
	buf := make([]byte, int(etc.SysConf.StorageConf.UnitSize)*1024)
	for true {
		n, err := f.Read(buf)

		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
		}
		req.Content = buf[:n]

		err = req.Do(host, resp)
		if resp.State != app.ResponseStateOK || err != nil {
			fmt.Println(resp.Err, err)
			return
		}

	}

	//传输结束
	req.Content = nil
	req.TransferStatus = app.UpEnd
	err = req.Do(host, resp)
	if resp.State != app.ResponseStateOK || err != nil {
		fmt.Println(resp.Err, err)
		return
	}

	fmt.Println("保存成功")

}

func TestStoRpcService_UpSmallFile2(t *testing.T) {
	etc.LoadConf("D:\\goproject\\github.com/timedb/wheatDFS\\etc\\github.com/timedb/wheatDFS.ini")
	logner.MakeLogging()
	//UploadSmallFile
	f, _ := os.Open("D:\\goproject\\github.com/timedb/wheatDFS\\README.md")
	hash, _ := hashTorch.Integrate(f)
	fmt.Println(hash)
	buff, _ := ioutil.ReadAll(f)
	req := app.MakeStoUploadSmallFileReq(buff, hash, ".txt")
	resp := new(app.StoUploadSmallFileResp)
	addr, _ := etc.MakeAddr("192.168.31.173", "8080", etc.StateDefault)
	req.Do(addr, resp)
	if resp.Successful() {
		fmt.Println(resp.FileKey)
	}
}

func TestStoRpcService_GetMaxFile2(t *testing.T) {
	etc.LoadConf("/home/lgq/code/go/goDFS/etc/github.com/timedb/wheatDFS.ini")
	logner.MakeLogging()

	req := app.MakeStoGetSmallFileReq("group/63/26/ff939323270d838e8ebab20f5de1a8c50.txt")
	resp := new(app.StoGetSmallFileResp)
	addr, _ := etc.MakeAddr("192.168.31.110", "8080", etc.StateDefault)
	req.Do(addr, resp)

	fmt.Println(resp.Content)

	f, _ := os.Create("/home/lgq/cmd.txt")
	defer f.Close()
	f.Write(resp.Content)

}

func TestStoRpcService_UploadMaxFile(t *testing.T) {
	etc.LoadConf("/home/lgq/code/go/goDFS/etc/github.com/timedb/wheatDFS.ini")
	logner.MakeLogging()

	f, _ := os.Open("/home/lgq/文档/html.rar")
	defer f.Close()

	hash, _ := hashTorch.Integrate(f)
	fmt.Println(hash)

	buff := make([]byte, 1024<<5)
	addr, _ := etc.MakeAddr("192.168.31.173", "8080", etc.StateDefault)
	resp := new(app.StoUploadSmallFileResp)

	var err error
	var n int

	req := app.MakeStoUploadMaxFileReq()
	req.Hash = hash
	req.Ext = ".rar"

	for err == nil {
		n, err = f.Read(buff)
		req.Content = buff[:n]
		req.Do(addr, resp)
		fmt.Println(buff[:n])
	}
}

func TestServer_CacheFile(t *testing.T) {
	etc.LoadConf("D:\\goproject\\github.com/timedb/wheatDFS\\etc\\github.com/timedb/wheatDFS.ini")
	logner.MakeLogging()

	counter := app.MakeCounter()
	go counter.Work()

	for i := 0; i < 10000; i++ {
		go func() {
			counter.Add(app.LoadNum)
			fmt.Println(counter.LoadNum)
		}()
	}

	time.Sleep(2 * time.Second)
	fmt.Println(counter.LoadNum)

}

func TestStoRpcService_GetCondition(t *testing.T) {
	addr, _ := etc.MakeAddr("192.168.31.109", "8080", etc.StateDefault)
	req := app.MakeStoGetConditionReq()
	resp := new(app.StoGetConditionResp)
	req.Do(addr, resp)
	fmt.Println(resp.Upload)
}
