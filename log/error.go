package log

import (
	"fmt"
	"runtime"
)

const (
	DbNoExistErr = "Database does not exist"
)

// ProtectRun 保护方式运行一个函数
func ProtectRun() {
	// 延迟处理的函数
	// 发生宕机时，获取panic传递的上下文并打印
	err := recover()
	switch err.(type) {
	case runtime.Error: // 运行时错误
		SysLog.panic(fmt.Sprintf("%v", err), Panic)
	}
}
