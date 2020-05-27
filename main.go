package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

var (
	Logger *logrus.Logger
)

// 任务Task结构
type Task struct {
	Id        int
	Title     string
	Crontab   string
	Command   string
	StartTime int64
	EndTime   int64
	Uid       string
	NotifyNum int
	NextTime  int64
}

func init() {
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)
	Logger.SetLevel(logrus.DebugLevel)
	Logger.Infoln("Init system.")
}

func main() {
	// todo 修改logger日志记录方式为，文件输出

	Logger.Infoln("Exec start")

	// 初始化数据传输通道
	Logger.Infoln("Create data channels.")
	var dispatcherChan = make(chan *Task, 10)
	var workerChan = make(chan *Task, 10)

	// 启动task解析调度器
	Logger.Infoln("Create dispatcherProcess.")
	go dispatcherProcess(dispatcherChan, workerChan)

	// 启动task任务执行器
	workerNum := 10
	Logger.Infoln("Create", workerNum, "workerProcess.")
	for i := 0; i < workerNum; i++ {
		go workerProcess(workerChan)
	}

	// 读取mysql中的task，程序启动的时候运行一次。
	getTasksInMysql(dispatcherChan)

	// 测试时保证不退出
	time.Sleep(1000 * time.Second)

}
