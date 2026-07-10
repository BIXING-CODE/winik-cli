package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Service 是单个后端的登录态。
type Service struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

// Config 持久化在 ~/.winik-cli/config.json，按服务分账。
type Config struct {
	Bixing Service `json:"bixing"` // mirror（app.bixing.com.cn）
	Winik  Service `json:"winik"`  // winik（winik.bixing.com.cn / .ai）
}

func path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".winik-cli", "config.json"), nil
}

func Load() (*Config, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("解析 %s 失败: %w", p, err)
	}
	// 兼容旧版单服务格式 {base_url, token}（旧版只存 bixing）
	if c.Bixing.Token == "" && c.Winik.Token == "" {
		var legacy Service
		if json.Unmarshal(data, &legacy) == nil && legacy.Token != "" {
			c.Bixing = legacy
		}
	}
	return &c, nil
}

func (c *Config) Save() error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}
