package notifytypes

import (
	"encoding/json"
	"fmt"
)

type DingDing struct {
	Title string `json:"title"`
	Api   string `json:"api"`
}

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

func (dd *DingDing) Notify(changedFile, newContent string) error {

	msg := NewMessage(dd.Title, changedFile, newContent)
	jsonData, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		fmt.Println("JSON 编码失败:", err)
		return err
	}
	fmt.Println(string(jsonData))
	//resp, err := http.Post(dd.Api, "application/json", bytes.NewBuffer(jsonData))
	//if err != nil {
	//	return err
	//}
	//defer resp.Body.Close()
	fmt.Println("new content: ", newContent)
	fmt.Println("Alert sent for", changedFile)
	return nil
}
