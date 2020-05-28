package notify

import (
	"gopkg.in/gomail.v2"
	"strconv"
)

type Email struct {}

func (*Email) SendMessage(mailTo string, body string ) error {
	//定义邮箱服务器连接信息，如果是阿里邮箱 pass填密码，qq邮箱填授权码
	mailConn := map[string]string {
		"user": "zhanghao@xxxx.com",
		"pass": "xxxx",
		"host": "smtp.mxhichina.com",
		"port": "465",
	}

	port, _ := strconv.Atoi(mailConn["port"]) //转换端口类型为int

	m := gomail.NewMessage()
	m.SetHeader("From", "<" + mailConn["user"] + ">")
	m.SetHeader("To", mailTo)
	m.SetHeader("Subject", "计划任务错误notify")
	m.SetBody("text/html", body)

	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])

	err := d.DialAndSend(m)
	return err

}
