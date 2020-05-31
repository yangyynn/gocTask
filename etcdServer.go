package main

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"gocTask/config"
	"gocTask/models"
	"time"
)

type EtcdServer struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
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
	}

	return
}

func (e *EtcdServer) AddTask(t *models.Task) (oldTask *models.Task, err error) {
	var (
		key     string
		value   []byte
		putResp *clientv3.PutResponse
	)

	key = config.TASKPREFIX + t.Title

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
	key = config.TASKPREFIX + title

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
	key := config.TASKPREFIX

	getResp, err := GEtcd.kv.Get(context.TODO(), key, clientv3.WithPrefix())
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
