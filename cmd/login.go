package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/BIXING-CODE/winik-cli/internal/config"
	"github.com/BIXING-CODE/winik-cli/internal/mirror"
)

// Login 走 require-verify-code + login 两步，token 落盘。
func Login(args []string) error {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	phone := fs.String("phone", "", "手机号（必填）")
	country := fs.String("country", "86", "国家码")
	prod := fs.Bool("prod", false, "使用生产环境（默认测试环境 test-app）")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *phone == "" {
		return fmt.Errorf("--phone 必填")
	}

	baseURL := mirror.BaseURLTest
	if *prod {
		baseURL = mirror.BaseURLProd
	}
	client := mirror.New(baseURL, "")

	if err := client.RequireVerifyCode(*country, *phone); err != nil {
		return fmt.Errorf("请求验证码失败: %w", err)
	}
	fmt.Printf("验证码已发送到 %s，请输入: ", *phone)
	reader := bufio.NewReader(os.Stdin)
	code, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	code = strings.TrimSpace(code)

	resp, err := client.Login(*country, *phone, code)
	if err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	cfg := &config.Config{BaseURL: baseURL, Token: resp.Token}
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("登录成功 userId=%d（token 已存 ~/.winik-cli/config.json，环境 %s）\n", resp.UserID, baseURL)
	return nil
}

// Whoami 查询当前登录用户。
func Whoami(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Token == "" {
		return fmt.Errorf("未登录，先执行 winik-cli login --phone <手机号>")
	}
	client := mirror.New(cfg.BaseURL, cfg.Token)
	me, err := client.ProfileMe()
	if err != nil {
		return err
	}
	fmt.Printf("环境: %s\n", cfg.BaseURL)
	for _, k := range []string{"id", "userId", "nickname", "phoneNumber"} {
		if v, ok := me[k]; ok {
			fmt.Printf("%s: %v\n", k, v)
		}
	}
	return nil
}
