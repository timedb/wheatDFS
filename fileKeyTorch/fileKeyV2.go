package fileKeyTorch

import (
	"fmt"
	"github.com/timedb/wheatDFS/etc"
	serverTorch2 "github.com/timedb/wheatDFS/serverTorch"
	"github.com/timedb/wheatDFS/torch/hashTorch"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FileKeyFace interface {
	Write(p []byte) (int, error)
	Read(p []byte) (int, error)
}

// FileKey 并发不安全的文件令牌
type FileKey struct {
	g1        int64
	g2        int64
	types     int
	Hash      string
	token     string
	file      *os.File //文件指针
	groupPath string
	offset    int64        //文件偏移量
	extension string       //扩展名
	name      string       //文件名
	cacheTime time.Time    //创建时间
	lockTime  sync.RWMutex //时间读写锁
}

///生成令牌
func (f *FileKey) createKey() {
	f.token = fmt.Sprintf("group/%02x/%02x/%s.%s", f.g1, f.g2, f.Hash[32:], f.extension)
}

//计算当前文件地址
func (f *FileKey) mathNowPath() string {

	//计算偏移后的g2地址
	g2 := f.g2 + f.offset
	var n int64
	if g2 > 0xff {
		n = g2 / 0xff
		g2 = g2 % 0xff
	}

	//计算偏移后的g1地址
	g1 := f.g1 + n
	if g1 > 0xff {
		g1 = g1 % 0xff
	}

	return fmt.Sprintf("%s/%02x/%02x/%s", f.groupPath, g1, g2, f.name)
}

//后移指针，只对大文件有效
func (f *FileKey) pointerBack() {
	if f.types == MinFile {
		return
	}

	f.offset += 1

}

//用户接口

//写入内容, 自动打开文件
func (f *FileKey) Write(p []byte) (int, error) {
	if f.file == nil {
		file, err := os.Create(f.mathNowPath())
		if err != nil {
			return 0, err
		}
		if f.types == MaxFile {
			defer f.Close() //关闭文件, 大文件时候生效
		}

		f.file = file
	}

	n, err := f.file.Write(p)
	if err == nil {
		f.pointerBack()
	}

	return n, err
}

// WriteByFile TODO 自动读入文件数据(目前不做处理)
func (f *FileKey) writeByFile(reader io.Reader) error {
	return nil
}

//读取文件接口
func (f *FileKey) Read(p []byte) (int, error) {

	if f.file == nil {
		file, err := os.Open(f.mathNowPath())
		if err != nil {
			//大文件中打开文件失败，认为到达尾部
			if f.types == MaxFile {
				return 0, io.EOF
			}

			return 0, err
		}
		f.file = file

	}
	//读取完毕后，都应该关闭当前的文件指针

	if f.types == MinFile {
		n, err := f.file.Read(p)
		if err != nil {
			f.Close()
			f.file = nil
			return n, err
		}
	} else {
		n, err := f.file.Read(p)
		if err != nil {
			defer f.Close() //关闭单曲指针
			//大文件处理错误需要后移
			if err == io.EOF {
				f.pointerBack() //移到下一位文件

				return 0, nil
			}
			return n, err
		}

		return n, err

	}

	return 0, nil

}

// ReadMaxFileCurrent 对大文件读取当前文件
func (f *FileKey) ReadMaxFileCurrent() ([]byte, error) {
	if f.Types() == MinFile {
		return nil, MaxFileSizeErr
	}
	f.Close() //先关闭文件

	file, err := os.Open(f.mathNowPath())
	if err != nil {
		return nil, io.EOF
	}

	defer file.Close() //完毕以后在关闭文件指针

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	f.pointerBack() //指针后移

	return buf, nil

}

// GeyToken 返回token
func (f *FileKey) GeyToken() string {
	return f.token
}

// ReadAll 读取全部的文件内容，关闭文件
func (f *FileKey) ReadAll() ([]byte, error) {
	if f.types == MaxFile {
		return nil, SmallFileSizeErr
	}

	if f.file == nil {
		file, err := os.Open(f.mathNowPath())
		if err != nil {
			return nil, err
		}
		f.file = file

	}

	b, err := ioutil.ReadAll(f.file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return b, nil

}

// Seek 相对于开始位置的偏移量
func (f *FileKey) Seek(offset int64) error {
	if offset < 0 {
		return OffsetErr
	}

	//对小文件就行偏移
	if f.types == MinFile && f.file != nil {
		_, err := f.file.Seek(offset, io.SeekStart) //相对于文件头
		return err
	}

	f.offset = offset
	return nil
}

// Open 打开文件
func (f *FileKey) Open() error {
	if f.file == nil {
		file, err := os.Open(f.mathNowPath())
		if err != nil {
			return err
		}
		f.file = file
	}

	return nil
}

// Close 关闭文件
func (f *FileKey) Close() {
	if f != nil {
		f.file.Close()
	}
	f.file = nil
}

// GetOffset 获取偏移量
func (f *FileKey) GetOffset() int64 {
	return f.offset
}

//令牌格式: group/a1/a2/3856b26b92ea4c0f7c0221a8107438f80.jpg

// MakeFileKeyByHash 使用hash来创建文件令牌
func MakeFileKeyByHash(hash string, ext string) *FileKey {
	//检查hash是否正确
	if !hashTorch.VerifyHash(hash) {
		return nil
	}

	//计算保存地址
	g1, err := strconv.ParseInt(hash[:2], 16, 64)
	if err != nil {
		return nil
	}

	g2, err := strconv.ParseInt(hash[2:4], 16, 64)
	if err != nil {
		return nil
	}

	fk := new(FileKey)
	fk.g1 = g1
	fk.g2 = g2
	fk.Hash = hash

	//构建令牌
	fk.extension = strings.Replace(ext, `.`, ``, 1) //去掉.
	fk.createKey()
	fk.name = fmt.Sprintf("%s.%s", hash[32:], fk.extension)

	if hash[len(hash)-1] == '0' {
		fk.types = MinFile
	} else if hash[len(hash)-1] == '1' {
		fk.types = MaxFile
	} else {
		return nil
	}
	fk.groupPath = serverTorch2.JoinPath(etc.SysConf.StorageConf.GroupPath)

	fk.cacheTime = time.Now()

	return fk

}

// MakeFileKeyByToken 使用令牌来创建文件令牌
func MakeFileKeyByToken(token string) *FileKey {
	args := strings.Split(token, "/")
	if len(args) != 4 {
		return nil
	}

	fk := new(FileKey)
	fk.token = token

	g1, err := strconv.ParseInt(args[1], 16, 64)
	if err != nil {
		return nil
	}

	g2, err := strconv.ParseInt(args[2], 16, 64)
	if err != nil {
		return nil
	}

	fk.g1 = g1
	fk.g2 = g2
	fk.name = args[3]

	if args[3][32] == '0' {
		fk.types = MinFile
	} else if args[3][32] == '1' {
		fk.types = MaxFile
	} else {
		return nil
	}
	fk.groupPath = serverTorch2.JoinPath(etc.SysConf.StorageConf.GroupPath)
	fk.cacheTime = time.Now()

	args = strings.Split(fk.name, ".")
	fk.extension = args[len(args)-1] //获取到后缀

	return fk

}

// ResetTime 更新当前时间
func (f *FileKey) ResetTime() {
	f.lockTime.RLock()
	defer f.lockTime.RUnlock()
	f.cacheTime = time.Now()
}

// PastDue 验证是否过期
func (f *FileKey) PastDue() bool {
	f.lockTime.RLock()
	defer f.lockTime.RUnlock()

	now := time.Now()
	if now.Second()-f.cacheTime.Second() > 60*etc.SysConf.StorageConf.MaxCacheTime {
		return false
	}

	return true

}

// Types 返回文件类型
func (f *FileKey) Types() int {
	return f.types
}

// Ext 获取扩展名
func (f *FileKey) Ext() string {
	return f.extension
}
