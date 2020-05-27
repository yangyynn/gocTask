package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

var db *sql.DB

func dbConn() {
	// Logger.Infoln("conn mysql...")
	var err error
	db, err = sql.Open("mysql", "root:123456@tcp(192.168.56.10:3306)/goctask?charset=utf8")
	if err != nil {
		logrus.WithField("err", err.Error()).Fatalln("conn database failed")
	}
}

// getTasksInMysql 读取mysql中的task
func getTasksInMysql(d chan *Task) {
	dbConn()

	Logger.Infoln("Read task data start...")

	//查询task数据，返回sql.Rows结果集
	sqlStr := "select id,t_title,t_crontab,t_content,t_start_time,t_end_time,c_id,notify_num from task where t_status = ?"
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

	var startTime, endTime string
	for rows.Next() {
		task := Task{}
		rows.Scan(&task.Id, &task.Title, &task.Crontab, &task.Command, &startTime, &endTime, &task.Uid, &task.NotifyNum)

		t, _ := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local)
		task.StartTime = t.Unix()
		t, _ = time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
		task.EndTime = t.Unix()

		d <- &task
	}
}

// updateRunStatus 更新运行时状态
func updateRunStatus(id int, kv map[string]interface{}) (int64, error) {
	dbConn()
	defer db.Close()
	updateStr := make([]string, 0)
	for k, v := range kv {
		updateStr = append(updateStr, fmt.Sprintf("%s = '%s'", k, v))
	}
	sql := fmt.Sprintf("update task set %s where id = ?", strings.Join(updateStr, "and"))
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
	defer db.Close()

	num := 0
	db.QueryRow("select (max(id)-min(id)+1) as total from notify where t_id = ? and created_at between ? and ?", id, nowStr, nextStr).Scan(&num)
	return num, nil
}

// saveNotify 保存通知
func saveNotify(id int, uId string) error  {
	dbConn()
	defer db.Close()

	insertNotify, _ := db.Prepare("insert into notify values(?,?,?)")
	defer insertNotify.Close()

	curTime:=time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(curTime)
	_, err := insertNotify.Exec(&id, &uId, &curTime)
	if err != nil {
		return err
	}
	return nil
}

// saveLog 保存任务执行记录
func saveLog(result *TaskResult) {

}
