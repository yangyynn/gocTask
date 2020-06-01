package gocTask

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ApiPort         int `json:"api_port"`
	ApiReadTimout   int `json:"api_read_timout"`
	ApiWriteTimeout int `json:"api_write_timeout"`
}

var GConfig *Config

func InitConfig(filename string) (err error) {
	var configContent []byte
	var conf Config

	// 配置读取
	if configContent, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	if err = json.Unmarshal(configContent, &conf); err != nil {
		return
	}

	GConfig = &conf

	return
}
