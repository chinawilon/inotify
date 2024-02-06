package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
)

type MyWatcher struct {
	Inotify *InotifyConf
	Watcher *fsnotify.Watcher
}

func NewMyWatcher(inotify *InotifyConf) (*MyWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = filepath.Walk(inotify.DirPath, func(path string, info os.FileInfo, err error) error {
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
	return &MyWatcher{
		Inotify: inotify,
		Watcher: watcher,
	}, err
}

func (w *MyWatcher) HandleEvents() {
	for {
		select {
		case event, ok := <-w.Watcher.Events:
			if !ok {
				return
			}
			// 如果是删除事件，先删除inotify中的记录
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				w.Inotify.Delete(event.Name)
				continue
			}
			// 这里进一步确保文件没有被删除
			info, err := os.Stat(event.Name)
			if err != nil {
				fmt.Println("Os stat error:", err)
				continue
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				if err == nil && info.IsDir() {
					w.Watcher.Add(event.Name)
				} else {
					// 这里必须是同步的，不然会影响后面的读取
					w.Inotify.Create(event.Name)
				}
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				go w.Inotify.Write(event.Name, info.Size())
			}
		case err, ok := <-w.Watcher.Errors:
			if !ok {
				log.Println("Error:", err)
				return
			}
		}
	}
}
