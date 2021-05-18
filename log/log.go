package log

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/timedb/wheatDFS/etc"
	"os"
	"runtime"
	"strconv"
	"time"
)

type Logger interface {
	checkError()
	conn()
	creatTable()
	Add(msg string, msgLevel int)
	Del(ordinal []int)
	CheckByTime(startTime string, endTime string) (massages []DetailLog)
	CheckByLevel(level string) (massages []DetailLog)
}

type Logging struct {
	db        *sql.DB //数据库链接
	err       error   //操作错误
	debugType bool    //debug
}

func MakeLogging() *Logging {

	// 单实例
	oneLock.Do(func() {
		logConf := new(Logging)

		logConf.debugType = etc.SysConf.Debug
		logConf.conn()

		//判断 debug是否为 true
		if !logConf.debugType {
			//建立链接，创建数据库和表

			logConf.creatTable()
		}

		SysLog = logConf
	})

	return SysLog
}

//将数据库处理错误保存到数据库中
func (log Logging) checkError() {
	if log.err != nil {
		st := fmt.Sprintf("%s", log.err)
		log.Add(st, Panic)
		panic(log.err)
	}
}

//链接数据库
func (log *Logging) conn() {
	path := etc.SysConf.LogConf.LogPath

	log.db, log.err = sql.Open("sqlite3", path)
	log.checkError()
}

//创建日志数据库
func (log *Logging) creatTable() {
	rows, _ := log.db.Query(
		`SELECT name FROM sqlite_master where name = 'log_save'`)
	if !rows.Next() {
		_, log.err = log.db.Exec(`create table "log_save"(
				"uid" integer primary key autoincrement,
				"log_level" varchar(256),
				"msg" text,
				"log_time" timestamp default (datetime('now', 'localtime')),
				"log_place" varchar(256)
			)`)
		log.checkError()
	}
}

// Add 插入数据
func (log *Logging) Add(msg string, msgLevel int) {

	var level string
	switch msgLevel {
	case Info:
		level = "INFO"
	case Error:
		level = "ERROR"
	case Panic:
		level = "PANIC"
	}

	_, path, line, _ := runtime.Caller(1)
	logPlace := path + "\tline:" + strconv.Itoa(line)
	if log.debugType {
		fmt.Println("level:", level, "msg:", msg, "place:", logPlace)
	} else {
		sqlS := fmt.Sprintf(
			`insert into log_save(log_level, msg, log_place) values ('%s', '%s', '%s')`, level, msg, logPlace)
		_, log.err = log.db.Exec(sqlS)
	}
	log.checkError()
}

// panic 保护函数使用
func (log *Logging) panic(msg string, msgLevel int) {
	var level string
	switch msgLevel {
	case Info:
		level = "Info"
	case Error:
		level = "Error"
	case Panic:
		level = "Panic"
	}

	_, path, line, _ := runtime.Caller(2)
	logPlace := path + "\tline:" + strconv.Itoa(line)
	if log.debugType {
		fmt.Println("level:", level, "msg:", msg, "place:", logPlace)
	} else {
		sqlS := fmt.Sprintf(
			`insert into log_save(log_level, msg, log_place) values ('%s', '%s', '%s')`, level, msg, logPlace)
		_, log.err = log.db.Exec(sqlS)
	}
	log.checkError()
}

// Del 根据日志的 uid 删除日志
func (log *Logging) Del(ordinal []int) {
	for _, li := range ordinal {
		_, log.err = log.db.Exec(fmt.Sprintf("delete from log_save where uid = '%d'", li))
		log.checkError()
	}
}

type DetailLog struct {
	id       int
	Level    string
	LogMsg   string
	LogTime  string
	LogPlace string
}

// 时间相关sql语句格式化
func timeQueFormat(startTime string, endTime string) (sqlS string) {
	//将时间字符串转为 Time 类型，且设置为本地时间
	stData, _ := time.ParseInLocation("2006-01-02 15:04", startTime, time.Local)
	edData, _ := time.ParseInLocation("2006-01-02 15:04", endTime, time.Local)

	//判断输入的时间是否存在空字符串 并做出相应处理
	if startTime == "" && endTime == "" {
		sqlS = fmt.Sprintf("select * from log_save")
	} else if endTime == "" {
		sqlS = fmt.Sprintf("select * from log_save where log_time >= '%s'", stData)
	} else if startTime == "" {
		sqlS = fmt.Sprintf("select * from log_save where log_time <= '%s'", edData)
	} else {
		sqlS = fmt.Sprintf("select * from log_save where log_time >= '%s' and log_time <= '%s'", stData, edData)
	}

	return sqlS
}

// CheckByTime 日志储存时间查找  startTime开始时间 endTime结束时间
func (log *Logging) CheckByTime(startTime string, endTime string) (massages []DetailLog) {
	var rows *sql.Rows

	rows, log.err = log.db.Query(timeQueFormat(startTime, endTime))

	log.checkError()
	for rows.Next() {
		var dl DetailLog
		//读取数据到 DetailLog 结构体中
		rows.Scan(&dl.id, &dl.Level, &dl.LogMsg, &dl.LogTime, &dl.LogPlace)
		massages = append(massages, dl)
	}
	return massages
}

// CheckByLevel 日志等级查找 level 日志等级
func (log *Logging) CheckByLevel(level string) (massages []DetailLog) {
	var rows *sql.Rows
	sqlS := fmt.Sprintf("select * from log_save where log_level = '%s'", level)
	rows, log.err = log.db.Query(sqlS)
	log.checkError()
	for rows.Next() {
		var dl DetailLog
		//读取数据到 DetailLog 结构体中
		rows.Scan(&dl.id, &dl.Level, &dl.LogMsg, &dl.LogTime, &dl.LogPlace)
		massages = append(massages, dl)
	}
	return massages
}

// CheckByTimeAndLevel 获取数据接口
func (log *Logging) CheckByTimeAndLevel(startTime string, endTime string, level string) []string {

	massages := log.CheckByTime(startTime, endTime)
	buf := make([]string, 0, len(massages))
	if level == "" {
		for _, massage := range massages {
			// level 为空返回全部时间区域内的log
			buf = append(buf, fmt.Sprintf("%s\t%s\t%s\t%s",
				massage.Level, massage.LogTime, massage.LogMsg, massage.LogPlace))
		}
		return buf
	} else {
		for _, massage := range massages {
			// level 为空返回指定等级时间区域内的log
			if massage.Level == level {
				buf = append(buf, fmt.Sprintf("%s\t%s\t%s\t%s",
					massage.Level, massage.LogTime, massage.LogMsg, massage.LogPlace))
			}
		}

		return buf
	}
}

// Exit 退出程序
func (log *Logging) Exit() {
	os.Exit(0)
}
