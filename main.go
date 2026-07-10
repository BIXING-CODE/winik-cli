package main

import (
	"fmt"
	"os"

	"github.com/BIXING-CODE/winik-cli/cmd"
)

const usage = `winik-cli — mirror 发行动命令行工具

用法:
  winik-cli login      登录 mirror 账号（保存 token 到 ~/.winik-cli/config.json）
  winik-cli action     发布一条行动（--title/--content/--cover/...，见 action -h）
  winik-cli whoami     查看当前登录用户

全局:
  配置文件 ~/.winik-cli/config.json（base_url + token）
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}
	var err error
	switch os.Args[1] {
	case "login":
		err = cmd.Login(os.Args[2:])
	case "action":
		err = cmd.Action(os.Args[2:])
	case "whoami":
		err = cmd.Whoami(os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Printf("未知命令: %s\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}
