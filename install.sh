#!/bin/sh
# winik-cli 一键安装脚本
# 用法: curl -fsSL https://raw.githubusercontent.com/BIXING-CODE/winik-cli/main/install.sh | sh
set -e

REPO="BIXING-CODE/winik-cli"
INSTALL_DIR="${WINIK_CLI_INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac

ASSET="winik-cli-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "下载 ${URL} ..."
TMP=$(mktemp)
curl -fL -o "$TMP" "$URL"
chmod +x "$TMP"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "$INSTALL_DIR/winik-cli"
else
  echo "需要 sudo 写入 $INSTALL_DIR"
  sudo mv "$TMP" "$INSTALL_DIR/winik-cli"
fi

# macOS 去隔离属性，避免 Gatekeeper 拦截
if [ "$OS" = "darwin" ]; then
  xattr -d com.apple.quarantine "$INSTALL_DIR/winik-cli" 2>/dev/null || true
fi

echo "安装完成: $INSTALL_DIR/winik-cli"
"$INSTALL_DIR/winik-cli" -h || true
