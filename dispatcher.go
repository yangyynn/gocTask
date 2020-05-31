package main

import (
	"github.com/gorhill/cronexpr"
	"gocTask/models"
	"log"
	"time"
)

// Dispatcher 调度task和worker
type Dispatcher struct {
	// 接收task通道
	TaskChan chan *models.Task
	// 空闲worker通道
	WorkerChan chan chan *models.Task
	// 立即执行task队列
	TaskQueue []*models.Task
	// 等待执行task队列
	WaitQueue []*models.Task
	// 空闲worker队列
	WorkerQueue []chan *models.Task
}

// NewDispatcher 创建一个dispatcher， taskChanCount任务通道缓冲数
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		TaskChan:   make(chan *models.Task),
		WorkerChan: make(chan chan *models.Task),
	}
}

// WorkerReady 空闲的worker保存进WorkerChan
func (d *Dispatcher) WorkerReady(w chan *models.Task) {
	d.WorkerChan <- w
}

// Run 开始运行dispatcher等待taskChan的输入进行处理
func (d *Dispatcher) Run() {

	//后台运行队列调度分配，如果waitQueue中任务达到执行时间，则将任务分配到立即执行队列。
	d.waitQueueDispatcher()

	go func() {
		for {

			var activeTask *models.Task
			var activeWorker chan *models.Task
			if len(d.TaskQueue) > 0 && len(d.WorkerQueue) > 0 {
				//有task需要处理，并且有空闲的worker
				activeTask = d.TaskQueue[0]
				activeWorker = d.WorkerQueue[0]
			}

			select {

			case task := <-d.TaskChan:
				//任务放入队列
				d.addQueue(task)

			case worker := <-d.WorkerChan:
				//空闲worker放入队列
				d.WorkerQueue = append(d.WorkerQueue, worker)

			case activeWorker <- activeTask:
				//调度worker执行任务
				d.TaskQueue = d.TaskQueue[1:]
				d.WorkerQueue = d.WorkerQueue[1:]

			}

		}
	}()
}

// addQueue 将task加入不同的队列
func (d *Dispatcher) addQueue(task *models.Task) {
	if task.Crontab == "" {
		d.TaskQueue = append(d.TaskQueue, task)
	} else {
		//获取最近执行时间，存入task.NextTime
		task.NextTime = d.nextTime(task.Crontab)
		d.WaitQueue = append(d.WaitQueue, task)
	}
}

// queueDispatcher 等待到了执行时间调度分配任务 每秒一次任务调度
func (d *Dispatcher) waitQueueDispatcher() {
	go func() {
		// 每秒一次任务调度
		rateLimit := time.Tick(1000 * time.Millisecond)
		for {
			<-rateLimit

			nowUnix := time.Now().Unix()

			for i, task := range d.WaitQueue {
				//如果还未到任务开始时间，则跳过任务
				if task.StartTime != 0 && nowUnix < task.StartTime {
					continue
				}

				//如果现在时间超过endTime，task过期，则删除task。
				if task.EndTime != 0 && nowUnix > task.EndTime {
					d.WaitQueue = append(d.WaitQueue[:i], d.WaitQueue[i+1:]...)
					continue
				}

				//执行时间到，复制task，存入taskChan，更新原task的下次执行时间
				if task.NextTime == time.Now().Unix() {
					tmpTask := *task
					tmpTask.Crontab = ""
					d.TaskChan <- &tmpTask
					task.NextTime = d.nextTime(task.Crontab)
				}

			}
		}
	}()
}

// nextTime 获取task crontab对应的最近一次执行时间戳
func (d *Dispatcher) nextTime(c string) int64 {
	expr, err := cronexpr.Parse(c)
	nextTime := expr.Next(time.Now()).Unix()
	if err != nil {
		log.Printf("parse crontab %s, err: %v \n", c, err.Error())
	}
	return nextTime
}
