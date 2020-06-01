package main

import (
	"encoding/json"
	"gocTask/config"
	"gocTask/models"
	"net"
	"net/http"
	"time"
)

type ApiServer struct {
	httpServer *http.Server
}

var GApiServer *ApiServer

func InitServer() error {

	mux := http.NewServeMux()
	mux.HandleFunc("/task/add", TaskAdd)
	mux.HandleFunc("/task/delete", TaskDelete)
	mux.HandleFunc("/task/list", TaskList)
	mux.HandleFunc("/task/kill", TaskKill)

	listen, err := net.Listen("tcp", config.GConfig.ApiServerAddr)
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		ReadTimeout:  time.Duration(config.GConfig.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(config.GConfig.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	GApiServer = &ApiServer{
		httpServer: httpServer,
	}

	go httpServer.Serve(listen)

	return nil
}

// TaskAdd 添加任务。将计划任务添加到etcd
func TaskAdd(w http.ResponseWriter, r *http.Request) {
	var (
		task     models.Task
		taskJson string
		err      error
		oldTask  *models.Task
		encodeW  []byte
	)
	if err = r.ParseForm(); err != nil {
		goto ERR
	}
	taskJson = r.PostForm.Get("task")
	GLog.Infof("add task is %s", taskJson)
	if err = json.Unmarshal([]byte(taskJson), &task); err != nil {
		goto ERR
	}
	oldTask, err = GEtcd.AddTask(&task)
	if err != nil {
		goto ERR
	}
	encodeW, err = models.EncodeResponse(0, "success", oldTask)
	if err == nil {
		w.Write(encodeW)
	}
	return
ERR:
	encodeW, err = models.EncodeResponse(0, err.Error(), nil)
	if err == nil {
		w.Write(encodeW)
	}
}

// TaskDelete 删除任务。将任务从etcd中删除
func TaskDelete(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		title   string
		encodeW []byte
		oldTask *models.Task
	)
	if err = r.ParseForm(); err != nil {
		goto ERR
	}
	title = r.PostForm.Get("title")
	oldTask, err = GEtcd.DeleteTask(title)
	if err != nil {
		goto ERR
	}
	encodeW, err = models.EncodeResponse(0, "success", oldTask)
	if err == nil {
		w.Write(encodeW)
	}
	return
ERR:
	encodeW, err = models.EncodeResponse(0, err.Error(), nil)
	if err == nil {
		w.Write(encodeW)
	}
}

// TaskList 读取etcd中所有的任务
func TaskList(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		tasks   []*models.Task
		encodeW []byte
	)
	tasks, err = GEtcd.ListTask()
	if err != nil {
		goto ERR
	}

	encodeW, _ = models.EncodeResponse(0, "success", tasks)
	if err == nil {
		w.Write(encodeW)
	}
	return
ERR:
	encodeW, _ = models.EncodeResponse(0, err.Error(), nil)
	if err == nil {
		w.Write(encodeW)
	}
}

// TaskKill 强杀执行中的任务
func TaskKill(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		title   string
		encodeW []byte
	)
	if err = r.ParseForm(); err != nil {
		goto ERR
	}
	title = r.PostForm.Get("title")
	err = GEtcd.KillTask(title)
	if err != nil {
		goto ERR
	}

	encodeW, _ = models.EncodeResponse(0, "success", title)
	if err == nil {
		w.Write(encodeW)
	}
	return
ERR:
	encodeW, _ = models.EncodeResponse(0, err.Error(), nil)
	if err == nil {
		w.Write(encodeW)
	}
}
