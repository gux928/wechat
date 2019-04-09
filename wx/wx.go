package wx

import (
	"crypto/sha1"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"

	"github.com/clbanning/mxj"
)

type weixinQuery struct {
	Signature    string `json:"signature"`
	Timestamp    string `json:"timestamp"`
	Nonce        string `json:"nonce"`
	EncryptType  string `json:"encrypt_type"`
	MsgSignature string `json:"msg_signature"`
	Echostr      string `json:"echostr"`
}

//WeixinClient 微信端消息及回复
type WeixinClient struct {
	Token          string
	Query          weixinQuery
	Message        map[string]interface{}
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Methods        map[string]func() bool
}

//NewClient 初始化微信消息
func NewClient(r *http.Request, w http.ResponseWriter, token string) (*WeixinClient, error) {

	weixinClient := new(WeixinClient)

	weixinClient.Token = token
	weixinClient.Request = r
	weixinClient.ResponseWriter = w

	weixinClient.initWeixinQuery()

	if weixinClient.Query.Signature != weixinClient.signature() {
		return nil, fmt.Errorf("Invalid Signature")
	}

	return weixinClient, nil
}

func (thisClient *WeixinClient) initWeixinQuery() {

	var q weixinQuery

	q.Nonce = thisClient.Request.URL.Query().Get("nonce")
	q.Echostr = thisClient.Request.URL.Query().Get("echostr")
	q.Signature = thisClient.Request.URL.Query().Get("signature")
	q.Timestamp = thisClient.Request.URL.Query().Get("timestamp")
	q.EncryptType = thisClient.Request.URL.Query().Get("encrypt_type")
	q.MsgSignature = thisClient.Request.URL.Query().Get("msg_signature")

	thisClient.Query = q
}

func (thisClient *WeixinClient) signature() string {

	strs := sort.StringSlice{thisClient.Token, thisClient.Query.Timestamp, thisClient.Query.Nonce}
	sort.Strings(strs)
	str := ""
	for _, s := range strs {
		str += s
	}
	h := sha1.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (thisClient *WeixinClient) initMessage() error {

	body, err := ioutil.ReadAll(thisClient.Request.Body)

	if err != nil {
		return err
	}

	m, err := mxj.NewMapXml(body)

	if err != nil {
		return err
	}

	if _, ok := m["xml"]; !ok {
		return errors.New("invalid Message")
	}

	message, ok := m["xml"].(map[string]interface{})

	if !ok {
		return errors.New("invalid Field `xml` Type")
	}

	thisClient.Message = message

	log.Println(thisClient.Message)

	return nil
}

func (thisClient *WeixinClient) text() {

	inMsg, ok := thisClient.Message["Content"].(string)

	if !ok {
		return
	}

	var reply TextMessage

	reply.InitBaseData(thisClient, "text")
	reply.Content = value2CDATA(fmt.Sprintf("我收到的是：%s", inMsg))

	replyXML, err := xml.Marshal(reply)

	if err != nil {
		log.Println(err)
		thisClient.ResponseWriter.WriteHeader(403)
		return
	}

	thisClient.ResponseWriter.Header().Set("Content-Type", "text/xml")
	thisClient.ResponseWriter.Write(replyXML)
}

//Run 不知道干啥的
func (thisClient *WeixinClient) Run() {

	err := thisClient.initMessage()

	if err != nil {

		log.Println(err)
		thisClient.ResponseWriter.WriteHeader(403)
		return
	}

	MsgType, ok := thisClient.Message["MsgType"].(string)

	if !ok {
		thisClient.ResponseWriter.WriteHeader(403)
		return
	}

	switch MsgType {
	case "text":
		thisClient.text()
		break
	default:
		break
	}

	return
}
