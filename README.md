<div align="center">

# 📂 nav

**终端里的文件导航器** —— 方向键浏览目录，回车打开文件，`ncd` 一键跳转

[简体中文](README.md) · [English](README.en.md)

![version](https://img.shields.io/github/v/tag/lodrg/nav?label=version&color=blue)
![platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)
![size](https://img.shields.io/badge/size-1.6MB-brightgreen)
![downloads](https://img.shields.io/github/downloads/lodrg/nav/total?color=green)

</div>

---

## ✨ 特性

| | |
|---|---|
| ⚡ **即时启动** | 单静态二进制，4 毫秒级启动，零运行时依赖 |
| ⌨️ **纯键盘操作** | 方向键 + vim `hjkl`，手不离键 |
| 📂 **路径常显** | 顶部实时显示当前目录与条目数 |
| 🗂️ **一键排序** | `Tab` 切换：名称 → 时间 → 大小 |
| 🧭 **选中即达** | `ncd` 选中目录后 shell 直接切过去 |
| 🖼️ **智能打开** | 按文件类型自动选择编辑器 / 默认应用 / 终端运行 |

## 🚀 一分钟上手

```bash
# macOS / Linux（一条命令装好，自动完成全部配置）
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh | sh

# Windows PowerShell
irm https://github.com/lodrg/nav/releases/latest/download/install.ps1 | iex
```

装完**新开一个终端**，就有两个命令：

| 命令 | 干什么 | 什么时候用 |
|---|---|---|
| `nav` | 弹出文件列表，浏览、打开文件 | 在**当前**目录里找文件打开 |
| `ncd` | 同上，但选中目录后 **shell 直接切过去** | 想去**别的**目录 |

> 安装脚本自动完成：下载对应平台二进制到 `~/.local/bin` → **加入 PATH** → 把 `ncd` 函数写入你的 shell 配置（zsh → `~/.zshrc`，bash → `~/.bashrc`，PowerShell → `$PROFILE`）→ 启用 `nav`/`ncd` 的 Tab 路径补全。

## 🖥️ 长什么样

```
$ ncd            ← 在任意目录敲这一行

📂 ~/projects/dev/nav-go   (14 项 · 名称)       ← 顶部：当前路径 / 条目数 / 排序方式
────────────────────────────────────────────
📁 dist        —      —       2026-08-08 21:32
📄 app.go      —    7.3 KB    2026-08-08 22:10   ◄ 高亮行 = 选中项
📄 README.md   —    5.0 KB    2026-08-08 22:15
────────────────────────────────────────────
↑↓/jk 移动 · →/l ⏎ 打开 · ←/h 上级 · Tab 排序 · 空格 预览 · q 退出

$              ← 按 q 退出，终端干干净净
```

## 📖 使用指南

### 打开文件：每种类型自动用对应方式

不需要指定，nav 根据文件类型**自动选择**最合适的打开方式：

| 文件类型 | 例子 | 自动打开方式 |
|---|---|---|
| 文本 / 代码 | `.py` `.md` `.json` `.txt` `.go` `.yml` … | **编辑器**（`$EDITOR`，自动回退 vim → vi → nano） |
| 图片 / 文档 / 视频 / 音频 | `.png` `.jpg` `.pdf` `.mp4` `.docx` `.zip` … | **系统默认应用** |
| 可执行文件 | 有执行权限的脚本/程序（`./run.sh`） | **在终端里运行**，结束按任意键返回 |
| 未知类型 | 没有扩展名的二进制 | 默认应用兜底 |

打开方式**完全自动**——`→` / `l` / `回车` 三个键打开一切。想先看内容？`空格` 快速预览，再按一次返回。

### 按键速查

| 按键 | 动作 |
|---|---|
| `↑` `↓` / `k` `j` | 移动选中行（首尾循环） |
| `→` `l` / `回车` | 目录→进入；文件→自动选择方式打开 |
| `←` `h` | 返回上级（**记住上次光标位置**） |
| `Tab` | 切换排序：名称 → 时间（新在前）→ 大小（大在前） |
| `空格` | 预览文本内容 |
| `.` | 显示 / 隐藏隐藏文件 |
| `q` `Esc` `Ctrl+C` | 退出 |

### `ncd` 模式的差异

`→` / `l` = 深入目录；`回车` = **智能处理当前项**（目录 → shell 切过去；文件 → 按类型打开）；`p` / `q` = 停在当前目录并输出路径。

> 💡 为什么有两个命令？`cd` 只能由 shell 自己执行（子进程无法改变父 shell 的目录），所以"跳转"必须是 shell 函数包一层。详见 [常见问题](#-常见问题)。

## 🤔 常见问题

**为什么有 `nav` 和 `ncd` 两个命令？**

这是 Unix 进程模型决定的：`nav` 是独立进程，它执行的任何 `cd` 都只影响自己，**无法改变父 shell 的工作目录**。所以"选中目录后切过去"必须由 shell 函数完成：

```
ncd  →  nav --print（弹层导航，选中后打印路径）→  ncd 收到路径 →  builtin cd
```

`nav` 负责"看和打开"，`ncd` 负责"走过去"，一个进程模型上的物理限制分出了两个命令。

**为什么 `ncd` 选中文件会打开而不是输出路径？**

v0.2.8 起 `ncd` 做了智能分发：目录 → `cd` 过去；文件 → `nav --open` 按类型打开（文本→编辑器、可执行→终端、其他→默认应用）。纯脚本场景仍可用 `nav --print` 直接拿路径。

**Linux 上文本文件打不开？**

文本类走编辑器回退链 `$VISUAL` → `$EDITOR` → vim → vi → nano → code。服务器没装任何编辑器时会明确提示，不会静默失败。装一个 `nano` 或设置 `export EDITOR=你的编辑器` 即可。

## ⚡ 性能

单静态二进制、零依赖，启动即点即出（Apple M1 Pro 实测）：

| 场景（中位数） | nav (Go) | Python 参考版 |
|---|---|---|
| 纯进程启动 | **19.7 ms** | 42.2 ms |
| 完整冷启动 · 普通目录 | **35.9 ms** | 56.4 ms |
| 完整冷启动 · 2050 项大目录 | **43.7 ms** | 73.6 ms |
| 二进制体积 | **1.61 MB** | 脚本 + 解释器 |

比 ranger（300–500 ms）快一个数量级，与 lf / nnn 同为编译型轻量派。
完整报告（方法学、拆解分析）：[PERF.md](PERF.md)

## 🔧 安装 / 升级 / 卸载

```bash
# 升级：重跑安装命令即可（自动覆盖旧版 + 更新 ncd 函数）
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh | sh

# 卸载：删二进制 + 移除 ncd 函数 + 还原 PATH
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh -o /tmp/nav-install.sh
NAV_UNINSTALL=1 sh /tmp/nav-install.sh
```

可覆盖的环境变量：`NAV_URL=<前缀>` 换下载源 · `NAV_DEST=<目录>` 换安装位置 · `NAV_NO_NCD=1` 跳过 shell 配置。

## 🗺️ 平台支持

| 平台 | 状态 | 打开文件方式 |
|---|---|---|
| macOS (arm64/amd64) | ✅ 完整支持 | `open` |
| Linux (arm64/amd64) | ✅ 完整支持 | `xdg-open` |
| Windows (arm64/amd64) | ✅ 支持（建议 Windows Terminal） | `cmd /c start` |

## 🛠️ 给开发者

```bash
./build.sh          # 六平台一次产出（需要 Go ≥ 1.22）
go test ./...       # 纯逻辑单测
```

```
main.go            参数解析 + TTY 守卫 + 入口
app.go             App 状态机 + 主循环 + 打开策略（suspend/恢复终端）
fs.go              扫描 / 类型判定 / 排序（名称·时间·大小）
render.go          宽度感知截断 + renderRegion 纯函数 + formatRow
keys.go            按键读取（ESC 三字节序列 + 100ms 超时）
term_*.go          终端层（cbreak：darwin/linux/windows 分平台）
opener_*.go        默认应用打开（open / xdg-open / cmd start）
```

**设计要点**：UI 走 stderr、stdout 只输出结果路径（`ncd` 命令替换依赖此契约）；cbreak 而非 raw（Ctrl+C 保持正常）；CJK / emoji 宽度感知截断，中文文件名不撕裂；单 goroutine 串行读键。

---

<div align="center">

⭐ 喜欢就用它，好用就把它推给需要的朋友

</div>
