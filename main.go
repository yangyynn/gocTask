package main

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
	"time"
)

var (
	Logger      *logrus.Logger
	NotifyWx    = false
	NotifyEmail = false
)

// 任务Task结构
type Task struct {
	Id          int
	Title       string
	Crontab     string
	Command     string
	StartTime   int64
	EndTime     int64
	NotifyId    string
	NotifyEmail string
	NotifyNum   int
	NextTime    int64
}

func init() {
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)
	Logger.SetLevel(logrus.DebugLevel)
	Logger.Infoln("系统初始化")
}

func fileLog() {
	path := "/tmp/goc_task.log"
	//日志轮转相关函数
	//`WithLinkName` 为最新的日志建立软连接
	//`WithRotationTime` 设置日志分割的时间，隔多久分割一次
	//WithMaxAge 和 WithRotationCount二者只能设置一个
	//`WithMaxAge` 设置文件清理前的最长保存时间
	//`WithRotationCount` 设置文件清理前最多保存的个数

	// 下面配置日志每隔 1 分钟轮转一个新文件，保留最近 3 分钟的日志文件，多余的自动清理掉。
	writer, _ := rotatelogs.New(
		path+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithMaxAge(time.Duration(180)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(60)*time.Second),
	)
	Logger.SetOutput(writer)
}

func main() {
	// 记录程序运行pid
	if pid := syscall.Getpid(); pid != 1 {
		ioutil.WriteFile("./goc_task.pid", []byte(strconv.Itoa(pid)), 0777)

	}
	defer os.Remove("./goc_task.pid")

	fileLog() //生成环境，修改logger日志记录方式为，文件输出

	Logger.Infoln("开始运行")

	// 初始化数据传输通道
	Logger.Infoln("创建数据channel")
	var dispatcherChan = make(chan *Task, 10)
	var workerChan = make(chan *Task, 10)

	// 启动task解析调度器
	Logger.Infoln("启动任务调度处理器")
	go dispatcherProcess(dispatcherChan, workerChan)

	// 启动task任务执行器
	workerNum := 10
	Logger.Infoln("启动", workerNum, "个任务处理器")
	for i := 0; i < workerNum; i++ {
		go workerProcess(workerChan)
	}

	// 读取mysql中的task，程序启动的时候运行一次。
	getTasksInMysql(dispatcherChan)

	// 测试时保证不退出
	time.Sleep(1000 * time.Second)

}
