package configmgr

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	once               = sync.Once{}
	confMgr *ConfigMgr = nil
)

// singleton
func Instance() *ConfigMgr {
	once.Do(func() {
		confMgr = newConfigMgr()
	})
	return confMgr
}

type ConfigMgr struct {
	conf *Config
}

func newConfigMgr() *ConfigMgr {
	mgr := &ConfigMgr{}
	conf := newConfig()
	err := conf.load("../../../conf/login.xml")
	if err != nil {
		fmt.Printf("load config error: %v\n", err)
	} else {
		mgr.conf = conf
	}
	return mgr
}

func (mgr *ConfigMgr) GetConfig() *Config {
	return mgr.conf
}

type Config struct {
	HttpListen HttpListenAddr `xml:"http_listen"`
}

type HttpListenAddr struct {
	IP   string `xml:"ip,omitempty"`
	Port int32  `xml:"port,omitempty"`
}

func newConfig() *Config {
	conf := &Config{}
	return conf
}

func (conf *Config) load(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(content, conf)
	if err != nil {
		return err
	}

	return nil
}
