#!/bin/sh
# nav 安装脚本（macOS / Linux，sh 兼容）
#
# 用法:
#   curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh | sh
#   NAV_URL=<自定义 Release 前缀> ./install.sh     # 覆盖下载源
#   NAV_LOCAL=1 ./install.sh                        # 强制用本地 dist/（开发期）
#   NAV_DEST=/usr/local/bin ./install.sh            # 自定义安装目录
set -e

REPO="${NAV_REPO:-lodrg/nav}"
BASE="${NAV_URL:-https://github.com/$REPO/releases/latest/download}"

detect() {
  case "$(uname -s)" in
    Darwin) os=darwin ;;
    Linux)  os=linux ;;
    *) echo "不支持的平台: $(uname -s)（Windows 请用 install.ps1）" >&2; exit 1 ;;
  esac
  case "$(uname -m)" in
    arm64|aarch64) arch=arm64 ;;
    x86_64|amd64)  arch=amd64 ;;
    *) echo "不支持的架构: $(uname -m)" >&2; exit 1 ;;
  esac
}
detect

DEST="${NAV_DEST:-$HOME/.local/bin}"
mkdir -p "$DEST"

local_src="$(dirname "$0")/dist/nav-$os-$arch"
if [ "${NAV_LOCAL:-0}" = "1" ] || { [ -z "${NAV_URL:-}" ] && [ -f "$local_src" ]; }; then
  echo "从本地安装: $local_src"
  cp "$local_src" "$DEST/nav"
else
  url="$BASE/nav-$os-$arch"
  echo "下载 $url"
  curl -fsSL "$url" -o "$DEST/nav" || {
    echo "下载失败: $url（检查网络或 NAV_URL）" >&2
    exit 1
  }
fi
chmod +x "$DEST/nav"

case ":$PATH:" in
  *":$DEST:"*) ;;
  *) echo "提示: $DEST 不在 PATH，请先加入: export PATH=\"$DEST:\$PATH\"" >&2 ;;
esac

echo "已安装: $DEST/nav ($os/$arch)"
"$DEST/nav" --version
