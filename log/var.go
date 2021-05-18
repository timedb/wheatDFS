package log

import "sync"

const (
	Info  = 1
	Error = 2
	Panic = 3
)

var (
	oneLock sync.Once
	SysLog  *Logging
)

//func init() {
//	SysLog = MakeLogging()
//}
