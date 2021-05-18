package main

import (
	"flag"
	"fmt"
	"github.com/timedb/wheatDFS/app"
	"github.com/timedb/wheatDFS/etc"
	"github.com/timedb/wheatDFS/log"
	"github.com/timedb/wheatDFS/serverTorch"
	"github.com/timedb/wheatDFS/storage"
	"github.com/timedb/wheatDFS/tracker"
	"os"
)

var (
	nc      string //新建一个conf.init
	conf    string //声明conf地址
	serType string
)

//绑定命令行参数
func init() {
	flag.StringVar(&nc, "nc", "", "Initializes a default configuration file."+
		"The output address needs to be specified")

	flag.StringVar(&conf, "conf", "./wheatDFS.ini", "Specifies the configuration file to start the service")
	flag.StringVar(&serType, "type", "", "Use type to specify a service type. The value can only be tracker or storage")
	flag.Parse()
}

var confInit = `
version = "2.1.1"
debug = false

[tracker]
# this parameter used in fixed Synchronization mechanism path value
persistencePath = "./sync.db"
# this parameter is conforming leader's ip
ip = "%s"
# this parameter is conforming leader's port
port = "5590"
# the database of the fast-upload path
esotericaPath = "./wheatDFS.eso"
# the maxinum of the syncgronization datas
syncMaxCount = 500

[storage]
# this parameter is making the storage path in your server
groupPath = "D:/goproject/wheatDFS/storage/test"
# the maximum number of storage accesses
maxCount = 10
# this is base bit unit
unitSize = 512.0
# you will wait the uploading and downloading of big files with this parameter's value
maxCacheTime = 10
# the storage port
port = "5591"
# storage cache address
cachePath = "./cache.sto"


[log]
# the path of log database
logPath = "./log.db"

[pool]
# Maximum number of connections
maxConnNum = 30
# Initial number of connections
initConnNum = 5
# time-out second
timeOut = 10

[client]
# the client prot
port = "5592"
cachePath = "cache.cli"

`

func main() {
	//创建配置文件
	if nc != "" {
		f, err := os.Create(nc)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		ip, _ := serverTorch.GetIPv4s()
		f.WriteString(fmt.Sprintf(confInit, ip))
		return
	}

	if serType == "tracker" {
		//读取公共部分
		etc.LoadConf(conf)
		log.MakeLogging()        // 启动日志器
		app.MakeRpcConnectPool() //创建连接池
		server := tracker.MakeServer()
		server.StartServer()
	} else if serType == "storage" {
		//读取公共部分
		etc.LoadConf(conf)
		log.MakeLogging()        // 启动日志器
		app.MakeRpcConnectPool() //创建连接池
		server := storage.MakeServer()
		server.Start()
	} else if serType == "" {
		fmt.Println("Type -h to see help")
		return
	} else {
		fmt.Println("Type can only be storage or tracker")
		return
	}

}
