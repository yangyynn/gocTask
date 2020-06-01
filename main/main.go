package main

import (
	"flag"
	"fmt"
	"gocTask"
	"gocTask/api"
	"runtime"
	"time"
)

// initEnv 初始化运行环境
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

var configFile string

// initArgs 初始化命令行参数
func initArgs() {
	flag.StringVar(&configFile, "configFile", "./conf.json", "配置文件conf.go地址")
	flag.Parse()
}

func main() {
	var err error

	initEnv()

	initArgs()
	err = gocTask.InitConfig(configFile)
	if err != nil {
		goto ERR
	}

	// 开启api服务，等待任务处理。
	err = api.InitServer()
	if err != nil {
		goto ERR
	}

	for {
		time.Sleep(1000 * time.Millisecond)
	}

	return

ERR:
	fmt.Println(err)

}
