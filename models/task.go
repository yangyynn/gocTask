package models

import (
	"github.com/gorhill/cronexpr"
	"time"
)

// Task 任务结构，title和command必须填写
type Task struct {
	Title     string //任务名称
	Crontab   string //linux 模式crontab，精确到秒，第一位是秒，如：*/4 * * * * * 为每四秒执行一次
	Command   string //shell命令 或者 url地址
	StartTime int64  //任务开始时间
	EndTime   int64  //任务结束时间
	Uid       int    //负责人ID，用于报警通知
}

// TaskEvent 任务事件
type TaskEvent struct {
	Event int
	Task  *Task
}

type TaskPlan struct {
	Task     *Task
	Expr     *cronexpr.Expression
	NextTime time.Time
}

type TaskExecute struct {
	Task     *Task
	PlanTime time.Time
	Realtime time.Time
}

type TaskResult struct {
	Task      *Task
	Output    []byte
	Err       error
	StartTime time.Time
	EndTime   time.Time
}
