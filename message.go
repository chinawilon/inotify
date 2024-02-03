package main

import "fmt"

// Message 结构体定义
type Message struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At struct {
		IsAtAll bool `json:"isAtAll"`
	} `json:"at"`
}

// NewMessage 函数用于创建 Message 实例
func NewMessage(title string, errorFile string, errorMessage string) *Message {
	msg := &Message{
		MsgType: "text",
	}

	msg.At.IsAtAll = true

	// 如果是错误信息，附加错误文件信息
	msg.Text.Content += fmt.Sprintf("%s\n错误文件：%s\n错误信息: %s", title, errorFile, errorMessage)

	return msg
}
