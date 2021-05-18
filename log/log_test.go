package log

import (
	"fmt"
	"testing"
	"wheatDFS/etc"
)

func TestConn(t *testing.T) {
	etc.LoadConf("/home/lgq/code/go/goDFS/etc/wheatDFS.ini")
	MakeLogging()
	SysLog.Add("111", Info)
}

func TestLogging_CheckByTimeAndLevel(t *testing.T) {
	etc.LoadConf("/home/lgq/code/go/goDFS/etc/wheatDFS.ini")
	MakeLogging()

	level := SysLog.CheckByTimeAndLevel("", "", "INFO")
	for _, log := range level {
		fmt.Println(log)
	}
}
