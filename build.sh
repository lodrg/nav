#!/bin/sh
# nav 六平台交叉编译（在任意一台装好 Go 的机器上执行）
# 产物: dist/nav-<os>-<arch>（windows 为 .exe），另附本机原生二进制
set -e
cd "$(dirname "$0")"

mkdir -p dist
for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64; do
  os=${target%/*}
  arch=${target#*/}
  out="dist/nav-${os}-${arch}"
  [ "$os" = "windows" ] && out="$out.exe"
  echo "== $target → $out =="
  GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-s -w" -o "$out" .
done

echo "== 本机原生 =="
go build -trimpath -ldflags "-s -w" -o dist/nav-native .
ls -lh dist/ | grep -v "^total"
echo "完成。"
