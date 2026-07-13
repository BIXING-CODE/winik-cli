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
# 1. 登录（bixing 与 winik 两套账号独立）
./winik-cli login bixing --phone 13800138000   # mirror，默认测试环境，--prod 切生产
# 或直接填已有 token 跳过验证码（app 抓包 header 里的 token 字段）
./winik-cli login bixing --token "xxx" [--prod]
./winik-cli login winik --email me@x.com [--intl]  # winik，默认 cn 区
./winik-cli login winik --token "xxx" [--intl]

# 2. 线下行动的地点必须是真实 POI（名称+经纬度），否则 app 端点开地点会报"地点信息错误"。
#    先用 place 命令检索预览（百度地图，和 app 选点同源同坐标系 bd09ll）：
./winik-cli place --query "工业艺术中心" --city 深圳
# [0] 深圳滨海艺术中心 ... lat=22.550980 lng=113.888602
# [1] ...

# 3. 发行动（本地图片自动上传 COS + 注册 fragment；--place 自动解析地点，--place-index 选候选序号）
./winik-cli action \
  --title "周末看展" \
  --content "戴上耳机逛展，走起" \
  --cover ./cover.jpg \
  --image ./p1.jpg --image ./p2.jpg \
  --start-at "2026-07-15 14:00" \
  --place "工业艺术中心" --place-city 深圳 --place-index 0 \
  --price 0

# 也可手动指定地点，但必须三件套齐全（location + lat + lng，bd09ll 坐标）：
#   --location "香山公园东门" --lat 39.99 --lng 116.19

# 线上行动（不需要地点）+ 文字时间 + 仅自己可见 + JSON 输出
./winik-cli action --title T --content C --cover-url https://... \
  --time-desc "本周任意晚上" --online --price 10 --self --json

./winik-cli whoami
```

## API 链路（逆向自 mirror Flutter 端，2026-07-10）

- Base：`https://app.bixing.com.cn`（prod）/ `https://test-app.bixing.com.cn`（test），路径前缀 `/v1`
- 鉴权：header `token: <raw>`（无 Bearer）；登录 `POST /v1/auth/require-verify-code` → `POST /v1/auth/login`
- 信封：`{code, message, data}`，code 0/200 成功
- 发行动：`POST /v1/user/action`（ActionRequest，camelCase；startAt 与 timeRangeDesc 二选一；mode 0=线下/1=线上；visibleStatus 0=self/1=public）
- **startAt 读写格式不对称（v0.3.2 修，实测踩坑）**：后端**返回**用 `yyyy-MM-dd HH:mm:ss`，**写入**只收 ISO8601（`2026-07-15T19:00:00.000`，mirror app 端 `toIso8601String()` 同款）——空格格式必 `451 参数错误`。CLI `--start-at` 入参仍是 `"2026-07-15 19:00"`，出线自动转 ISO
- 图片：`POST /v1/upload/get-cos-cert`（scopeType=Fragment/metaType=Picture，PascalCase）→ 腾讯云 COS PUT（q-sign HMAC-SHA1 + x-cos-security-token）→ `POST /v1/user/fragment/batch-save` 得 fragment id / sourceUrl
- 地点：百度地图 `GET place/v3/region`（server AK 同 app 端 `AppConfig.baiduMapServerKey`），坐标系 **bd09ll**，线下行动的 location/lat/lng 三件套缺一不可（否则 app 端点开地点报错）
- 参照实现：`mirror_frontend/lib/app/action/create_action_v2/create_action_v2_controller.dart` + `data/apis/api_user_provider.dart` + `core/net/mr_net.dart` + `core/utils/cos_utils.dart`
