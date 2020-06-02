package main

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"gocTask/alarm"
	"gocTask/config"
	"gocTask/models"
	"time"
)

type Notify struct {
}

var GNotify *Notify

func InitNotify() {
	GNotify = &Notify{}
}

func (n *Notify) Run(result *models.TaskResult) {
	go func() {
		// 解析执行任务的返回结果
		rs := struct {
			Code int `json:"code"`
		}{}
		if result.Err == nil {
			// 任务运行结果返回，如果返回code不为200，则更改结果状态。
			err := json.Unmarshal(result.Output, &rs)
			if err != nil {
				GLog.WithFields(logrus.Fields{
					"result":    string(result.Output),
					"taskTitle": result.Task.Title,
					"err":       err,
				}).Warningln("任务返回的结果json解析失败")
				result.Code = 500
				result.Err = err
			} else if rs.Code != 200 {
				result.Code = rs.Code
				result.Err = errors.New(string(result.Output))
			}
		}

		// 保存执行记录到数据库
		err := n.resultSave(result)
		if err != nil {
			GLog.WithFields(logrus.Fields{
				"taskTitle": result.Task.Title,
				"err":       result.Err,
			}).Warningln("任务运行记录log保存失败")
		}

		// 判断是否需要通知，执行错误，并且不是抢占锁错误。
		if result.Code != 200 && result.Code != 501 && result.Task.IsAlarm == "1" {
			n.resultAlarm(result)
		}

	}()
}

// resultSave 保存任务执行结果记录
func (n *Notify) resultSave(result *models.TaskResult) error {
	GLog.WithFields(logrus.Fields{
		"taskTitle": result.Task.Title,
	}).Infoln("save log")
	return nil
}

// resultAlarm 发送Alarm信息
func (n *Notify) resultAlarm(result *models.TaskResult) {
	// 读取今天已经报警的次数
	day := time.Now().Format("20060102")
	numKey := config.TASK_ALARM_NUM_DIR + day + "/" + result.Task.Title

	alarmNum, err := GEtcd.GetAlarmNum(numKey)
	if err != nil {
		GLog.WithFields(logrus.Fields{
			"err":    err,
			"numKey": numKey,
		}).Warningf("读取Alarm次数错误")
	}
	if err == nil && alarmNum < config.GConfig.AlarmMaxNum {
		// 如果包含报错，提交alarm警告
		if len(config.GConfig.AlarmType) > 0 {
			for _, t := range config.GConfig.AlarmType {
				var err error
				if t == "qyweixin" {
					// 微信企业号通知
					wx := alarm.NewAlarm("qyweixin")
					err = wx.SendMessage(result.Task.Uid, result.Err.Error())
					if err != nil {
						GLog.WithFields(logrus.Fields{
							"alarmType": t,
							"err":       err,
							"taskTitle": result.Task.Title,
						}).Warningf("Alarm消息发送失败")
					}
				} else if t == "email" && result.Task.NotifyEmail != "" {
					nf := alarm.NewAlarm("email")
					err = nf.SendMessage(result.Task.NotifyEmail, result.Err.Error())
					if err != nil {
						GLog.WithFields(logrus.Fields{
							"alarmType": t,
							"err":       err,
							"taskTitle": result.Task.Title,
						}).Warningf("Alarm消息发送失败")
					}
				}
			}

			// 保存警告次数
			err := GEtcd.SetAlarmNum(numKey, alarmNum+1)
			if err != nil {
				GLog.WithFields(logrus.Fields{
					"err":    err,
					"numKey": numKey,
				}).Warningf("保存Alarm次数错误")
			}
		}
	}
}
