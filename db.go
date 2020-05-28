package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"hash/crc32"
	"strconv"
	"strings"
	"time"
)

var db *sql.DB

func dbConn() {
	// Logger.Infoln("conn mysql...")
	var err error
	db, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/goctask?charset=utf8")
	if err != nil {
		logrus.WithField("err", err.Error()).Fatalln("数据库open错误")
	}
	// 设置连接池
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.Ping()
}

func dbClose()  {
	db.Close()
}

// getTasksInMysql 读取mysql中的task
func getTasksInMysql(d chan *Task) {
	dbConn()

	Logger.Infoln("读取数据库中需要运行的任务")

	//查询task数据，返回sql.Rows结果集
	sqlStr := "select id,t_title,t_crontab,t_content,t_start_time,t_end_time,notify_id,notify_num,notify_email from goc_task where t_status = ?"
	rows, err := db.Query(sqlStr, 1)

	defer dbClose()
	defer rows.Close()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
			"sql": sqlStr,
		}).Fatalln("读取任务sql失败")
	}

	var startTime, endTime string
	for rows.Next() {
		task := Task{}
		rows.Scan(&task.Id, &task.Title, &task.Crontab, &task.Command, &startTime, &endTime, &task.NotifyId, &task.NotifyNum, &task.NotifyEmail)

		t, _ := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local)
		task.StartTime = t.Unix()
		t, _ = time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
		task.EndTime = t.Unix()

		d <- &task
	}
}

// updateRunStatus 更新运行时状态
func updateRunStatus(id int, kv map[string]interface{}) (int64, error) {
	updateStr := make([]string, 0)
	for k, v := range kv {
		updateStr = append(updateStr, fmt.Sprintf("%s = '%s'", k, v))
	}
	dbConn()
	defer dbClose()
	sql := fmt.Sprintf("update goc_task set %s where id = ?", strings.Join(updateStr, "and"))
	ret, err := db.Exec(sql, id)
	if err != nil {
		return 0, err
	}
	//获取影响⾏数
	return ret.RowsAffected()
}

// 获取当天报警次数
func getNotifyNum(id int) (int, error){
	curTime:=time.Now()  //获取系统当前时间
	dh, _ := time.ParseDuration("+24h")
	nextStr := curTime.Add(dh).Format("2006-01-02") //后一天日期
	nowStr := time.Now().Format("2006-01-02")

	dbConn()
	defer dbClose()

	num := 0
	db.QueryRow("select (max(id)-min(id)+1) as total from goc_notify where t_id = ? and created_at between ? and ?", id, nowStr, nextStr).Scan(&num)
	return num, nil
}

// saveNotify 保存通知
func saveNotify(id int) error  {
	dbConn()
	defer dbClose()

	insertNotify, _ := db.Prepare("insert into goc_notify(t_id,created_at) values(?,?)")
	defer insertNotify.Close()

	curTime:=time.Now().Format("2006-01-02 15:04:05")
	_, err := insertNotify.Exec(id, curTime)
	if err != nil {
		return err
	}
	return nil
}

// saveTaskLog 保存执行结果进mysql
func saveTaskLog(r *TaskResult) error {
	tableId := getHashTableId(r.Task.Title, 2)
	tableName := "goc_task_log_"+tableId

	dbConn()
	defer dbClose()

	insertP, _ := db.Prepare("insert into "+tableName+" (t_id,l_status,l_result,l_use_time,created_at) values (?,?,?,?,?)")
	defer insertP.Close()

	useTime := r.UseTime
	_, err := insertP.Exec(r.Task.Id, r.RunCode, r.Result, useTime, time.Now().Unix())
	if err != nil {
		return err
	}
	return nil
}

// getTableId 分表函数，根据传入的参数,分表个数返回表名
func getHashTableId(arg string, size uint32) string {
	hashValue := crc32.ChecksumIEEE([]byte(arg))
	tableId := hashValue % size
	return strconv.Itoa(int(tableId))
}

