package clientTorch

import (
	"github.com/timedb/wheatDFS/etc"
	"strings"
)

//PathHostsToAddr 生成Addr
func PathHostsToAddr(ip string) (*etc.Addr, error) {

	str := strings.Split(ip, ":")
	if len(str) != 2 {
		err := ParameterErr
		return nil, err
	}
	addr, err := etc.MakeAddr(str[0], str[1], etc.StateDefault)
	if err != nil {
		return nil, err
	}

	return addr, err
}
