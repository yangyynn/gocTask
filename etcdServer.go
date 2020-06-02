package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"gocTask/config"
	"gocTask/models"
	"strconv"
	"strings"
	"time"
)

// EtcdServer
type EtcdServer struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
	watch  clientv3.Watcher
}

var GEtcd *EtcdServer

// InitEtcd
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

// AddTask
func (e *EtcdServer) AddTask(t *models.Task) (oldTask *models.Task, err error) {
	var (
		key     string
		value   []byte
		putResp *clientv3.PutResponse
	)

	key = config.TASK_TASK_DIR + t.Title

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

// DeleteTask
func (e *EtcdServer) DeleteTask(title string) (oldTask *models.Task, err error) {
	var (
		key     string
		delResp *clientv3.DeleteResponse
	)
	key = config.TASK_TASK_DIR + title

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

// ListTask
func (e *EtcdServer) ListTask() ([]*models.Task, error) {
	getResp, err := GEtcd.kv.Get(context.TODO(), config.TASK_TASK_DIR, clientv3.WithPrefix())
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

// TaskAlarmNum 获取任务警告次数
func (e *EtcdServer) GetAlarmNum(key string) (num int, err error) {
	var numResp *clientv3.GetResponse
	if numResp, err = GEtcd.kv.Get(context.TODO(), key); err != nil {
		return
	}
	if len(numResp.Kvs) > 0 {
		return strconv.Atoi(string(numResp.Kvs[0].Value))
	} else {
		return 0, nil
	}
}

// TaskAlarmNum 保存警告次数
func (e *EtcdServer) SetAlarmNum(key string, num int) (err error) {
	_, err = GEtcd.kv.Put(context.TODO(), key, strconv.Itoa(num))
	return
}

// KillTask 关闭执行中的任务
func (e *EtcdServer) KillTask(title string) error {
	key := config.TASK_KILL_DIR + title

	GLog.Infof("etcd add kill task, title is ", title)

	_, err := e.kv.Put(context.TODO(), key, title)
	if err != nil {
		return err
	}

	return nil
}

// watchTasks 监听etcd中的所有任务
func (e *EtcdServer) WatchTasks() error {
	getResponse, err := e.kv.Get(context.TODO(), config.TASK_TASK_DIR, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	GLog.Infoln("watching task")
	go func() {
		startRevision := getResponse.Header.Revision + 1
		watchChan := e.watch.Watch(context.TODO(), config.TASK_TASK_DIR, clientv3.WithPrefix(), clientv3.WithRev(startRevision))
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
					taskTitle := strings.TrimLeft(string(event.Kv.Key), config.TASK_TASK_DIR)
					taskEvent.Task = &models.Task{Title: taskTitle}
				}
				GDis.TaskEventChan <- taskEvent
			}
		}
	}()

	return nil
}

func (e *EtcdServer) WatchKillTask() error {
	go func() {
		watchChan := e.watch.Watch(context.TODO(), config.TASK_KILL_DIR, clientv3.WithPrefix())
		for wResp := range watchChan {
			for _, event := range wResp.Events {
				taskEvent := &models.TaskEvent{}
				switch event.Type {
				case mvccpb.PUT:
					taskName := strings.TrimPrefix(string(event.Kv.Value), config.TASK_KILL_DIR)
					taskEvent.Event = config.TASK_KILL_EVENT
					taskEvent.Task.Title = taskName
					GDis.TaskEventChan <- taskEvent
				case mvccpb.DELETE:
				}
			}
		}
	}()

	return nil
}

// EtcdMutex
type EtcdMutex struct {
	key        string
	ttl        int64
	ctx        context.Context
	cancelFunc context.CancelFunc
	leaseId    clientv3.LeaseID
	IsLocked   bool // 是否上锁成功
}

// initMutex 初始化分布式锁
func (em *EtcdMutex) TryLock(key string) (err error) {
	var (
		leaseResp     *clientv3.LeaseGrantResponse
		leaseRespChan <-chan *clientv3.LeaseKeepAliveResponse
		txn           clientv3.Txn
		txnRes        *clientv3.TxnResponse
	)

	// 创建上下文
	em.ctx, em.cancelFunc = context.WithCancel(context.TODO())
	// 创建续租
	leaseResp, err = GEtcd.lease.Grant(em.ctx, 5)
	// 续租id
	em.leaseId = leaseResp.ID
	// 保持续租
	if leaseRespChan, err = GEtcd.lease.KeepAlive(em.ctx, em.leaseId); err != nil {
		goto FAIL
	}
	// 每隔1秒检查租约
	go func() {
		var keepRes *clientv3.LeaseKeepAliveResponse
		for {
			select {
			case keepRes = <-leaseRespChan:
				// 如果续约失败
				if keepRes == nil {
					goto END
				}
			}
			time.Sleep(1 * time.Second)
		}
	END:
	}()

	// 创建事务txn
	txn = GEtcd.kv.Txn(context.TODO())
	// 锁路径
	em.key = key
	// 事务枪锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(em.key), "=", 0)).
		Then(clientv3.OpPut(em.key, "", clientv3.WithLease(em.leaseId))).
		Else(clientv3.OpGet(em.key))

	// 提交事务
	if txnRes, err = txn.Commit(); err != nil {
		// 提交创建锁失败
		goto FAIL

	}
	// 如果抢锁失败
	if !txnRes.Succeeded {
		// 锁被占用
		err = errors.New("锁被占用")
		goto FAIL
	}

	em.IsLocked = true

	return

FAIL:
	// 释放上下文,取消续约
	em.cancelFunc()
	GEtcd.lease.Revoke(context.TODO(), em.leaseId) // 释放租约
	return err
}

// UnLock 释放锁
func (em *EtcdMutex) UnLock() {
	if em.IsLocked {
		em.cancelFunc()                                // 取消自动续租协程
		GEtcd.lease.Revoke(context.TODO(), em.leaseId) // 释放租约
	}
}
