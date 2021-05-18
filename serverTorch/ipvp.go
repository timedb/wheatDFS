package serverTorch

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

// GetAvailablePort 获取一个可用的本地端口
func GetAvailablePort() (string, error) {

	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return "0", err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return "0", err
	}

	defer listener.Close()
	return strconv.Itoa(listener.Addr().(*net.TCPAddr).Port), nil

}

// GetIPv4s 获取一个本机的局域网的ip
func GetIPv4s() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}
