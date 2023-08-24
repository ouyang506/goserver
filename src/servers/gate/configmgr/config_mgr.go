package configmgr

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
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
func Instance() *ConfigMgr {
	once.Do(func() {
		confMgr = newConfigMgr()
		conf := newConfig()
		err := conf.load("../../../conf/gate.xml")
		if err != nil {
			fmt.Printf("load config error: %v", err)
		} else {
			confMgr.conf = conf
		}
	})
	return confMgr
}

func newConfigMgr() *ConfigMgr {
	mgr := &ConfigMgr{}
	return mgr
}

func (mgr *ConfigMgr) GetConfig() *Config {
	return mgr.conf
}

type Config struct {
	RegistryConf RegistryConfig `xml:"registry"`
	ListenConf   ListenConfig   `xml:"listen"`
	Outer        OuterConfig    `xml:"outer"`
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

type OuterConfig struct {
	OuterIp  string `xml:"outer_ip,omitempty"`
	ListenIp string `xml:"listen_ip,omitempty"`
	Port     int    `xml:"port,omitempty"`
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

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(content, conf)
	if err != nil {
		return err
	}

	return nil
}
