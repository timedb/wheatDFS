package etc

import (
	"errors"
	"fmt"
	"github.com/timedb/wheatDFS/serverTorch"
	"regexp"
	"sync"
)

var (
	SysConf      *Conf //总配置文件
	oneConf      sync.Once
	SysLocalHost *Addr //启动后服务器的地址
)

const (
	StateTracker = 0 //tracker
	StateStorage = 1 //storage
	StateDefault = 2 //default
)

// Addr 公用类
type Addr struct {
	Host      string //地址
	Port      string //端口
	StateType int
	Weight    int //本地的权重
}

// GetAddress 地址
func (a *Addr) GetAddress() string {
	return fmt.Sprintf("%s:%s", a.Host, a.Port)
}

// GetDjbHash 计算hash
func (a *Addr) GetDjbHash() int {

	hash := 0

	addr := a.GetAddress()
	for i, v := range addr {
		hash += int(v) << i
	}

	return hash
}

// MathDjbHah 计算字符串的hash
func MathDjbHah(s string) int {
	hash := 0
	for i, v := range s {
		hash += int(v) << i
	}

	return hash
}

// MakeAddr 新建地址
func MakeAddr(ip string, port string, state int) (*Addr, error) {

	if m, _ := regexp.MatchString(`^192|172|10|127|100|\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}$`, ip); !m {
		return nil, errors.New(fmt.Sprintf(IpError, ip))
	}

	var weight int
	if state == StateStorage { //只有Storage使用
		weight = serverTorch.GetServerWeight()
	}

	return &Addr{
		ip,
		port,
		state,
		weight,
	}, nil
}

//基础响应类
