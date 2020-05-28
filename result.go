package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"goctask/notify"
	"strconv"
)

// 任务执行结果，保存结果，通知结果
type TaskResult struct {
	RunCode   int
	UseTime   float64
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
			}).Warningln("任务返回的结果json解析失败")
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
	err := saveTaskLog(result)
	if err != nil {
		Logger.WithFields(logrus.Fields{
			"taskId": result.Task.Id,
			"err":    err,
		}).Warningln("任务运行记录log保存失败")
	}

	// 判断是否需要通知
	if result.RunCode != 200 {
		sendNotify(result)
	}
}

// sendNotify 任务执行出错通知 微信企业号/Email
func sendNotify(r *TaskResult) {

	// 获取今天已经报警次数
	num, _ := getNotifyNum(r.Task.Id)

	if num < r.Task.NotifyNum {

		var err error

		if NotifyWx {
			// 微信企业号通知
			nf := notify.NewNotify("qyweixin")
			err := nf.SendMessage(r.Task.NotifyId, r.Result)
			if err != nil {
				Logger.WithFields(logrus.Fields{
					"err": err,
					"taskId": r.Task.Id,
				}).Warningf("企业微信消息发送失败")
			} else {
				Logger.WithFields(logrus.Fields{
					"message": r.Result,
					"taskId": r.Task.Id,
				}).Infof("企业微信消息发送成功")
			}
		}

		if NotifyEmail {
			nf := notify.NewNotify("email")
			err := nf.SendMessage(r.Task.NotifyEmail, r.Result)
			if err != nil {
				Logger.WithFields(logrus.Fields{
					"err": err,
					"taskId": r.Task.Id,
				}).Warningf("Email消息发送失败")
			} else {
				Logger.WithFields(logrus.Fields{
					"message": r.Result,
					"taskId": r.Task.Id,
				}).Infof("Email消息发送成功")
			}
		}

		// 记录报警
		err = saveNotify(r.Task.Id)
		if err != nil {
			Logger.WithFields(logrus.Fields{
				"err":err,
				"taskId":r.Task.Id,
			}).Warningf("保存任务失败通知信息到数据库失败")
		}

		Logger.WithFields(logrus.Fields{
			"message": r.Task.Title,
			"taskId":r.Task.Id,
		}).Infof("任务失败通知保存成功")

	}

}
