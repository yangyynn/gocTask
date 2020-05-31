package main

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"gocTask/config"
	"gocTask/models"
	"strings"
	"time"
)

type EtcdServer struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
	watch  clientv3.Watcher
}

var GEtcd *EtcdServer

func InitEtcd() (err error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.GConfig.EtcdServer,
		DialTimeout: time.Duration(config.GConfig.EtcdDialTimeout) * time.Millisecond,
	})
	if err != nil {
		return
	}

	GEtcd = &EtcdServer{
		client: cli,
		kv:     clientv3.NewKV(cli),
		lease:  clientv3.NewLease(cli),
		watch:  clientv3.NewWatcher(cli),
	}

	return
}

func (e *EtcdServer) AddTask(t *models.Task) (oldTask *models.Task, err error) {
	var (
		key     string
		value   []byte
		putResp *clientv3.PutResponse
	)

	key = config.TASK_PREFIX + t.Title

	if value, err = json.Marshal(t); err != nil {
		return
	}

	GLog.Infof("etcd add task, task key is %s value is %s", key, string(value))

	putResp, err = e.kv.Put(context.TODO(), key, string(value), clientv3.WithPrevKV())
	if err != nil {
		return
	}

	if putResp.PrevKv != nil {
		err = json.Unmarshal(putResp.PrevKv.Value, &oldTask)
		if err != nil {
			return nil, nil
		}
	}

	return
}

func (e *EtcdServer) DeleteTask(title string) (oldTask *models.Task, err error) {
	var (
		key     string
		delResp *clientv3.DeleteResponse
	)
	key = config.TASK_PREFIX + title

	GLog.Infof("etcd delete task, task key is %s", key)

	delResp, err = GEtcd.kv.Delete(context.TODO(), key, clientv3.WithPrevKV())
	if err != nil {
		return
	}

	if len(delResp.PrevKvs) != 0 {
		err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldTask)
		if err != nil {
			return nil, nil
		}
	}

	return oldTask, nil
}

func (e *EtcdServer) ListTask() ([]*models.Task, error) {
	getResp, err := GEtcd.kv.Get(context.TODO(), config.TASK_PREFIX, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	tasks := make([]*models.Task, 0)

	for _, kv := range getResp.Kvs {
		t := &models.Task{}
		err := json.Unmarshal(kv.Value, t)
		if err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// watchTasks 监听etcd中的所有任务
func (e *EtcdServer) WatchTasks() error {
	getResponse, err := e.kv.Get(context.TODO(), config.TASK_PREFIX, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	GLog.Infoln("watching task")
	go func() {
		startRevision := getResponse.Header.Revision + 1
		watchChan := e.watch.Watch(context.TODO(), config.TASK_PREFIX, clientv3.WithPrefix(), clientv3.WithRev(startRevision))
		for wResp := range watchChan {
			for _, event := range wResp.Events {
				taskEvent := &models.TaskEvent{}
				switch event.Type {
				case mvccpb.PUT:
					taskEvent.Event = config.TASK_PUT_EVENT
					err = json.Unmarshal(event.Kv.Value, &taskEvent.Task)
					if err != nil {
						continue
					}
				case mvccpb.DELETE:
					taskTitle := strings.TrimLeft(string(event.Kv.Key), config.TASK_PREFIX)
					taskEvent.Task = &models.Task{Title: taskTitle}
				}
				GDis.TaskEventChan <- taskEvent
			}
		}
	}()

	return nil
}
