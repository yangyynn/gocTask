package alarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gocTask/config"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var curToken = &WxToken{}

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
		strconv.Itoa(config.GConfig.QyweixinAgentid),
		map[string]string{"content": content},
		"0",
	}

	var postData []byte
	postData, err = json.Marshal(data)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)

	var resp *http.Response
	resp, err = http.Post(addr, "application/x-www-form-urlencoded", bytes.NewBuffer(postData))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

func (*QyWeiXin) getToken() (string, error) {

	if curToken.expiresTime > time.Now().Unix() {
		return curToken.AccessToken, nil
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", config.GConfig.QyweixinCorpId, config.GConfig.QyweixinCorpSecret)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenBytes []byte
	tokenBytes, err = ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(tokenBytes, &curToken)
	if err != nil {
		return "", nil
	}

	curToken.expiresTime = time.Now().Unix() + int64(curToken.ExpiresIn)

	return curToken.AccessToken, nil
}
