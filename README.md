# nav（Go 版）— 内联弹出式目录导航器

Python 版的跨平台重写：**单静态二进制、零运行时依赖、启动 ~4ms**（Python 版 ~70ms 的 18 倍）。
交互契约与 Python 版完全一致（UI 走 stderr、stdout 只输出结果路径、`--print` 语义相同），
已通过同一套 PTY 冒烟测试回归。

## 平台支持

| 平台 | 状态 | 打开文件方式 |
|---|---|---|
| macOS (arm64/amd64) | ✅ 完整支持（主战场） | `open` |
| Linux (arm64/amd64) | ✅ 代码已实现，交叉编译通过 | `xdg-open` |
| Windows (arm64/amd64) | ✅ 代码已实现，交叉编译通过（建议 Windows Terminal） | `cmd /c start` |
| 其他 POSIX (FreeBSD 等) | 提供 `term_other.go` 兜底 | `xdg-open` |

## 构建（六平台一次产出）

```bash
./build.sh        # 需要本机装有 Go（任意版本 ≥1.22）
# 产物: dist/nav-{darwin,linux,windows}-{amd64,arm64}(.exe) + dist/nav-native
```

## 安装（任意平台，一行）

```bash
# macOS / Linux
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh | sh

# Windows PowerShell
irm https://github.com/lodrg/nav/releases/latest/download/install.ps1 | iex

# 开发期本地安装（不依赖网络）
NAV_LOCAL=1 ./install.sh
```

安装脚本自动检测系统（darwin/linux × amd64/arm64）下载对应二进制到 `~/.local/bin`（Windows 为 `%LOCALAPPDATA%\bin`）。可覆盖：`NAV_URL=<前缀>` 换下载源，`NAV_DEST=<目录>` 换安装位置。Release 附 `SHA256SUMS` 校验文件。

安装后建议配置 shell 函数（目录跳转）：

```bash
# zsh / bash
ncd() {
  local d
  d="$(command nav --print "$@")" || return $?
  if [[ -n "$d" ]]; then
    if [[ -d "$d" ]]; then builtin cd "$d"; else print -r -- "$d"; fi
  fi
}
```

```powershell
# PowerShell
function ncd {
  $d = nav --print @args
  if ($LASTEXITCODE -eq 0 -and $d -and (Test-Path $d -PathType Container)) { Set-Location $d }
}
```

## 用法与按键

```
nav [路径]          弹出导航器（默认当前目录）
nav --print [路径]  选择后打印绝对路径并退出 → cd "$(nav --print)" / ncd
nav --once [路径]   打开一个文件后立即退出

↑/↓ 移动 · →/Enter 打开/进入 · ← 上级 · e 编辑器 · a 默认应用
空格 预览 · . 隐藏文件 · q/Esc 退出 · p (--print) 输出当前路径
```

`--print` 模式：`→` 深入目录，`⏎` 选中当前项（目录或文件）输出路径并退出，
`p`/`q`/`Ctrl+C` 输出当前目录并退出。

## 性能（实测，Apple M1 Pro）

| 指标 | Go (arm64 原生) | Python 版 |
|---|---|---|
| 启动（中位数，30 次） | **3.9 ms** | ~70 ms |
| 二进制大小 | **2.5 MB** | 脚本 + 解释器 |
| 运行时依赖 | 无（静态链接） | 需要 Python ≥3.9 |

## 架构

```
main.go            参数解析 + TTY 守卫 + 入口
app.go             App 状态机 + 主循环 + 打开策略（suspend/恢复终端）
fs.go              扫描/类型判定（dir|exe|text|media|unknown）/排序
render.go          宽度感知截断 + renderRegion 纯函数 + formatRow
keys.go            按键读取（ESC 三字节序列 + 100ms 超时 + OA 兼容）
term_darwin.go     终端层（cbreak: 清 ICANON|ECHO 保留 ISIG；TIOCGETA）
term_linux.go      终端层（TCGETS）
term_windows.go    终端层（ConsoleMode）
opener_*.go        默认应用打开（open / xdg-open / cmd start）
*_test.go          纯逻辑单测（go test ./...）
```

设计要点（与 Python 版共享）：

- **UI 走 stderr，stdout 只承载结果路径** —— `ncd`/`cd "$(...)"` 命令替换依赖
- **cbreak 而非 raw**：保留 ISIG（Ctrl+C 是正常 SIGINT）和 OPOST（\n 正常换行）
- **宽度感知截断**：自带 East Asian Width 宽字符表（CJK/Hangul/emoji），零依赖
- **单 goroutine 读键**：串行读 fd0，避免并发读同一 fd
- **信号处理**：SIGINT/SIGTERM 恢复终端；`--print` 模式输出当前目录

## 测试

```bash
go test ./...                          # 纯逻辑单测
NAV_BIN=./dist/nav-arm64-native python3 /tmp/nav_pty_test.py   # PTY 交互冒烟
python3 /tmp/nav_zsh_test.py           # zsh 全链路（ncd → cd → pwd）
```

## 与 Python 版的关系

- `~/Desktop/dev/nav/`（Python）—— 参考实现/离线兜底，不再更新
- `~/Desktop/dev/nav-go/`（Go）—— 主版本，建议所有新机器装这个
