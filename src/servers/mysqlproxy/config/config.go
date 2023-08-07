package config

import (
	"encoding/xml"
	"io"
	"os"
)

type Config struct {
	Registry  RegistryConfig `xml:"registry"`
	MysqlConf MysqlConfig    `xml:"mysql"`
}

type RegistryConfig struct {
	IP   string `xml:"ip,omitempty"`
	Port int32  `xml:"port,omitempty"`
}

type MysqlConfig struct {
	Host     string `xml:"host,omitempty"`
	Database string `xml:"database,omitempty"`
	Username string `xml:"username,omitempty"`
	Password string `xml:"password,omitempty"`
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
