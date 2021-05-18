package etc

import (
	"github.com/BurntSushi/toml"
	"log"
)

// LoadConf 读取配置文件
func LoadConf(path string) {

	//单例模式
	oneConf.Do(func() {
		conf := new(Conf)
		if _, err := toml.DecodeFile(path, conf); err != nil {
			log.Panic(err)
		}

		conf.ExamineMustConf()

		//初始化
		addr, err := MakeAddr(conf.TrackerConf.CandidateIp, conf.TrackerConf.CandidatePort, StateTracker)
		if err != nil {
			log.Fatal(err)

		}
		conf.TrackerConf.CandidateHost = addr

		SysConf = conf
	})

	return
}

// ExamineMustConf 检查配置文件的情况
func (c *Conf) ExamineMustConf() {
	temTracker := "Check whether the tracker exists persistence_path, ip, port"
	temStorage := "Check whether the storage exists group_path"

	if trc := c.TrackerConf; trc.CandidateIp == "" || trc.PersistencePah == "" || trc.CandidatePort == "" {
		defer log.Println(temTracker)
		log.Fatal(ConfMustNotInErr)
	}

	if stor := c.StorageConf; stor.GroupPath == "" {
		defer log.Panic(temStorage)
		log.Fatal(ConfMustNotInErr)
	}

}
