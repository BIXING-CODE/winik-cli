// Package api 是 bixing(mirror) 与 winik 共用的最小 HTTP 客户端。
// 两侧响应信封相同：{code, message, data}，code 0/200 为成功；仅鉴权头不同。
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	PathPrefix string // mirror 为 "/v1"，winik 为 ""
	Token      string
	AuthHeader string // mirror: "token"；winik: "Authorization"
	AuthPrefix string // mirror: ""；winik: "Bearer "
	HTTP       *http.Client
}

type envelope struct {
	Code    *int            `json:"code"`
	Status  *int            `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Results json.RawMessage `json:"results"`
	Result  json.RawMessage `json:"result"`
}

// Do 发送 JSON 请求并把信封里的 data 解到 out（out 可为 nil）。
func (c *Client) Do(method, path string, body any, out any) error {
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 60 * time.Second}
	}
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, c.BaseURL+c.PathPrefix+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set(c.AuthHeader, c.AuthPrefix+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("%s %s: HTTP %d，响应非 JSON: %.200s", method, path, resp.StatusCode, raw)
	}
	code := -1
	if env.Code != nil {
		code = *env.Code
	} else if env.Status != nil {
		code = *env.Status
	}
	if code != 0 && code != 200 {
		return fmt.Errorf("%s %s: code=%d message=%s", method, path, code, env.Message)
	}
	if out != nil {
		data := env.Data
		if data == nil {
			data = env.Results
		}
		if data == nil {
			data = env.Result
		}
		if data == nil {
			return fmt.Errorf("%s %s: 成功但无 data 字段", method, path)
		}
		return json.Unmarshal(data, out)
	}
	return nil
}
