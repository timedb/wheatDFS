package etc

import "time"

type Conf struct {
	TrackerConf *Tracker `toml:"tracker"`
	StorageConf *Storage `toml:"storage"`
	LogConf     *Log     `toml:"log"`
	Version     string   `toml:"version"`
	Debug       bool     `toml:"debug"`
	Pool        *RpcPool `toml:"pool"`
	Client      *Client  `toml:"client"`
}

// Tracker tracker配置文件
type Tracker struct {
	PersistencePah string `toml:"persistencePath"`
	CandidateIp    string `toml:"ip"` //初始继承者
	CandidatePort  string `toml:"port"`
	CandidateHost  *Addr  //地址集
	Port           string `toml:"traPort"`
	EsotericaPath  string `toml:"esotericaPath"`
	SyncMaxCount   int    `toml:"syncMaxCount"`
}

// Storage storage配置文件
type Storage struct {
	GroupPath       string  `toml:"groupPath"`
	MaxStorageCount int     `toml:"maxCount"`     //最大storage接入数
	UnitSize        float64 `toml:"unitSize"`     //最小文件单位大小, 小文件的大小kb
	MaxCacheTime    int     `toml:"maxCacheTime"` //最大缓存时间、分钟
	Port            string  `toml:"port"`
	CachePath       string  `toml:"cachePath"` //文件缓存地址
}

// Log 日志配置文件
type Log struct {
	LogPath string `toml:"logPath"`
}

type RpcPool struct {
	TimeOut      time.Duration `toml:"timeOut"`      //最大连接时间，秒
	InitConnNum  int           `toml:"initConnNum"`  //初始连接数
	MaxConnNum   int           `toml:"maxConnNum"`   //最大了解数
	MaxReConnNum int           `toml:"maxReConnNum"` //reConn number
}

type Client struct {
	Port      string `toml:"port"`
	CachePath string `toml:"cachePath"` //文件缓存地址
}
