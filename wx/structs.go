package wx

import (
	"encoding/xml"
	"strconv"
	"time"
)

//Base 基础微信消息结构
type Base struct {
	FromUserName CDATAText
	ToUserName   CDATAText
	MsgType      CDATAText
	CreateTime   CDATAText
}

//InitBaseData 初始化基础消息
func (b *Base) InitBaseData(w *WeixinClient, msgtype string) {

	b.FromUserName = value2CDATA(w.Message["ToUserName"].(string))
	b.ToUserName = value2CDATA(w.Message["FromUserName"].(string))
	b.CreateTime = value2CDATA(strconv.FormatInt(time.Now().Unix(), 10))
	b.MsgType = value2CDATA(msgtype)
}

//CDATAText 微信自定义消息类型
type CDATAText struct {
	Text string `xml:",innerxml"`
}

//TextMessage 文本消息结构
type TextMessage struct {
	XMLName xml.Name `xml:"xml"`
	Base
	Content CDATAText
}
