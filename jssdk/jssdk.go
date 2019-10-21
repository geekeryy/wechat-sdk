package jssdk

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	ACCESS_TOKEN_URL = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	JSAPI_TICKET_URL = "https://api.weixin.qq.com/cgi-bin/ticket/getticket?type=jsapi&access_token=%s"
)

type Jssdk struct {
	accessToken      *accessToken
	jsspiTicket      *jsapiTicket
	appId, appSecret string
	storage          Storage
}

type accessToken struct {
	Access_token string `json:"access_token"`
	Expires_in   int64  `json:"expires_in"`
	Expire_time  int64  `json:"expire_time"`
}

type jsapiTicket struct {
	Jsapi_ticket string `json:"ticket"`
	Expires_in   int64  `json:"expires_in"`
	Expire_time  int64  `json:"expire_time"`
}

type Storage interface {
	saveAccessToken([]byte) error
	saveJsapiTicket([]byte) error
	getJsapiTicket() (string, error)
	getAccessToken() (string, error)
}

func NewJssdk(appId, appSecret string, s Storage) *Jssdk {
	if s == nil {
		s = &_storage{
			accessTokenPath: "access_token.json",
			jsapiTicketPath: "jsapi_ticket.json",
		}
	}
	w := &Jssdk{
		accessToken: &accessToken{},
		jsspiTicket: &jsapiTicket{},
		appId:       appId,
		appSecret:   appSecret,
		storage:     s,
	}
	return w

}

// sha1加密
func (w *Jssdk) sha1(data []byte) string {
	sha1 := sha1.New()
	sha1.Write(data)
	return hex.EncodeToString(sha1.Sum([]byte(nil)))
}

// 获取签名内容
func (w *Jssdk) GetSignPackage(url string) (string,error) {
	if _,err:=w.GetJsapiTicket();err!=nil{
		return "",err
	}

	timestamp := time.Now().Unix()
	nonceStr := w.createNonceStr(16)

	// 这里参数的顺序要按照 key 值 ASCII 码升序排序
	str := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s",w.jsspiTicket.Jsapi_ticket,nonceStr,timestamp,url)

	signature := w.sha1([]byte(str))

	signPackage :=make(map[string]string)
	signPackage["appId"]=w.appId
	signPackage["nonceStr"]=nonceStr
	signPackage["timestamp"]=fmt.Sprintf("%d",timestamp)
	signPackage["url"]=url
	signPackage["signature"]=signature
	signPackage["rawString"]=str

	bytes, e := json.Marshal(signPackage)
	return string(bytes),e
}

// 创建随机字符串
func (w *Jssdk) createNonceStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// 获取AccessToken
// 通过accessToken!=nil来判断是否存在token
// error仅用作错误信息返回
func (w *Jssdk) GetAccessToken() (*accessToken, error) {

	// 本地读取AccessToken
	s, e := w.storage.getAccessToken()
	if e != nil || s == "" {
		return nil, e
	}
	if err := json.Unmarshal([]byte(s), w.accessToken); err != nil {
		return nil, err
	}

	// AccessToken过期判断
	if w.accessToken.Expire_time > time.Now().Unix() {
		return w.accessToken, nil
	}

	// 向微信服务器请求AccessToken
	url := fmt.Sprintf(ACCESS_TOKEN_URL, w.appId, w.appSecret)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bytes, w.accessToken)
	if err != nil || w.accessToken == nil {
		return nil, err
	}

	// 修改AccessToken过期时间
	w.accessToken.Expire_time = time.Now().Unix() + 7000
	tokenStr, err := json.Marshal(w.accessToken)
	if err != nil {
		return w.accessToken, err
	}

	// 存储AccessToken
	if err := w.storage.saveAccessToken(tokenStr); err != nil {
		return w.accessToken, err
	}

	return w.accessToken, nil

}

func (w *Jssdk) GetJsapiTicket() (*jsapiTicket, error) {

	// 本地读取JsapiTicket
	s, e := w.storage.getJsapiTicket()
	if e != nil || s == "" {
		return nil, e
	}
	if err := json.Unmarshal([]byte(s), w.jsspiTicket); err != nil {
		return nil, err
	}

	// JsapiTicket过期判断
	if w.jsspiTicket.Expire_time > time.Now().Unix() {
		return w.jsspiTicket, nil
	}

	// 获取AccessToken
	if _, e := w.GetAccessToken(); e!=nil && w.accessToken==nil{
		return nil,e
	}

	// 向微信服务器请求JsapiTicket
	url := fmt.Sprintf(JSAPI_TICKET_URL, w.accessToken.Access_token)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	//{"errcode":0,"errmsg":"ok","ticket":"LIKLckvwlJT9cWIhEQTwfK7_uwZRuHdb8RaqHNtL2IvHpM6A4bNkahw9Q_-XTzq-gBaPRa2xPEwc7D-SVqPARg","expires_in":7200}
	err = json.Unmarshal(bytes, w.jsspiTicket)
	if err != nil || w.jsspiTicket == nil {
		return nil, err
	}

	// 修改JsapiTicket过期时间
	w.jsspiTicket.Expire_time = time.Now().Unix() + 7000
	tokenStr, err := json.Marshal(w.jsspiTicket)
	if err != nil {
		return w.jsspiTicket, err
	}

	// 存储JsapiTicket
	if err := w.storage.saveJsapiTicket(tokenStr); err != nil {
		return w.jsspiTicket, err
	}

	return w.jsspiTicket, nil

}

// 默认存储类型
// 存储格式：json
// 存储位置：当前目录
type _storage struct {
	accessTokenPath string
	jsapiTicketPath string
}

func (s *_storage) saveAccessToken(bytes []byte) error {
	return s.save(s.accessTokenPath, bytes)
}

func (s *_storage) getAccessToken() (string, error) {
	bytes, e := s.read(s.accessTokenPath)
	return string(bytes), e
}

func (s *_storage) saveJsapiTicket(bytes []byte) error {
	return s.save(s.jsapiTicketPath, bytes)
}

func (s *_storage) getJsapiTicket() (string, error) {
	bytes, e := s.read(s.jsapiTicketPath)
	return string(bytes), e
}

func (s *_storage) read(filename string) ([]byte, error) {
	_, e := os.Stat(filename)
	if e != nil {
		s.save(filename, []byte("{}"))
	}
	return ioutil.ReadFile(filename)
}

func (s *_storage) save(filename string, bytes []byte) error {
	return ioutil.WriteFile(filename, bytes, os.ModePerm)
}
