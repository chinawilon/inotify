package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type InotifyFile struct {
	lps int64  // 最后读取位置
	fp  string // 文件路径
	lt  int64  // 最后访问时间
	mu  sync.Mutex
	fs  *os.File
}

func NewInotifyFile(path string, fileSize int64) (*InotifyFile, error) {
	return &InotifyFile{
		fp:  path,
		lps: fileSize,
		lt:  time.Now().Unix(),
	}, nil
}

func (inf *InotifyFile) Close() error {
	inf.fs = nil
	return inf.fs.Close()
}

// ReadNewContent 获取最新的内容
func (inf *InotifyFile) ReadNewContent(fileSize int64) string {
	inf.lt = time.Now().Unix() // 最近访问时间
	if inf.lps >= fileSize {
		inf.lps = fileSize
		return ""
	}
	// 判断文件是否打开了
	if inf.fs == nil {
		f, err := os.Open(inf.fp)
		if err != nil {
			fmt.Println("打开文件失败", err)
			return ""
		}
		inf.fs = f
		// 这里使用一个goroutine进行超时跟踪
		go func() {
			for {
				timer := time.NewTimer(1 * time.Minute)
				last := inf.lt
				<-timer.C
				// 如果超时时间内都没有访问那么就关闭fd
				if last == inf.lt {
					inf.Close()
					break
				}
			}
		}()
	}
	// 读取新增的内容
	buffer := make([]byte, fileSize-inf.lps)
	_, err := inf.fs.ReadAt(buffer, inf.lps)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return ""
	}
	// 设置最后读取位置
	inf.lps = fileSize
	return string(buffer)
}
