// Package mirror 封装 bixing mirror 后端的接口。
// 鉴权: header "token: <raw token>"；路径前缀 /v1。
package mirror

import "github.com/BIXING-CODE/winik-cli/internal/api"

const (
	BaseURLProd = "https://app.bixing.com.cn"
	BaseURLTest = "https://test-app.bixing.com.cn"
)

type Client struct{ api.Client }

func New(baseURL, token string) *Client {
	return &Client{api.Client{
		BaseURL:    baseURL,
		PathPrefix: "/v1",
		Token:      token,
		AuthHeader: "token",
	}}
}

// do 保留旧内部方法名，转发到共享层。
func (c *Client) do(method, path string, body, out any) error {
	return c.Do(method, path, body, out)
}
