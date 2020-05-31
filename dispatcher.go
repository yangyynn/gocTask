package main

import (
	"github.com/gorhill/cronexpr"
	"gocTask/config"
	"gocTask/models"
	"time"
)

var GDis *Dispatcher

// Dispatcher 调度task和worker
type Dispatcher struct {
	// 接收task通道
	TaskEventChan chan *models.TaskEvent
	// 任务执行表
	TaskPlanMap map[string]*models.TaskPlan
	// 任务执行表
	TaskExecuteMap map[string]*models.TaskExecute
	// 执行结果通道
	TaskResultChan chan *models.TaskResult
}

// InitDispatcher 创建一个dispatcher， taskChanCount任务通道缓冲数
func InitDispatcher() {
	GDis = &Dispatcher{
		TaskEventChan:  make(chan *models.TaskEvent, 100),
		TaskPlanMap:    make(map[string]*models.TaskPlan),
		TaskExecuteMap: make(map[string]*models.TaskExecute),
		TaskResultChan: make(chan *models.TaskResult, 100),
	}
	GDis.getTasks()
	GDis.watchTasks()
	GDis.run()
}

// getTasks 获取Etcd中的所有任务
func (d *Dispatcher) getTasks() {
	tasks, err := GEtcd.ListTask()
	if err != nil {
		GLog.Warnf("get tasks from etcd failed, err: %s", err.Error())
	}

	for _, task := range tasks {
		taskEvent := &models.TaskEvent{
			Event: config.TASK_PUT_EVENT,
			Task:  task,
		}
		d.TaskEventChan <- taskEvent
	}
}

// watchTasks 监听Etcd中的任务
func (d *Dispatcher) watchTasks() {
	GEtcd.WatchTasks()
}

// run 开始运行dispatcher等待taskChan的输入进行处理
func (d *Dispatcher) run() {
	go func() {
		dispatcherTime := d.doDispatcher()
		rateTimer := time.NewTimer(dispatcherTime)
		for {

			select {
			case taskEvent := <-d.TaskEventChan:
				//任务放入队列
				d.doEvent(taskEvent)

			case taskResult := <-d.TaskResultChan:
				d.doResult(taskResult)

			case <-rateTimer.C:
			}
			dispatcherTime = d.doDispatcher()
			rateTimer.Reset(dispatcherTime)
		}
	}()
}

// doEvent
func (d *Dispatcher) doEvent(event *models.TaskEvent) {
	switch event.Event {
	case config.TASK_PUT_EVENT:
		expr, _ := cronexpr.Parse(event.Task.Crontab)
		taskPlan := &models.TaskPlan{
			Task:     event.Task,
			Expr:     expr,
			NextTime: expr.Next(time.Now()),
		}
		// 加入任务执行表
		d.TaskPlanMap[event.Task.Title] = taskPlan
	case config.TASK_DEL_EVENT:
		// 删除task
		delete(d.TaskPlanMap, event.Task.Title)
	}
}

// doDispatcher
func (d *Dispatcher) doDispatcher() (sleepTime time.Duration) {
	var (
		now        = time.Now()
		lastDoTime *time.Time
	)

	if len(d.TaskPlanMap) == 0 {
		return 1 * time.Second
	}

	for _, plan := range d.TaskPlanMap {
		//如果还未到任务开始时间，则跳过任务
		if plan.Task.StartTime != 0 && now.Unix() < plan.Task.StartTime {
			continue
		}

		//如果现在时间超过endTime，task过期，则删除task。
		if plan.Task.EndTime != 0 && now.Unix() > plan.Task.EndTime {
			delete(d.TaskPlanMap, plan.Task.Title)
			continue
		}

		if plan.NextTime.Before(now) || plan.NextTime.Equal(now) {
			d.doWork(plan)
			// 重置下次执行时间
			plan.NextTime = plan.Expr.Next(now)
		}

		if lastDoTime == nil || plan.NextTime.Before(*lastDoTime) {
			lastDoTime = &plan.NextTime
		}
	}
	return (*lastDoTime).Sub(now)
}

// doWork 执行任务
func (d *Dispatcher) doWork(plan *models.TaskPlan) {
	_, isExecute := d.TaskExecuteMap[plan.Task.Title]
	if isExecute {
		GLog.Infoln("执行中……跳过，task is", plan.Task.Title)
		return
	}

	d.TaskExecuteMap[plan.Task.Title] = &models.TaskExecute{
		Task:     plan.Task,
		PlanTime: plan.NextTime,
		Realtime: time.Now(),
	}

	GLog.Infof("执行任务: %s, PlanTime: %s, RealTime %s ", plan.Task.Title, plan.NextTime, time.Now())
	GWorker.Run(d.TaskExecuteMap[plan.Task.Title])
}

// doResult 处理work结果
func (d *Dispatcher) doResult(result *models.TaskResult) {
	delete(d.TaskExecuteMap, result.Task.Title)
	GLog.Infof("执行任务结果: %s, err: %s, UseTime %s", result.Output, result.Err.Error(), result.UseTime)

	//todo notify(result)
}

// receiveResult 接受work结果
func (d *Dispatcher) receiveResult(result *models.TaskResult) {
	d.TaskResultChan <- result
}
