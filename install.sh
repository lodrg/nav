#!/bin/sh
# nav 安装脚本（macOS / Linux，sh 兼容）
#
# 用法:
#   curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh | sh
#   NAV_URL=<前缀> ./install.sh        # 覆盖下载源
#   NAV_LOCAL=1 ./install.sh           # 强制用本地 dist/（开发期）
#   NAV_DEST=/usr/local/bin ./install.sh   # 自定义安装目录
#   NAV_NO_NCD=1 ./install.sh          # 跳过 ncd 函数自动配置
#   NAV_UNINSTALL=1 ./install.sh       # 卸载（删二进制 + 移除 ncd 函数）
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

uninstall() {
  for rc in "$HOME/.zshrc" "$HOME/.bashrc"; do
    [ -f "$rc" ] || continue
    if grep -qF "# >>> nav ncd >>>" "$rc"; then
      if [ "$(uname)" = "Darwin" ]; then
        sed -i.bak "/# >>> nav ncd >>>/,/# <<< nav ncd <<</d" "$rc" && rm -f "$rc.bak"
      else
        sed -i "/# >>> nav ncd >>>/,/# <<< nav ncd <<</d" "$rc"
      fi
      echo "已从 $rc 移除 ncd 函数"
    fi
  done
  if [ -f "$DEST/nav" ]; then
    rm -f "$DEST/nav"
    echo "已删除 $DEST/nav"
  fi
  echo "卸载完成"
}

if [ "${NAV_UNINSTALL:-0}" = "1" ]; then
  uninstall
  exit 0
fi

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

# ---- ncd 函数自动配置（zsh/bash，marker 标记可卸载） ----
NCD_FUNC='# >>> nav ncd >>>
# ncd: 用 nav 导航并 cd（⏎ 选中当前项→cd 过去；→ 深入；q 停在当前目录）
ncd() {
  local d
  d="$(command nav --print "$@")" || return $?
  if [[ -n "$d" ]]; then
    if [[ -d "$d" ]]; then
      builtin cd "$d"
    else
      printf '\''%s\n'\'' "$d"
    fi
  fi
}
# <<< nav ncd <<<'

setup_ncd() {
  [ "${NAV_NO_NCD:-0}" = "1" ] && return
  shell="${SHELL:-}"
  rc=""
  case "$shell" in
    *zsh*) rc="$HOME/.zshrc" ;;
    *bash*) rc="$HOME/.bashrc" ;;
    "")
      if [ -f "$HOME/.zshrc" ]; then rc="$HOME/.zshrc"
      elif [ -f "$HOME/.bashrc" ]; then rc="$HOME/.bashrc"
      else
        echo "无法检测 shell，请手动配置 ncd（见 README）"
        return
      fi ;;
    *)
      echo "未知 shell: $shell，请手动配置 ncd（见 README）"
      return ;;
  esac

  [ -f "$rc" ] || touch "$rc"
  if grep -qF "# >>> nav ncd >>>" "$rc"; then
    echo "ncd 已配置（${rc}），跳过"
    return
  fi
  printf '\n%s\n' "$NCD_FUNC" >> "$rc"
  echo "已把 ncd 函数写入 ${rc}（新开终端或 source ${rc} 生效）"
}
setup_ncd

echo ""
echo "安装完成。用法:"
echo "  nav     浏览/打开文件（⏎ = 打开，→ = 进入目录）"
echo "  ncd     导航并 cd（⏎ = 选中目录并切过去，→ = 深入，q = 停在当前目录）"
echo ""
"$DEST/nav" --version
