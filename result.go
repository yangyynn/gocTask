package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	qyWeixin "goctask/notify"
	"strconv"
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
// 出错则通知任务负责人
func resultProcess(result *TaskResult) {
	// 解析执行任务的返回结果
	rs := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{}
	if result.RunCode == 200 {
		// 任务运行结果返回，如果返回code不为200，则显示更改结果状态。
		err := json.Unmarshal([]byte(result.Result), &rs)
		if err != nil {
			Logger.WithFields(logrus.Fields{
				"result": result.Result,
				"err":    err,
			}).Warningln("task worker result json decode failed")
			result.RunCode = 500
			result.Result = fmt.Sprintf("result:%s, err:%s", result.Result, err.Error())
		} else {
			code, _ := strconv.Atoi(rs.Code)
			if code != 200 {
				result.RunCode = code
				result.Result = rs.Message
			}
		}
	}
	// 保存执行记录到log表
	saveLog(result)

	// 判断是否需要通知
	if result.RunCode != 200 {
		notify(result)
	}
}

// notify 任务执行出错通知 微信企业号
func notify(r *TaskResult) {

	// 获取今天已经报警次数
	num, _ := getNotifyNum(r.Task.Id)
	fmt.Printf("num %d maxnum %d\n", num, r.Task.NotifyNum)

	if num < r.Task.NotifyNum {
		// 微信企业号通知
		err := qyWeixin.SendMessage(r.Task.Uid, r.Result)
		if err != nil {
			Logger.WithFields(logrus.Fields{
				"err": err,
			}).Warningf("qyweixin notify failed")
		} else {
			Logger.WithFields(logrus.Fields{
				"message": r.Result,
			}).Infof("qyweixin notify")
		}

		//todo 邮件通知

		// 记录报警
		err = saveNotify(r.Task.Id, r.Task.Uid)
		if err != nil {
			Logger.WithFields(logrus.Fields{
				"err":err,
			}).Infof("save task notify failed")
		}

	}

}
