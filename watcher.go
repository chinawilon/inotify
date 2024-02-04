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

func (w *MyWatcher) HandleEvents(end chan<- struct{}) {
	defer w.Watcher.Close()
	for {
		select {
		case event, ok := <-w.Watcher.Events:
			if !ok {
				end <- struct{}{}
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					fmt.Println("New directory created:", event.Name)
					w.Watcher.Add(event.Name)
				} else {
					w.Inotify.JudgeContent(event.Name)
				}
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				w.Inotify.JudgeContent(event.Name)
			}
		case err, ok := <-w.Watcher.Errors:
			if !ok {
				log.Println("Error:", err)
				end <- struct{}{}
				return
			}
		}
	}
}
