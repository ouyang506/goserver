package config

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

type Config struct {
	Gates Gates `xml:"gates"`
}

type Gates struct {
	GateInfos []GateAddr `xml:"gate"`
}

type GateAddr struct {
	IP   string `xml:"ip,attr,omitempty"`
	Port int32  `xml:"port,attr,omitempty"`
}

func NewConfig() *Config {
	conf := &Config{}
	return conf
}

func (conf *Config) Load(fileName string) error {
	// 如果查找不到在上一层查找
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
