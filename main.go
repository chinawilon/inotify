package main

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type inotifyConf struct {
	DirPath     string
	ErrorKey    string
	NoticeTitle string
	DingdingAPI string
}

func main() {
	arguments := os.Args[1:]
	if len(arguments) == 0 {
		fmt.Println("缺少配置文件！")
		fmt.Println("usage: inotify /path/to/config/inotify.json ")
		os.Exit(0)
	}

	// 读取JSON文件内容
	fileContent, err := ioutil.ReadFile(arguments[0])
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	// 定义结构体变量，用于存储解析后的数据
	var conf inotifyConf

	// 解析JSON数据到结构体
	err = json.Unmarshal(fileContent, &conf)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	// 初始化所有文件的最后偏移位置
	initSeek(conf.DirPath)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = filepath.Walk(conf.DirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	go handleEvents(&conf, watcher)

	select {}
}
