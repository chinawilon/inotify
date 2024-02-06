package main

import (
	"fmt"
	"os"
	"sync"
)

type InotifyFile struct {
	lps int64 // 最后读取位置
	mu  sync.Mutex
	fs  *os.File
}

func NewInotifyFile(path string, fileSize int64, isCreate bool) (*InotifyFile, error) {
	// 这里保持长期打开状态
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// 如果是新创建，那么就从头开始计算
	if isCreate {
		fileSize = 0
	}
	return &InotifyFile{
		fs:  file,
		lps: fileSize,
	}, nil
}

// ReadNewContent 获取最新的内容
func (inf *InotifyFile) ReadNewContent(fileSize int64) string {
	if inf.lps >= fileSize {
		inf.lps = fileSize
		return ""
	}
	// 读取新增的内容
	buffer := make([]byte, fileSize-inf.lps)
	_, err := inf.fs.ReadAt(buffer, inf.lps)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return ""
	}
	// 把最后读取位置设置文件大小
	inf.lps = fileSize
	return string(buffer)
}
