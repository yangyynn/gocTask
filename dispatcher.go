package main

import (
	"github.com/rfyiamcool/cronlib"
	"github.com/sirupsen/logrus"
	"time"
)

var rateLimiter = time.Tick(500 * time.Millisecond)

func dispatcherProcess(d chan *Task, w chan *Task) {
	// waitQueue 等待下次运行任务
	var waitQueue = make([]*Task, 0)
	for {
		<-rateLimiter //限制为每半秒运行一次

		select {

		case task := <-d:
			//判断是否已有任务
			if hasTask(task.Id, waitQueue) {
				Logger.Infof("dispatcher task id %d was exits", task.Id)
				break
			}

			//通道接收到需要调度的任务，记录任务下次运行时间，并存入waitQueue队列
			nextTime := taskNextDoTimeUnix(task)
			if nextTime == 0 {
				Logger.Infof("dispatcher task[%d] failed, because next time is 0", task.NextTime)
			} else {
				task.NextTime = nextTime
				waitQueue = append(waitQueue, task)
				_, err := updateRunStatus(task.Id, map[string]interface{}{"t_run_status": "1"})
				if err != nil {
					Logger.WithFields(logrus.Fields{
						"taskId": task.Id,
						"err":    err,
					}).Warningln("update run_status failed")
				}
			}

		default:
			//获取任务是否到执行时间，如果到了，则发送给workerChan。
			now := time.Now().Unix()

			for i, task := range waitQueue {

				if task.StartTime != 0 && task.StartTime > now {
					// 还未到任务开始运行时间
					continue
				}

				if task.EndTime != 0 && task.EndTime < now {
					// 已超过任务运行结束时间
					waitQueue = append(waitQueue[:i], waitQueue[i+1:]...)
					continue
				}

				if task.NextTime == now {
					w <- task
				}

				nextTime := taskNextDoTimeUnix(task)
				if nextTime == 0 {
					Logger.Infof("task [%d] is delete because next time is 0", task.NextTime)
					//删除任务
					waitQueue = append(waitQueue[:i], waitQueue[i+1:]...)
				} else {
					task.NextTime = nextTime
				}
			}
		}
	}
}

// hasTask 判断任务是否已存在
func hasTask(id int, queue []*Task) bool {
	for _, task := range queue {
		if task.Id == id {
			return true
		}
	}
	return false
}

// hasTask 返回下一次任务运行时间，如果没有则返回 0
func taskNextDoTimeUnix(task *Task) int64 {
	//解析task crontab
	parse, err := cronlib.Parse(task.Crontab)
	if err != nil {
		Logger.WithFields(logrus.Fields{
			"crontab": task.Crontab,
			"err":     err,
		}).Warningln("parse failed")
	}
	nt := parse.Next(time.Now())

	if nt.IsZero() {
		return 0
	} else {
		return nt.Unix()
	}
}
