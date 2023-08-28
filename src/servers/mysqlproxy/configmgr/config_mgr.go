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
	err := conf.load("../conf/mysqlproxy.xml")
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
	RegistryConf RegistryConfig `xml:"registry"`
	ListenConf   ListenConfig   `xml:"listen"`
	MysqlConf    MysqlConfig    `xml:"mysql"`
}

type RegistryConfig struct {
	EtcdConf EtcdRegistryConfig `xml:"etcd,omitempty"`
}

type EtcdRegistryConfig struct {
	Endpoints struct {
		Items []string `xml:"item,omitempty"`
	} `xml:"endpoints,omitempty"`
	Username string `xml:"username,omitempty"`
	Password string `xml:"password,omitempty"`
}

type ListenConfig struct {
	Ip   string `xml:"ip,omitempty"`
	Port int    `xml:"port,omitempty"`
}

type MysqlConfig struct {
	IP          string `xml:"ip,omitempty"`
	Port        int    `xml:"port,omitempty"`
	Database    string `xml:"database,omitempty"`
	Username    string `xml:"username,omitempty"`
	Password    string `xml:"password,omitempty"`
	PoolMaxConn int    `xml:"pool_max_conn,omitempty"`
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
