package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"gocTask/config"
	"os"
	"runtime"
	"time"
)

var (
	// GArgs 全局变量，启动时命令行配置的参数map
	GArgs = make(map[string]string)
	// GLog 日志全局变量
	GLog *logrus.Logger
)

// initEnv init environment
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// initArgs
func initArgs() {
	var filename string
	flag.StringVar(&filename, "config", "./config/config.json", "input file config.json path")
	flag.Parse()
	GArgs["filename"] = filename
}

// initLog
func initLog() {
	GLog = logrus.New()
	GLog.Level = logrus.DebugLevel
	GLog.Out = os.Stdout

	file, err := os.OpenFile(config.GConfig.LogFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		GLog.Out = file
	} else {
		GLog.Infof("Failed to log to file %s, using default stderr", config.GConfig.LogFile)
	}
}

func init() {
	// 初始化环境
	initEnv()
	// 读取命令行配置
	initArgs()
	// 读取config.json配置
	err := config.InitConfig(GArgs["filename"])
	if err != nil {
		GLog.Panicln(err)
	}
	// 配置log记录
	initLog()

}

func main() {
	// 开启apiServer服务
	GLog.Infoln("init api server")
	err := InitServer()
	if err != nil {
		GLog.Panicln(err)
	}

	// 连接Etcd
	GLog.Infoln("init Etcd server")
	err = InitEtcd()
	if err != nil {
		GLog.Panicln(err)
	}

	//init 调度器
	GLog.Infoln("init dispatcher")
	InitDispatcher()

	//init Worker
	GLog.Infoln("init dispatcher")
	InitWorker()

	for {
		time.Sleep(1 * time.Second)
	}
}
