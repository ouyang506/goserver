package config

import (
	"encoding/xml"
	"io/ioutil"
	"os"
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
