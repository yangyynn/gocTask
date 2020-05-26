package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

// 任务执行结果，保存结果，通知结果
type TaskResult struct {
	RunCode   int
	StartTime int64
	EndTime   int64
	Result    string //json格式的返回{"code":"200","message":""}成功 {"code":"99","message":""}错误
	Task      *Task
}

// resultProcess 任务执行完结果处理，保存任务执行log
// 出错则通知任务负责人（微信，邮件等方式）
func resultProcess(result *TaskResult) {
	// 解析执行任务的返回结果
	rs := struct {
		code    int
		message string
	}{}
	if result.RunCode == 200 {
		// 任务运行返回不是200运行成功，则显示更改结果状态。
		err := json.Unmarshal([]byte(result.Result), &rs)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"result": result.Result,
				"err":    err,
			}).Warningln("task worker result json decode failed")
			result.RunCode = 500
		}
	}

	rs.code = result.RunCode
	rs.message = result.Result

	// 保存执行记录到log表

	// 判断是否需要通知
	if rs.code != 200 {
		notify(rs.message, result.Task)
	}
}

// notify 任务执行出错通知
func notify(message string, task *Task) {

}
