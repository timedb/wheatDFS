package hashTorch

import (
	"fmt"
	"github.com/timedb/wheatDFS/etc"
	"io"
	"os"
	"testing"
)

func TestGetSmallHash(t *testing.T) {
	path := "C:\\Users\\HASEE\\Pictures\\哔哩哔哩动画\\1.jpg"
	data, _ := os.Open(path)
	defer data.Close()
	a, _ := GetSmallHash(data)
	fmt.Println(a)
}

func TestGetBigHash(t *testing.T) {
	etc.LoadConf("F:\\GO\\goDFS\\etc\\github.com/timedb/wheatDFS.ini")
	path := "F:\\学习\\汇编学习\\《汇编语言(第3版) 》王爽著.pdf"
	data, _ := os.Open(path)
	defer data.Close()
	a, _ := GetBigHash(data)
	fmt.Println(a)
}

func TestIntegrate(t *testing.T) {
	etc.LoadConf("F:\\GO\\goDFS\\etc\\github.com/timedb/wheatDFS.ini")
	//path := "F:\\1.txt"
	path := "F:\\学习\\汇编学习\\《汇编语言(第3版) 》王爽著.pdf"
	data, _ := os.Open(path)
	defer data.Close()
	a, _ := Integrate(data)
	fmt.Println(a)

	hf := MakeMaxFileHash()

	buf := make([]byte, 1024<<5)
	for {
		n, err := data.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
		}
		hf.Write(buf[:n])

	}

	fmt.Println(hf.GetHash())

}
