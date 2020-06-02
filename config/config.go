package config

import (
	"encoding/json"
	"io/ioutil"
)

const (
	TASK_TASK_DIR      = "/cron/t/"
	TASK_LOCK_DIR      = "/cron/l/"
	TASK_KILL_DIR      = "/cron/k/"
	TASK_ALARM_NUM_DIR = "/cron/a/"

	TASK_PUT_EVENT  = 1
	TASK_DEL_EVENT  = 2
	TASK_KILL_EVENT = 3
)

type Config struct {
	WorkerNum          int      `json:"worker_num"`
	LogFile            string   `json:"log_file"`
	ApiServerAddr      string   `json:"api_server_addr"`
	ApiReadTimeout     int      `json:"api_server_read_timeout"`
	ApiWriteTimeout    int      `json:"api_server_write_timeout"`
	EtcdServer         []string `json:"etcd_server"`
	EtcdDialTimeout    int      `json:"etcd_dial_timeout"`
	AlarmMaxNum        int      `json:"alarm_max_num"`
	AlarmType          []string `json:"alarm_type"`
	QyweixinCorpId     string   `json:"qyweixin_corp_id"`
	QyweixinCorpSecret string   `json:"qyweixin_corp_secret"`
	QyweixinAgentid    int      `json:"qyweixin_agentid"`
	EmailUser          string   `json:"email_user"`
	EmailPass          string   `json:"email_pass"`
	EmailHost          string   `json:"email_host"`
	EmailPort          string   `json:"email_port"`
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
