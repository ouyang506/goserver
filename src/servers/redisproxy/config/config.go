package config

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

type ConfigMgr struct {
	conf *Config
}

// singleton
func GetConfigMgr() *ConfigMgr {
	once.Do(func() {
		confMgr = newConfigMgr()
		conf := newConfig()
		err := conf.load("../../../conf/redisproxy.xml")
		if err != nil {
			fmt.Printf("load config error: %v", err)
		} else {
			confMgr.conf = conf
		}
	})
	return confMgr
}

func GetConfig() *Config {
	return GetConfigMgr().conf
}

func newConfigMgr() *ConfigMgr {
	mgr := &ConfigMgr{}
	return mgr
}

type Config struct {
	RegistryConf RegistryConfig `xml:"registry"`
	ListenConf   ListenConfig   `xml:"listen"`
	RedisConf    RedisConfig    `xml:"redis"`
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

type RedisConfig struct {
	Type      int `xml:"type,omitempty"`
	Endpoints struct {
		Items []string `xml:"item,omitempty"`
	} `xml:"endpoints,omitempty"`
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
