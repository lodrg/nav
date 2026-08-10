<div align="center">

# nav

在终端里弹出当前目录的文件列表，用方向键或 hjkl 浏览，回车打开文件，`ncd` 一键跳转目录。

[English](README.en.md) · 简体中文

![version](https://img.shields.io/github/v/tag/lodrg/nav?label=version&color=blue)
![platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)
![size](https://img.shields.io/badge/size-1.6MB-brightgreen)
![downloads](https://img.shields.io/github/downloads/lodrg/nav/total?color=green)

</div>

- 单静态二进制，约 1.6 MB，零运行时依赖
- macOS / Linux / Windows
- 打开文件时按类型自动选择方式（编辑器 / 默认应用 / 终端运行）
- Tab 切换排序（名称 / 时间 / 大小）
- 弹层顶部实时显示当前路径

## 安装

macOS / Linux：

```bash
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh | sh
```

Windows PowerShell：

```powershell
irm https://github.com/lodrg/nav/releases/latest/download/install.ps1 | iex
```

安装脚本自动完成：下载对应平台的二进制到 `~/.local/bin`（Windows 为 `%LOCALAPPDATA%\bin`）、加入 PATH、把 `ncd` 函数写进 shell 配置（zsh 写 `~/.zshrc`，bash 写 `~/.bashrc`），并启用 `nav` / `ncd` 的 Tab 路径补全。新开终端后即可使用。

升级：重跑安装命令。卸载：

```bash
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh -o /tmp/nav-install.sh
NAV_UNINSTALL=1 sh /tmp/nav-install.sh
```

## 用法

| 命令 | 作用 |
|---|---|
| `nav` | 浏览当前目录：打开文件、进入子目录 |
| `ncd` | 浏览并切换：选中目录后 shell 直接 cd 过去；选中文件则按类型打开 |

两者共用同一套界面，区别只在退出语义：`nav` 用于在当前目录找文件并打开（回车打开后弹层不关闭），`ncd` 用于去别的目录（回车选中后退出，路径交给 shell）。详见[常见问题](#常见问题)。

### 打开文件

不指定方式，nav 按文件类型自动选择：

| 类型 | 例子 | 打开方式 |
|---|---|---|
| 文本 / 代码 | `.py` `.md` `.json` `.txt` | `$EDITOR`（回退 vim → vi → nano） |
| 图片 / 文档 / 媒体 | `.png` `.pdf` `.mp4` `.docx` | 系统默认应用 |
| 可执行文件 | 有执行权限的脚本、程序 | 在终端中运行 |
| 其他 | 无扩展名的二进制 | 默认应用 |

无扩展名的文本文件通过二进制嗅探识别。

### 按键

| 按键 | 作用 |
|---|---|
| `↑` `↓` / `k` `j` | 移动（首尾循环） |
| `→` `l` `⏎` | 目录：进入；文件：打开 |
| `←` `h` | 返回上级（记住光标位置） |
| `Tab` | 排序：名称 / 时间（新在前）/ 大小（大在前） |
| `空格` | 预览文本内容 |
| `.` | 显示 / 隐藏隐藏文件 |
| `q` `Esc` `Ctrl+C` | 退出 |

`ncd` 模式下：`→` `l` 进入目录，`⏎` 选中（目录 → cd，文件 → 打开），`p` / `q` 输出当前目录。

## 性能

Apple M1 Pro 实测（中位数，测试方法见 [PERF.md](PERF.md)）：

| 场景 | nav (Go) | Python 参考版 |
|---|---|---|
| 进程启动 | 19.7 ms | 42.2 ms |
| 冷启动，普通目录 | 35.9 ms | 56.4 ms |
| 冷启动，2050 项目录 | 43.7 ms | 73.6 ms |

单进程、无守护、按键响应即时。作为对比，ranger 冷启动约 300–500 ms。

## 平台

| 平台 | 架构 | 打开文件方式 |
|---|---|---|
| macOS | arm64 / amd64 | `open` |
| Linux | arm64 / amd64 | `xdg-open` |
| Windows | arm64 / amd64 | `cmd /c start`（建议 Windows Terminal） |

## 常见问题

**为什么有两个命令？**

子进程无法改变父 shell 的工作目录，所以"切换目录"必须由 shell 函数完成：`ncd` 运行 `nav --print` 拿到选中的路径，再执行 `cd`。`nav` 负责浏览和打开，`ncd` 负责跳转，两者的回车语义不同。

**Linux 上文本文件打不开？**

文本走编辑器回退链 `$VISUAL` → `$EDITOR` → vim → vi → nano → code。系统里没有任何编辑器时会给出明确提示，不会静默失败；装一个或设置 `export EDITOR=...` 即可。

**`ncd` 对文件做了什么？**

目录 → `cd` 过去；文件 → `nav --open` 按类型打开。纯脚本场景可直接用 `nav --print` 拿路径。

## 开发

```bash
./build.sh          # 交叉编译六平台
go test ./...       # 单测
```

代码结构：

```
main.go             入口与参数解析（package main）
internal/fs/        扫描 / 类型判定 / 排序（名称·时间·大小）
internal/render/    宽度感知渲染（纯函数，可单测）
internal/core/      App 状态机 / 按键处理 / 打开策略
internal/term/      终端层（cbreak：darwin/linux/windows 分平台）
internal/opener/    跨平台打开（open / xdg-open / cmd start）
```

设计要点：UI 输出到 stderr，stdout 只输出结果路径（`ncd` 依赖此契约）；CJK / emoji 宽度感知截断。
