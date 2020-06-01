package models

import "encoding/json"

type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

func EncodeResponse(errno int, msg string, data interface{}) ([]byte, error) {
	resp := Response{
		Errno: errno,
		Msg:   msg,
		Data:  data,
	}

	return json.Marshal(resp)
}
