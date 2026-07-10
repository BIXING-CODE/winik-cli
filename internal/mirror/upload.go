package mirror

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CosCert struct {
	TmpSecretID  string `json:"tmpSecretId"`
	TmpSecretKey string `json:"tmpSecretKey"`
	SessionToken string `json:"sessionToken"`
	Token        string `json:"token"`
	StartTime    int64  `json:"startTime"`
	ExpiredTime  int64  `json:"expiredTime"`
	Bucket       string `json:"bucket"`
	Region       string `json:"region"`
	KeyPath      string `json:"keyPath"`
}

// GetCosCert 取 COS 临时上传凭证。scopeType/metaType 为 PascalCase 枚举串
// （scope: Fragment/Avatar/...；meta: Picture/Video/...）。
func (c *Client) GetCosCert(scopeType, metaType string) (*CosCert, error) {
	body := map[string]any{"scopeType": scopeType, "metaType": metaType}
	var out CosCert
	if err := c.do("POST", "/upload/get-cos-cert", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// putObjectCOS 用临时凭证把文件 PUT 到腾讯云 COS（XML API + q-sign HMAC-SHA1 签名）。
func putObjectCOS(cert *CosCert, key string, data []byte) error {
	keyTime := fmt.Sprintf("%d;%d", time.Now().Unix()-60, time.Now().Unix()+900)
	signKey := hmacSHA1(cert.TmpSecretKey, keyTime)
	httpString := "put\n/" + key + "\n\n\n"
	hashed := sha1.Sum([]byte(httpString))
	stringToSign := "sha1\n" + keyTime + "\n" + hex.EncodeToString(hashed[:]) + "\n"
	signature := hmacSHA1(signKey, stringToSign)
	auth := "q-sign-algorithm=sha1&q-ak=" + cert.TmpSecretID +
		"&q-sign-time=" + keyTime + "&q-key-time=" + keyTime +
		"&q-header-list=&q-url-param-list=&q-signature=" + signature

	url := fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", cert.Bucket, cert.Region, key)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)
	token := cert.SessionToken
	if token == "" {
		token = cert.Token
	}
	if token != "" {
		req.Header.Set("x-cos-security-token", token)
	}
	resp, err := (&http.Client{Timeout: 120 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("COS PUT %s: HTTP %d: %.300s", key, resp.StatusCode, msg)
	}
	return nil
}

func hmacSHA1(key, msg string) string {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

type fragmentUploadRequest struct {
	ScopeType string `json:"scopeType,omitempty"`
	MetaType  string `json:"metaType,omitempty"`
	Source    string `json:"source,omitempty"`
	Width     *int   `json:"width,omitempty"`
	Height    *int   `json:"height,omitempty"`
}

type COSURLResponse struct {
	ID        *int   `json:"id"`
	SourceURL string `json:"sourceUrl"`
	Type      string `json:"type"`
}

// UploadPicture 完成单张图片的完整链路：取凭证 → COS PUT → fragment 注册，返回 fragment。
func (c *Client) UploadPicture(path string) (*COSURLResponse, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cert, err := c.GetCosCert("Fragment", "Picture")
	if err != nil {
		return nil, fmt.Errorf("取 COS 凭证失败: %w", err)
	}
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixMilli(), filepath.Base(path))
	key := strings.TrimPrefix(cert.KeyPath+fileName, "/")
	if err := putObjectCOS(cert, key, data); err != nil {
		return nil, err
	}

	var w, h *int
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
		w, h = &cfg.Width, &cfg.Height
	}
	body := map[string]any{"fragments": []fragmentUploadRequest{{
		ScopeType: "Fragment",
		MetaType:  "Picture",
		Source:    key,
		Width:     w,
		Height:    h,
	}}}
	var out []COSURLResponse
	if err := c.do("POST", "/user/fragment/batch-save", body, &out); err != nil {
		return nil, fmt.Errorf("fragment 注册失败: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("fragment 注册成功但返回为空")
	}
	return &out[0], nil
}
