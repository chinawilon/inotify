package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

func main() {
	arguments := os.Args[1:]
	if len(arguments) == 0 {
		log.Fatalln("缺少配置文件\nUsage: inotify /path/to/config/inotify.json")
	}
	var ins []*InotifyConf
	// 读取JSON文件内容
	fileContent, err := ioutil.ReadFile(arguments[0])
	if err != nil {
		log.Fatalln("Error reading JSON file:", err)
	}
	// 解析JSON数据到结构体
	err = json.Unmarshal(fileContent, &ins)
	if err != nil {
		log.Fatalln("Error unmarshalling JSON:", err)
	}

	// 使用多个goroutine
	var wg sync.WaitGroup
	for _, in := range ins {
		wg.Add(1)
		go func(in *InotifyConf) {
			defer wg.Done()
			// 初始化所有文件的最后偏移位置
			inotify := NewInotify(in)
			if err := inotify.InitSeek(); err != nil {
				fmt.Println(err)
				return
			}
			watcher, err := NewMyWatcher(inotify)
			if err != nil {
				fmt.Println(err)
				return
			}
			watcher.HandleEvents()
		}(in)
	}
	wg.Wait()
}
