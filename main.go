package main

import (
	"fmt"
	"os"

	"github.com/BIXING-CODE/winik-cli/cmd"
)

const usage = `winik-cli — bixing(mirror) / winik 双服务命令行工具

用法:
  winik-cli login bixing   登录 mirror（--phone 验证码 或 --token 直填；--prod 切生产）
  winik-cli login winik    登录 winik（--email 验证码 或 --token 直填；--intl 切海外区）
  winik-cli action         发布一条行动到 bixing（--title/--content/--cover/...，见 action -h）
  winik-cli place          百度地点检索预览（--query 深圳工业艺术中心 --city 深圳）
  winik-cli whoami         查看登录态（可跟 bixing / winik，缺省都查）

全局:
  配置文件 ~/.winik-cli/config.json（按服务分存 base_url + token）
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
	case "place":
		err = cmd.Place(os.Args[2:])
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
