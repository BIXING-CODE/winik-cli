package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/BIXING-CODE/winik-cli/internal/config"
	"github.com/BIXING-CODE/winik-cli/internal/mirror"
	"github.com/BIXING-CODE/winik-cli/internal/winik"
)

const loginUsage = `用法:
  winik-cli login bixing --phone <手机号> [--prod]      mirror 手机验证码登录（默认测试环境）
  winik-cli login bixing --token <token> [--prod]       mirror 直填 token
  winik-cli login winik  --email <邮箱> [--intl]        winik 邮箱验证码登录（默认 cn 区）
  winik-cli login winik  --token <token> [--intl]       winik 直填 token
`

// Login 分服务登录：login bixing / login winik。
func Login(args []string) error {
	if len(args) == 0 {
		fmt.Print(loginUsage)
		return fmt.Errorf("请指定服务: bixing 或 winik")
	}
	switch args[0] {
	case "bixing":
		return loginBixing(args[1:])
	case "winik":
		return loginWinik(args[1:])
	default:
		fmt.Print(loginUsage)
		return fmt.Errorf("未知服务: %s（可选 bixing / winik）", args[0])
	}
}

func loginBixing(args []string) error {
	fs := flag.NewFlagSet("login bixing", flag.ExitOnError)
	phone := fs.String("phone", "", "手机号（与 --token 二选一）")
	country := fs.String("country", "86", "国家码")
	token := fs.String("token", "", "直接写入已有 token，跳过验证码登录")
	prod := fs.Bool("prod", false, "使用生产环境（默认测试环境 test-app）")
	if err := fs.Parse(args); err != nil {
		return err
	}

	baseURL := mirror.BaseURLTest
	if *prod {
		baseURL = mirror.BaseURLProd
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if *token != "" {
		cfg.Bixing = config.Service{BaseURL: baseURL, Token: *token}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("bixing token 已写入（环境 %s），可用 winik-cli whoami bixing 验证\n", baseURL)
		return nil
	}

	if *phone == "" {
		return fmt.Errorf("--phone 与 --token 必填其一")
	}
	client := mirror.New(baseURL, "")
	if err := client.RequireVerifyCode(*country, *phone); err != nil {
		return fmt.Errorf("请求验证码失败: %w", err)
	}
	code, err := promptCode(fmt.Sprintf("验证码已发送到 %s，请输入: ", *phone))
	if err != nil {
		return err
	}
	resp, err := client.Login(*country, *phone, code)
	if err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}
	cfg.Bixing = config.Service{BaseURL: baseURL, Token: resp.Token}
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("bixing 登录成功 userId=%d（环境 %s）\n", resp.UserID, baseURL)
	return nil
}

func loginWinik(args []string) error {
	fs := flag.NewFlagSet("login winik", flag.ExitOnError)
	email := fs.String("email", "", "邮箱（与 --token 二选一）")
	token := fs.String("token", "", "直接写入已有 token，跳过验证码登录")
	intl := fs.Bool("intl", false, "使用海外区 winik.bixing.ai（默认国内区 winik.bixing.com.cn）")
	if err := fs.Parse(args); err != nil {
		return err
	}

	baseURL := winik.BaseURLCN
	if *intl {
		baseURL = winik.BaseURLIntl
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if *token != "" {
		cfg.Winik = config.Service{BaseURL: baseURL, Token: *token}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("winik token 已写入（环境 %s），可用 winik-cli whoami winik 验证\n", baseURL)
		return nil
	}

	if *email == "" {
		return fmt.Errorf("--email 与 --token 必填其一")
	}
	client := winik.New(baseURL, "")
	if err := client.SendCode(*email); err != nil {
		return fmt.Errorf("请求验证码失败: %w", err)
	}
	code, err := promptCode(fmt.Sprintf("验证码已发送到 %s（dev 环境看服务端 docker logs），请输入: ", *email))
	if err != nil {
		return err
	}
	resp, err := client.Login(*email, code)
	if err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}
	cfg.Winik = config.Service{BaseURL: baseURL, Token: resp.Token}
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("winik 登录成功 user_id=%d（环境 %s）\n", resp.UserID, baseURL)
	return nil
}

func promptCode(prompt string) (string, error) {
	fmt.Print(prompt)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// Whoami 查询登录态：whoami [bixing|winik]，缺省两个都查。
func Whoami(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	which := ""
	if len(args) > 0 {
		which = args[0]
	}

	shown := false
	if which == "" || which == "bixing" {
		if cfg.Bixing.Token == "" {
			fmt.Println("bixing: 未登录（winik-cli login bixing ...）")
		} else {
			me, err := mirror.New(cfg.Bixing.BaseURL, cfg.Bixing.Token).ProfileMe()
			if err != nil {
				fmt.Printf("bixing(%s): token 无效或请求失败: %v\n", cfg.Bixing.BaseURL, err)
			} else {
				fmt.Printf("bixing(%s): id=%v nickname=%v\n", cfg.Bixing.BaseURL, me["id"], me["nickname"])
			}
		}
		shown = true
	}
	if which == "" || which == "winik" {
		if cfg.Winik.Token == "" {
			fmt.Println("winik: 未登录（winik-cli login winik ...）")
		} else {
			me, err := winik.New(cfg.Winik.BaseURL, cfg.Winik.Token).Me()
			if err != nil {
				fmt.Printf("winik(%s): token 无效或请求失败: %v\n", cfg.Winik.BaseURL, err)
			} else {
				fmt.Printf("winik(%s): id=%v email=%v name=%v\n", cfg.Winik.BaseURL, me["id"], me["email"], me["name"])
			}
		}
		shown = true
	}
	if !shown {
		return fmt.Errorf("未知服务: %s（可选 bixing / winik）", which)
	}
	return nil
}
