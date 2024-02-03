package main

import (
	"log"
	"os"
)

func main() {
	arguments := os.Args[1:]
	if len(arguments) == 0 {
		log.Fatalln("缺少配置文件\nUsage: inotify /path/to/config/inotify.json")
	}

	// 初始化所有文件的最后偏移位置
	inotify := NewInotify(arguments[0])
	if err := inotify.InitSeek(); err != nil {
		log.Fatalln(err)
	}

	watcher, err := NewMyWatcher(inotify)
	if err != nil {
		log.Fatalln(err)
	}
	end := make(chan struct{})
	go watcher.HandleEvents(end)
	<-end
}
