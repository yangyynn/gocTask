package notify

type NotifyObj interface {
	SendMessage(string, string) error
}

func NewNotify(s string) NotifyObj {
	if s == "qyweixin" {
		return &QyWeiXin{}
	}else if s == "email" {
		return &Email{}
	}else{
		return nil
	}
}