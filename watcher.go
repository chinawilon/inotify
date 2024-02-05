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
			info, err := os.Stat(event.Name)
			if err != nil {
				fmt.Println("Os stat error:", err)
				continue
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				if err == nil && info.IsDir() {
					fmt.Println("New directory created:", event.Name)
					w.Watcher.Add(event.Name)
				} else {
					go w.Inotify.JudgeContent(event.Name, info.Size())
				}
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				go w.Inotify.JudgeContent(event.Name, info.Size())
			}
		case err, ok := <-w.Watcher.Errors:
			if !ok {
				log.Println("Error:", err)
				return
			}
		}
	}
}
