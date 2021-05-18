package hashTorch

import (
	"crypto/sha256"
	"fmt"
	"github.com/timedb/wheatDFS/etc"
	"io"
	"mime/multipart"
)

type MaxFileHash struct {
	buf  []byte //缓冲区
	hash string //最终hash
	temp []byte // 中间保存hash
	size int
}

// 流写入
func (f *MaxFileHash) Write(p []byte) {
	f.buf = append(f.buf, p...)
	if len(f.buf) < f.size {
		return
	}

	cont := f.buf[:f.size]
	vt := sha256.Sum256(cont) //动态声明
	f.buf = f.buf[f.size:]

	f.temp = append(f.temp, vt[:]...)

}

// 得到大文件哈希
func (f *MaxFileHash) GetHash() string {
	vt := sha256.Sum256(f.buf)
	f.buf = append(f.temp, vt[:]...)

	vt = sha256.Sum256(f.buf)

	return fmt.Sprintf("%x1", vt)

}

// makeFile
func MakeMaxFileHash() *MaxFileHash {
	f := new(MaxFileHash)
	f.size = int(etc.SysConf.StorageConf.UnitSize) << 10
	return f

}

// Integrate 统一哈希接口,对文件大小进行判断
func Integrate(l multipart.File) (string, error) {
	// 对传入文件进行大小判断
	sum := 0
	buf := make([]byte, 1024)
	for {
		n, err := l.Read(buf)
		sum += n
		if err == io.EOF || err != nil {
			break
		}
		if float64(sum) > etc.SysConf.StorageConf.UnitSize*1024 {
			l.Seek(0, io.SeekStart) //偏移
			hash, err := GetBigHash(l)
			l.Seek(0, io.SeekStart) //偏移

			return hash, err
		}
	}
	l.Seek(0, io.SeekStart) //偏移
	hash, err := GetSmallHash(l)
	l.Seek(0, io.SeekStart) //偏移
	return hash, err
}

// GetSmallHash 小文件哈希
func GetSmallHash(l io.Reader) (string, error) {
	content := make([]byte, 0, 1024<<5)
	for {
		MetaData := make([]byte, 1024<<5)
		n, err := l.Read(MetaData)
		if err != nil && err != io.EOF {
			return "", err
		} else if err == io.EOF {
			break
		}
		content = append(content, MetaData[:n]...)
	}
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x0", hash), nil
}

// GetBigHash 大文件哈希,读取到一定程度之后,进行一次哈hash,最后所有hash列表进行一次hash
func GetBigHash(l io.Reader) (string, error) {

	c := int(etc.SysConf.StorageConf.UnitSize) << 10 //*1024

	content := make([]byte, c)
	hashList := make([]byte, 0, c)
	for {
		n, err := l.Read(content)
		if err != nil {
			if err == io.EOF {
				break
			}

			return "", err
		}
		vt := sha256.Sum256(content[:n])
		hashList = append(hashList, vt[:]...)

	}
	// 计算哈希列表的哈希

	p := sha256.Sum256(hashList)

	return fmt.Sprintf("%x1", p), nil
}

// VerifyHash 检查hash是否错误
func VerifyHash(hash string) bool {
	if len(hash) != 65 || (hash[len(hash)-1] != '0' && hash[len(hash)-1] != '1') {
		return false
	}

	return true
}
