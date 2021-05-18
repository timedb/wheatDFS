package tracker

import (
	"github.com/timedb/wheatDFS/fileKeyTorch"
	"sync"
)

// MaxFileCache 大文件缓存器
type MaxFileCache struct {
	data *sync.Map //储存地址
}

func (m *MaxFileCache) Store(key string, value *fileKeyTorch.FileKey) {
	m.data.Store(key, value)
}
