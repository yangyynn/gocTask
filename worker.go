package main

import (
	"fmt"
	"gocTask/models"
	"log"
	"time"
)

type Worker struct {
	dis *Dispatcher
}

func CreateConcurrentWorker(dis *Dispatcher) *Worker {
	return &Worker{
		dis: dis,
	}
}

func (w *Worker) Run() {
	in := make(chan *models.Task)
	go func() {
		for {
			//worker channel准备好了
			w.dis.WorkerReady(in)
			//接受到数据开始处理
			task := <-in
			result, err := w.doing(task)
			if err != nil {
				continue
			}

			w.notify(result)
		}
	}()
}

// doing 执行任务
func (w *Worker) doing(task *models.Task) (string, error) {
	//通道消耗时间
	start := time.Now()

	time.Sleep(500 * time.Millisecond) // 模拟任务执行

	workUseTime := time.Since(start)

	log.Printf("%s, use time: [%dms]\n", task.Title, workUseTime)
	return fmt.Sprintf("api: %s", task.Title), nil
}

// notify 通知任务执行结果。
// 记录执行结果
// 错误则通知负责人错误
func (w *Worker) notify(result string) {
	fmt.Println(result)
}
