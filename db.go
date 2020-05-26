package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"strings"
)

var db *sql.DB

func dbConn() {
	logger.Infoln("conn mysql...")
	var err error
	db, err = sql.Open("mysql", "root:123456@tcp(192.168.56.10:3306)/goctask?charset=utf8")
	if err != nil {
		logrus.WithField("err", err.Error()).Fatalln("conn database failed")
	}
}

// getTasksInMysql 读取mysql中的task
func getTasksInMysql(d chan *Task) {
	dbConn()

	logger.Infoln("Read task data start...")

	//查询task数据，返回sql.Rows结果集
	sqlStr := "select id,t_title,t_crontab,t_content,t_start_time,t_end_time,c_id from task where t_status = ?"
	rows, err := db.Query(sqlStr, 1)

	//关闭数据库
	defer db.Close()
	defer rows.Close()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
			"sql": sqlStr,
		}).Fatalln("query sql failed")
	}

	for rows.Next() {
		task := Task{}
		rows.Scan(&task.Id, &task.Title, &task.Crontab, &task.Command, &task.StartTime, &task.EndTime, &task.Uid)

		d <- &task
	}
}

// updateRunStatus 更新运行时状态
func updateRunStatus(id int, kv map[string]interface{}) (int64, error) {
	dbConn()
	updateStr := make([]string, 0)
	for k, v := range kv {
		updateStr = append(updateStr, fmt.Sprintf("%s = '%s'", k, v))
	}
	ret, _ := db.Exec("update task set ? where id > ?", strings.Join(updateStr, "and"), id)
	//获取影响⾏数
	return ret.RowsAffected()
}
