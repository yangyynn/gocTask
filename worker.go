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
		//接受到数据开始处理
		//消耗时间
		start := time.Now()
		cmd := exec.CommandContext(context.TODO(), "/bin/bash", "-c", taskExecute.Task.Command)

		output, err := cmd.CombinedOutput()

		result := models.TaskResult{
			Task:    taskExecute.Task,
			Output:  output,
			Err:     err,
			UseTime: time.Since(start),
		}

		GDis.receiveResult(&result)
	}()
}
