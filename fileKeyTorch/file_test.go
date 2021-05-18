package fileKeyTorch

import (
	"fmt"
	"github.com/timedb/wheatDFS/etc"
	"github.com/timedb/wheatDFS/log"
	"github.com/timedb/wheatDFS/torch/hashTorch"
	"io"
	"os"
	"testing"
)

func TestMakeFileKeyByHash(t *testing.T) {
	etc.LoadConf("D:\\goproject\\github.com/timedb/wheatDFS\\etc\\github.com/timedb/wheatDFS.ini")
	log.MakeLogging()

	path := "D:\\goproject\\github.com/timedb/wheatDFS\\go.mod"
	data, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer data.Close()
	a, _ := hashTorch.Integrate(data)
	fmt.Println(a)
	key := MakeFileKeyByHash(a, "jpg")
	fmt.Println(MakeFileKeyByToken(key.token))

}

//测试回调
func TestFileKey_Seek(t *testing.T) {
	etc.LoadConf("/home/lgq/code/go/goDFS/etc/github.com/timedb/wheatDFS.ini")
	hash := "820e0d99bcbcf8ff43c59ebee0d773073856b26b92ea4c0f7c0221a8107438f80"
	key := MakeFileKeyByHash(hash, "txt")
	bufs, _ := key.ReadAll()
	fmt.Println(string(bufs))

}

//测试大文件保存，写入
func TestMaxFuleKey(t *testing.T) {
	etc.LoadConf("D:\\goproject\\github.com/timedb/wheatDFS\\etc\\github.com/timedb/wheatDFS.ini")

	path := "F:\\视频\\MyTv\\fodfs\\安装.mp4"

	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	hash, err := hashTorch.Integrate(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(hash)

	fk := MakeFileKeyByHash(hash, "mp4")
	defer fk.Close()

	buf := make([]byte, int(1024*etc.SysConf.StorageConf.UnitSize))
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			return
		}

		fk.Write(buf[:n])

	}

	fmt.Println(fk.token)

}

//大文件处理， 读取
func TestFileReadMan_Close(t *testing.T) {
	etc.LoadConf("D:\\goproject\\github.com/timedb/wheatDFS\\etc\\github.com/timedb/wheatDFS.ini")

	token := "group/cb/e5/f3da59410da46eb6b353beb4b2617e411.mp4"
	tk := MakeFileKeyByToken(token)

	path, _ := os.Create("F:\\视频\\MyTv\\fodfs\\copy\\copy1.mp4")
	defer path.Close()

	//测试接口用的谁用，挨打
	//buf, _ := ioutil.ReadAll(tk)
	//path.Write(buf)

	//大文件分块传输
	buf := make([]byte, int(1024*etc.SysConf.StorageConf.UnitSize))
	for {
		n, err := tk.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
		}

		path.Write(buf[:n])

	}

}

func TestMakeFileKeyByToken(t *testing.T) {
	etc.LoadConf("D:\\goproject\\github.com/timedb/wheatDFS\\etc\\github.com/timedb/wheatDFS.ini")
	token := "group/6c/a2/052f160713310849cbe13582a76eca1e0.md"
	fk := MakeFileKeyByToken(token)
	fmt.Println(fk)

}
