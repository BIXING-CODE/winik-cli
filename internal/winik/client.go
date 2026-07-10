// Package winik 封装 winik 后端（FastAPI）的接口。
// 鉴权: Authorization: Bearer <token>；登录: 邮箱验证码。
package winik

import "github.com/BIXING-CODE/winik-cli/internal/api"

const (
	BaseURLCN   = "https://winik.bixing.com.cn"
	BaseURLIntl = "https://winik.bixing.ai"
)

type Client struct{ api.Client }

func New(baseURL, token string) *Client {
	return &Client{api.Client{
		BaseURL:    baseURL,
		Token:      token,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	}}
}

// SendCode 发送邮箱验证码（dev 环境 SMTP 未配置时验证码打在服务端 docker logs）。
func (c *Client) SendCode(email string) error {
	return c.Do("POST", "/users/email/send-code", map[string]string{"email": email}, nil)
}

type LoginResponse struct {
	Token     string `json:"token"`
	UserID    int    `json:"user_id"`
	IsNewUser bool   `json:"is_new_user"`
}

// Login 邮箱+验证码登录（不存在则自动注册；白名单邮箱 code 可填 "1234"）。
func (c *Client) Login(email, code string) (*LoginResponse, error) {
	body := map[string]string{"email": email, "code": code}
	var out LoginResponse
	if err := c.Do("POST", "/users/login", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Me 查询当前用户。
func (c *Client) Me() (map[string]any, error) {
	var out map[string]any
	if err := c.Do("GET", "/users/me", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
