package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

type InotifyConf struct {
	DirPath          string
	FilterFile       string
	ErrorKey         string
	NoticeTitle      string
	DingdingAPI      string
	filterRe         *regexp.Regexp
	errorRe          *regexp.Regexp
	lastReadPosition map[string]int64
}

func NewInotify(confPath string) *InotifyConf {
	var inotify = InotifyConf{}
	// 读取JSON文件内容
	fileContent, err := ioutil.ReadFile(confPath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return nil
	}
	// 解析JSON数据到结构体
	err = json.Unmarshal(fileContent, &inotify)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil
	}
	inotify.filterRe = regexp.MustCompile(inotify.FilterFile)
	inotify.errorRe = regexp.MustCompile(inotify.ErrorKey)
	inotify.lastReadPosition = make(map[string]int64)

	return &inotify
}

func (inotify *InotifyConf) IsFilterFile(file string) bool {
	return inotify.filterRe.MatchString(file)
}

func (inotify *InotifyConf) HasErrorKey(content string) bool {
	return inotify.errorRe.MatchString(content)
}

func (inotify *InotifyConf) InitSeek() error {
	return filepath.Walk(inotify.DirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !inotify.IsFilterFile(path) {
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
		inotify.lastReadPosition[path] = lastOffset
		return nil
	})
}

func (inotify *InotifyConf) SendAlert(changedFile, newContent string) error {

	msg := NewMessage(inotify.NoticeTitle, changedFile, newContent)
	jsonData, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		fmt.Println("JSON 编码失败:", err)
		return err
	}
	//fmt.Println(string(jsonData))
	resp, err := http.Post(inotify.DingdingAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Println("Alert sent for", changedFile)
	return nil
}

func (inotify *InotifyConf) ReadNewContent(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 获取上一次读取的位置
	lastPosition, ok := inotify.lastReadPosition[filePath]
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

	// 检查是否有错误发生
	if err := scanner.Err(); err != nil {
		fmt.Println("Error:", err)
	}

	// 更新上一次读取的位置
	inotify.lastReadPosition[filePath], _ = file.Seek(0, 1)
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return newContent, nil
}

// JudgeContent 判断文件和新增内容是否满足条件
func (inotify *InotifyConf) JudgeContent(file string) {
	if !inotify.IsFilterFile(file) {
		return
	}
	content, err := inotify.ReadNewContent(file)
	if err != nil {
		log.Println("Error reading new content:", err)
		return
	}
	if inotify.HasErrorKey(content) {
		err := inotify.SendAlert(file, content)
		if err != nil {
			log.Println("sendAlert err:", err)
		}
	}
}
