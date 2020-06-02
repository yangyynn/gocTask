package alarm

type Alarm interface {
	SendMessage(string, string) error
}

func NewAlarm(s string) Alarm {
	if s == "qyweixin" {
		return &QyWeiXin{}
	} else if s == "email" {
		return &Email{}
	} else {
		return nil
	}
}
