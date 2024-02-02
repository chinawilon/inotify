package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var lastReadPosition = make(map[string]int64)

func initSeek(dir string) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		defer file.Close()

		// 获取文件大小
		lastOffset, err := file.Seek(0, 2)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		// 记录最后偏移位置（文件大小）
		lastReadPosition[path] = lastOffset
		return nil
	})

	if err != nil {
		fmt.Printf("错误：%v\n", err)
	}
}

func judgeContent(conf *inotifyConf, file string) {
	content, err := readNewContent(file)
	if err != nil {
		log.Println("Error reading new content:", err)
	} else {
		err := sendAlert(conf, file, content)
		if err != nil {
			log.Println("sendAlert err:", err)
		}
	}
}

func handleEvents(conf *inotifyConf, watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					fmt.Println("New directory created:", event.Name)
					watcher.Add(event.Name)
				} else {
					judgeContent(conf, event.Name)
				}
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				judgeContent(conf, event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error:", err)
		}
	}
}

func readNewContent(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 获取上一次读取的位置
	lastPosition, ok := lastReadPosition[filePath]
	if !ok {
		lastPosition = 0
	}

	// 如果文件大小改变了？就获取当前文件的最后偏移
	fileSize, _ := file.Seek(0, 2)
	if lastPosition > fileSize {
		// 这里没有其他更好的办法，因为某些原因导致了日志的转换
		lastPosition = fileSize
	}

	// 将文件指针移动到上一次读取的位置
	file.Seek(lastPosition, 0)

	// 使用bufio.Scanner读取新增的内容
	scanner := bufio.NewScanner(file)
	var newContent string
	for scanner.Scan() {
		newContent += scanner.Text() + "\n"
	}

	fmt.Println("content:", newContent)
	// 检查是否有错误发生
	if err := scanner.Err(); err != nil {
		fmt.Println("Error:", err)
	}

	// 更新上一次读取的位置
	lastReadPosition[filePath], _ = file.Seek(0, 1)
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return newContent, nil
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
	msg.Text.Content += fmt.Sprintf("\n%s\n错误文件：%s\n错误信息: %s", title, errorFile, errorMessage)

	return msg
}

func sendAlert(conf *inotifyConf, changedFile, newContent string) error {
	if strings.Contains(newContent, conf.ErrorKey) {
		msg := NewMessage(conf.NoticeTitle, changedFile, newContent)
		jsonData, err := json.MarshalIndent(msg, "", "  ")
		if err != nil {
			fmt.Println("JSON 编码失败:", err)
			return err
		}
		resp, err := http.Post(conf.DingdingAPI, "application/json", bytes.NewBuffer([]byte(jsonData)))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var bd = make([]byte, 100)
		resp.Body.Read(bd)
		fmt.Println(string(bd))

		fmt.Println("Alert sent for", changedFile)
	} else {
		fmt.Println("No alert for", changedFile)
	}

	return nil
}
