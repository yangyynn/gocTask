package main

import (
	"gocTask/config"
	"gocTask/models"
	"math/rand"
	"os/exec"
	"time"
)

type Worker struct{}

var (
	GWorker *Worker
)

func InitWorker() {
	GWorker = &Worker{}
}

func (w *Worker) Run(taskExecute *models.TaskExecute) {
	go func() {
		result := models.TaskResult{
			Code:   200,
			Task:   taskExecute.Task,
			Output: make([]byte, 0),
		}
		//接受到数据开始处理
		result.StartTime = time.Now()
		// 抢锁
		etcdMutex := &EtcdMutex{ttl: 10}

		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

		err := etcdMutex.TryLock(config.TASK_LOCK_DIR + taskExecute.Task.Title)
		defer etcdMutex.UnLock()
		if err != nil {
			result.Code = 501
			// 没抢到锁，不执行任务
			result.Err = err
			result.EndTime = time.Now()
		} else {

			result.StartTime = time.Now()
			// 执行Task
			cmd := exec.CommandContext(taskExecute.CancelCtx, "/bin/bash", "-c", taskExecute.Task.Command)

			output, err := cmd.CombinedOutput()
			if err != nil {
				result.Code = 500
			}
			result.Output = output
			result.Err = err
			result.EndTime = time.Now()

		}
		GDis.receiveResult(&result)
	}()
}
