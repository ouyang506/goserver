package config

import (
	"encoding/xml"
	"io"
	"os"
)

type Config struct {
	LoginServers LoginServers `xml:"login_servers"`
}

type LoginServers struct {
	LoginServer []AddrInfo `xml:"login_server"`
}

type AddrInfo struct {
	IP   string `xml:"ip,attr,omitempty"`
	Port int32  `xml:"port,attr,omitempty"`
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