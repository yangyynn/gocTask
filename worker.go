package main

import (
	"github.com/sirupsen/logrus"
	"os/exec"
	"time"
)

func workerProcess(w chan *Task) {
	for {
		task := <-w
		result := working(task)
		go resultProcess(result)
	}
}

// working 执行任务
func working(task *Task) *TaskResult {
	// 更新执行状态为 执行中
	_, err := updateRunStatus(task.Id, map[string]interface{}{"t_run_status": "2"})
	if err != nil {
		Logger.WithFields(logrus.Fields{
			"taskId": task.Id,
			"err":    err,
		}).Warningln("更新数据库任务运行状态失败")
	}

	taskResult := TaskResult{Task: task}

	// 开始时间
	now := time.Now()
	Logger.WithFields(logrus.Fields{
		"taskId": task.Id,
	}).Infoln("任务开始执行")

	// linux下命令行执行命令
	cmd := exec.Command("sh", "-c", task.Command)
	var opBytes []byte
	opBytes, err = cmd.Output()

	Logger.WithFields(logrus.Fields{
		"taskId": task.Id,
	}).Infoln("任务执行结束")
	// 执行用时
	taskResult.UseTime = time.Since(now).Truncate(time.Millisecond).Seconds()

	// 执行结果
	if err != nil {
		Logger.WithFields(logrus.Fields{
			"command": task.Command,
			"err":     err,
		}).Warningf("任务ID:%d 运行失败", task.Id)
		taskResult.RunCode = 500
		taskResult.Result = err.Error()
	} else {
		taskResult.RunCode = 200
		taskResult.Result = string(opBytes)
	}
	// 更新执行状态为 等待执行
	_, err = updateRunStatus(task.Id, map[string]interface{}{"t_run_status": "1"})
	if err != nil {
		Logger.WithFields(logrus.Fields{
			"taskId": task.Id,
			"err":    err,
		}).Warningln("更新数据库任务运行状态失败")
	}

	return &taskResult
}
