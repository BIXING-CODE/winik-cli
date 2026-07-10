package mirror

// RequireVerifyCode 请求短信验证码。
func (c *Client) RequireVerifyCode(countryCode, phone string) error {
	body := map[string]any{
		"countryCode": countryCode,
		"phoneNumber": phone,
		"smsBiz":      "0",
	}
	return c.do("POST", "/auth/require-verify-code", body, nil)
}

type LoginResponse struct {
	UserID           int    `json:"userId"`
	Token            string `json:"token"`
	NewUser          bool   `json:"newUser"`
	ProfileCompleted bool   `json:"profileCompleted"`
	PhoneNumber      string `json:"phoneNumber"`
}

// Login 用验证码登录，返回 token。
func (c *Client) Login(countryCode, phone, code string) (*LoginResponse, error) {
	body := map[string]any{
		"countryCode": countryCode,
		"phoneNumber": phone,
		"code":        code,
	}
	var out LoginResponse
	if err := c.do("POST", "/auth/login", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type UserProfile struct {
	ID       *int   `json:"id"`
	UserID   *int   `json:"userId"`
	Nickname string `json:"nickname"`
}

// ProfileMe 查询当前登录用户。
func (c *Client) ProfileMe() (map[string]any, error) {
	var out map[string]any
	if err := c.do("GET", "/user/profile/me", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
