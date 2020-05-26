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
		logger.WithFields(logrus.Fields{
			"taskId": task.Id,
			"err":    err,
		}).Warningln("update run_status failed")
	}

	taskResult := TaskResult{Task: task}

	// 开始时间
	taskResult.StartTime = time.Now().UnixNano() / 1e6

	// linux下命令行执行命令
	cmd := exec.Command("sh", "-c", task.Command)
	var opBytes []byte
	opBytes, err = cmd.Output()

	// 结束时间
	taskResult.EndTime = time.Now().UnixNano() / 1e6

	// 执行结果
	if err != nil {
		logger.WithFields(logrus.Fields{
			"command": task.Command,
			"err":     err,
		}).Warningf("task[%d] worker failed", task.Id)
		taskResult.RunCode = 500
		taskResult.Result = err.Error()
	} else {
		// 更新执行状态为 等待执行
		_, err = updateRunStatus(task.Id, map[string]interface{}{"t_run_status": "1"})
		if err != nil {
			logger.WithFields(logrus.Fields{
				"taskId": task.Id,
				"err":    err,
			}).Warningln("update run_status failed")
		}
		logger.WithFields(logrus.Fields{
			"command": task.Command,
			"err":     err,
		}).Infoln("worker failed")
		taskResult.RunCode = 200
		taskResult.Result = string(opBytes)
	}
	return &taskResult
}
