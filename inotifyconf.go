package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

type InotifyConf struct {
	DirPath     string
	FilterFile  string
	ErrorKey    string
	NoticeTitle string
	DingdingAPI string
	ExcludeKey  string
	filterRe    *regexp.Regexp
	errorRe     *regexp.Regexp
	excludeRe   *regexp.Regexp
	ifs         map[string]*InotifyFile
	mu          sync.Mutex
	gp          *Group
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
	inotify.excludeRe = regexp.MustCompile(inotify.ExcludeKey)
	inotify.ifs = make(map[string]*InotifyFile)
	inotify.gp = &Group{}
	return &inotify
}

func (inotify *InotifyConf) IsFilterFile(file string) bool {
	return inotify.filterRe.MatchString(file)
}

func (inotify *InotifyConf) HasErrorKey(content string) bool {
	return inotify.errorRe.MatchString(content) && (inotify.ExcludeKey == "" || !inotify.excludeRe.MatchString(content))
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
		if _, ok := inotify.ifs[path]; !ok {
			inotify.ifs[path], err = NewInotifyFile(path, info.Size())
		}
		return err
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

// JudgeContent 判断文件和新增内容是否满足条件
func (inotify *InotifyConf) JudgeContent(file string, fileSize int64) {
	var err error
	if !inotify.IsFilterFile(file) {
		return
	}
	inotify.mu.Lock()
	inf, ok := inotify.ifs[file]
	if !ok {
		inf, err = NewInotifyFile(file, fileSize)
		if err != nil {
			fmt.Println("New inotify file error:", err)
			inotify.mu.Unlock()
			return
		}
		inotify.ifs[file] = inf
	}
	inotify.mu.Unlock()
	inotify.gp.Do(file, func() {
		content := inf.ReadNewContent(fileSize)
		if inotify.HasErrorKey(content) {
			err = inotify.SendAlert(file, content)
			if err != nil {
				fmt.Println("Send alert error:", err)
			}
		}
	})

}
