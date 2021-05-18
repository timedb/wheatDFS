package etc

import (
	"fmt"
	"testing"
)

func TestLoadConf(t *testing.T) {
	LoadConf("./wheatDFS.ini")
	fmt.Println(SysConf.TrackerConf.CandidateHost)
}

func TestMakeAddr(t *testing.T) {
	ip := "127.0.12.13"
	addr, err := MakeAddr(ip, "20000", StateStorage)
	fmt.Println(addr, err)
	hash := addr.GetDjbHash()
	fmt.Println(hash)
}
