#!/bin/sh
# nav 六平台交叉编译（在任意一台装好 Go 的机器上执行）
# 用法: ./build.sh [版本号]   # 默认取 git tag/describe，无 git 则 dev
# 产物: dist/nav-<os>-<arch>（windows 为 .exe），另附本机原生二进制
set -e
cd "$(dirname "$0")"

VERSION="${1:-$(git describe --tags --always 2>/dev/null || echo dev)}"
echo "版本: $VERSION"

mkdir -p dist
for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64; do
  os=${target%/*}
  arch=${target#*/}
  out="dist/nav-${os}-${arch}"
  [ "$os" = "windows" ] && out="$out.exe"
  echo "== $target → $out =="
  GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o "$out" .
done

echo "== 本机原生 =="
go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o dist/nav-native .
ls -lh dist/ | grep -v "^total"
echo "完成。"
