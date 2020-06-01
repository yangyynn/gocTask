package config

import (
	"encoding/json"
	"io/ioutil"
)

const (
	TASK_PREFIX   = "/cron/t/"
	TASK_LOCK_DIR = "/cron/l/"

	TASK_PUT_EVENT = 1
	TASK_DEL_EVENT = 2
)

type Config struct {
	WorkerNum       int      `json:"worker_num"`
	LogFile         string   `json:"log_file"`
	ApiServerAddr   string   `json:"api_server_addr"`
	ApiReadTimeout  int      `json:"api_server_read_timeout"`
	ApiWriteTimeout int      `json:"api_server_write_timeout"`
	EtcdServer      []string `json:"etcd_server"`
	EtcdDialTimeout int      `json:"etcd_dial_timeout"`
}

var GConfig *Config

func InitConfig(filename string) (err error) {
	var file []byte
	file, err = ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	config := Config{}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return
	}

	GConfig = &config

	return
}
