package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	// 企业号的标识
	CORPID = "wxab0457202f91c2fc"
	// 管理组凭证密钥
	CORPSECRET = "v9DbPwXIpmu5RYcMa3xqzxR51VyEzIHJYhpUVrpNxPk"
	// 应用id
	AGENTID = 1000063
)

var curToken = WxToken{}

type QyWeiXin struct {
	touser  string
	msgtype string
	agentid string
	text    map[string]string
	safe    string
}

type WxToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	expiresTime int64
}

func (q *QyWeiXin) SendMessage(userId string, content string) error {
	token, err := q.getToken()
	if err != nil {
		return err
	}

	data := QyWeiXin{
		userId,
		"text",
		strconv.Itoa(AGENTID),
		map[string]string{"content": content},
		"0",
	}

	var postData []byte
	postData, err = json.Marshal(data)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)

	resp, err1 := http.Post(addr, "application/x-www-form-urlencoded", bytes.NewBuffer(postData))
	if err1 != nil {
		return err1
	}

	//todo 处理微信发送报错

	defer resp.Body.Close()
	return nil
}

func (*QyWeiXin) getToken() (string, error) {

	if curToken.expiresTime > time.Now().Unix() {
		return curToken.AccessToken, nil
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", CORPID, CORPSECRET)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenBytes []byte
	tokenBytes, err = ioutil.ReadAll(resp.Body)

	token := WxToken{}
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		return "", nil
	}

	token.expiresTime = time.Now().Unix() + int64(token.ExpiresIn)

	return token.AccessToken, nil
}
