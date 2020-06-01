package main

import "gocTask/models"

type Notify struct {
	NotifyChan chan *models.TaskResultNotify
}

var GNotify *Notify

func InitNotify() {
	GNotify = &Notify{}
}
