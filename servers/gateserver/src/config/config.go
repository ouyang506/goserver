package config

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"path"
)

type Config struct {
	Outer    OuterConfig    `xml:"outer"`
	Registry RegistryConfig `xml:"registry"`
}

type OuterConfig struct {
	IP   string `xml:"ip,omitempty"`
	Port int    `xml:"port,omitempty"`
}

type RegistryConfig struct {
	ServerType int32  `xml:"server_type,omitempty"`
	InstanceId int32  `xml:"instance_id,omitempty"`
	IP         string `xml:"ip,omitempty"`
	Port       int32  `xml:"port,omitempty"`
}

func NewConfig() *Config {
	conf := &Config{}
	return conf
}

func (conf *Config) Load(fileName string) error {
	// 如果查找不到在上一层查找
	f, err := os.Open(fileName)
	if err != nil && os.IsNotExist(err) {
		fileName = path.Join("../", fileName)
		f, err = os.Open(fileName)
	}

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
