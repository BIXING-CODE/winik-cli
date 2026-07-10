# winik-cli

mirror（`/Users/mac/Documents/flutter/bixing/mirror_frontend`）"发行动"功能的 Go 命令行客户端。
纯标准库、零依赖、单二进制。也是 winik-server 调 mirror 接口的参考实现。

## 安装（用户视角）

```bash
# 一键安装（需 repo 已发布 Release）
curl -fsSL https://raw.githubusercontent.com/BIXING-CODE/winik-cli/main/install.sh | sh

# 或到 Releases 页手动下载对应平台二进制
```

## 构建（开发视角）

```bash
go build -o winik-cli .
# 发布新版本：git tag v0.1.0 && git push --tags  → GitHub Actions 自动编全平台二进制挂 Releases
```

## 用法

```bash
# 1. 登录（默认测试环境 test-app.bixing.com.cn，--prod 切生产）
./winik-cli login --phone 13800138000

# 2. 发行动（本地图片自动上传 COS + 注册 fragment）
./winik-cli action \
  --title "周末爬山" \
  --content "香山看日出，走起" \
  --cover ./cover.jpg \
  --image ./p1.jpg --image ./p2.jpg \
  --start-at "2026-07-15 06:00" \
  --location "香山公园东门" --lat 39.99 --lng 116.19 \
  --price 0

# 线上行动 + 文字时间 + 仅自己可见 + JSON 输出
./winik-cli action --title T --content C --cover-url https://... \
  --time-desc "本周任意晚上" --online --price 10 --self --json

./winik-cli whoami
```

## API 链路（逆向自 mirror Flutter 端，2026-07-10）

- Base：`https://app.bixing.com.cn`（prod）/ `https://test-app.bixing.com.cn`（test），路径前缀 `/v1`
- 鉴权：header `token: <raw>`（无 Bearer）；登录 `POST /v1/auth/require-verify-code` → `POST /v1/auth/login`
- 信封：`{code, message, data}`，code 0/200 成功
- 发行动：`POST /v1/user/action`（ActionRequest，camelCase；startAt 与 timeRangeDesc 二选一，startAt 线格式 `yyyy-MM-dd HH:mm:ss`；mode 0=线下/1=线上；visibleStatus 0=self/1=public）
- 图片：`POST /v1/upload/get-cos-cert`（scopeType=Fragment/metaType=Picture，PascalCase）→ 腾讯云 COS PUT（q-sign HMAC-SHA1 + x-cos-security-token）→ `POST /v1/user/fragment/batch-save` 得 fragment id / sourceUrl
- 参照实现：`mirror_frontend/lib/app/action/create_action_v2/create_action_v2_controller.dart` + `data/apis/api_user_provider.dart` + `core/net/mr_net.dart` + `core/utils/cos_utils.dart`
