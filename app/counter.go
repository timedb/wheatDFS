package app

import (
	"encoding/gob"
	"github.com/timedb/wheatDFS/etc"
	"github.com/timedb/wheatDFS/log"
	"os"
	"sync"
	"time"
)

var (
	counterChan chan int
	counterOne  sync.Once
	counter     *Counter
)

const (
	SyncNum = iota
	UploadNum
	DownloadNum
	LoadNum
)

// Counter 计数器, 记入Storage的信息
type Counter struct {
	SyncMum     int
	UploadNum   int
	DownloadNum int
	LoadNum     int
	HistoryLoad []LoadDate //历史负载
	CachePath   string
	HisCount    int //写入负载的次数
}

func MakeCounter() *Counter {

	counterOne.Do(func() {
		counterChan = make(chan int, 2000)

		path := etc.SysConf.StorageConf.CachePath

		counter = new(Counter)

		counter.CachePath = path
		counter.HistoryLoad = make([]LoadDate, 10)
		//根据path地址解析
		counter.decode()

	})

	return counter

}

//解码
func (c *Counter) decode() {
	f, err := os.Open(c.CachePath)
	if err != nil {
		log.SysLog.Add(err.Error(), log.Error)
		return
	}
	defer f.Close()

	dec := gob.NewDecoder(f)

	err = dec.Decode(c)

	if err != nil {
		log.SysLog.Add(err.Error(), log.Error)
		return
	}
}

//编码
func (c *Counter) encode() {
	if c.CachePath == "" {
		return
	}

	f, err := os.Create(c.CachePath)
	if err != nil {
		log.SysLog.Add(err.Error(), log.Error)
		return
	}
	defer f.Close()

	enc := gob.NewEncoder(f)

	err = enc.Encode(c)
	if err != nil {
		log.SysLog.Add(err.Error(), log.Error)
		return
	}
}

//add
func (c *Counter) addCounter(option int) {

	switch option {
	case SyncNum:
		c.SyncMum += 1
	case UploadNum:
		c.UploadNum += 1
	case DownloadNum:
		c.DownloadNum += 1
	case LoadNum:
		c.LoadNum += 1
	}
}

//更新HistoryLoad
func (c *Counter) updateHistoryLoad(time string, data int) {

	date := LoadDate{
		Time: time,
		Load: data,
	}

	c.HistoryLoad[c.HisCount%len(c.HistoryLoad)] = date //修改数据
	c.HisCount += 1

}

// GetCounter 获取信息
func (c *Counter) GetCounter() *Counter {

	return c
}

// Work 定期保存
func (c *Counter) Work() {
	tickerCache := time.NewTicker(time.Minute * 20)
	tickerLoad := time.NewTicker(time.Minute * 20)

	old := c.LoadNum

	for true {
		select {
		//保存
		case <-tickerCache.C:
			c.encode() //保存
		case t := <-tickerLoad.C:
			hour := t.Format("2006-01-02 15")
			src := c.GetCounter()
			c.updateHistoryLoad(hour, src.LoadNum-old)
			old = src.LoadNum
		case op := <-counterChan:
			c.addCounter(op) //添加更新
		}
	}

}

func (c *Counter) Add(option int) {
	counterChan <- option
}

// LoadDate 时间负载数
type LoadDate struct {
	Time string
	Load int
}
