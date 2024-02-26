package notifytypes

import (
	"fmt"
	"os/exec"
)

type Shell struct {
	Command string `json:"command"`
}

func (s *Shell) Notify(changedFile, newContent string) error {
	// 创建 Cmd 结构体
	command := exec.Command(s.Command, changedFile, newContent)
	// 执行脚本并获取输出
	output, err := command.CombinedOutput()
	if err != nil {
		fmt.Println("Error executing script:", err)
		return err
	}
	// 打印脚本输出
	fmt.Println(string(output))
	return nil
}
