package main

import (
	"context"
	"gocTask/models"
	"os/exec"
	"time"
)

type Worker struct{}

var GWorker *Worker

func InitWorker() {
	GWorker = &Worker{}
}

func (w *Worker) Run(taskExecute *models.TaskExecute) {
	go func() {
		result := models.TaskResult{
			Task:   taskExecute.Task,
			Output: make([]byte, 0),
		}
		//接受到数据开始处理
		result.StartTime = time.Now()
		// 抢锁
		etcdMutex := &EtcdMutex{ttl: 10}
		err := etcdMutex.TryLock(taskExecute.Task.Title)
		defer etcdMutex.UnLock()
		if err != nil {
			// 没抢到锁，不执行任务
			result.Err = err
			result.EndTime = time.Now()
		} else {

			result.StartTime = time.Now()
			// 执行Task
			cmd := exec.CommandContext(context.TODO(), "/bin/bash", "-c", taskExecute.Task.Command)

			output, err := cmd.CombinedOutput()

			result.Output = output
			result.Err = err
			result.EndTime = time.Now()

		}
		GDis.receiveResult(&result)
	}()
}
