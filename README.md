# nav — 终端里的文件导航器

在终端里敲一个命令，当前目录的文件列表**就在你眼前弹出**。
方向键（或 vim 的 `hjkl`）上下选择，回车打开文件、右键进入文件夹。
想切目录？`ncd` 选中即到，不用再手打 `cd`。

单文件、零依赖、启动只要 **4 毫秒**。macOS / Linux / Windows 都能用。

```
$ ncd            ← 在任意目录敲这一行

📂 ~/projects/dev/nav-go   (14 项)        ← 顶部：当前在哪、多少项
────────────────────────────────────────
📁 dist        —      —       2026-08-08 21:32
📄 app.go      —    7.3 KB    2026-08-08 22:10   ◄ 高亮行 = 选中项
📄 README.md   —    5.0 KB    2026-08-08 22:15
────────────────────────────────────────
↑↓/jk 移动 · →/l 进入 · ←/h 上级 · e 编辑 · a 应用 · 空格 预览 · q 退出

$              ← 按 q 退出，终端干干净净，什么都没留下
```

## 一分钟上手

```bash
# macOS / Linux（一条命令装好，自动配置）
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh | sh

# Windows PowerShell
irm https://github.com/lodrg/nav/releases/latest/download/install.ps1 | iex
```

装完**新开一个终端**（或 `source ~/.zshrc`），就有了两个命令：

| 命令 | 干什么 | 什么时候用 |
|---|---|---|
| `nav` | 弹出文件列表，浏览、打开文件 | 在**当前**目录里找文件打开 |
| `ncd` | 同上，但选中目录后 **shell 直接切过去** | 想去**别的**目录 |

> 安装脚本会自动把 `ncd` 写进你的 shell 配置（zsh → `~/.zshrc`，bash → `~/.bashrc`，PowerShell → `$PROFILE`），不用手动配。

## 打开文件：每种类型自动用对应方式

不需要指定，nav 会根据文件类型**自动选择**最合适的打开方式：

| 文件类型 | 例子 | 自动打开方式 |
|---|---|---|
| 文本 / 代码 | `.py` `.md` `.json` `.txt` `.go` `.yml` … | **编辑器**（`$EDITOR`，默认 vim） |
| 图片 / 文档 / 视频 / 音频 | `.png` `.jpg` `.pdf` `.mp4` `.docx` `.zip` … | **系统默认应用**（看图、播放器、Office…） |
| 可执行文件 | 有执行权限的脚本/程序（`./run.sh`） | **在终端里运行**，结束按任意键返回 |
| 未知类型 | 没有扩展名的二进制 | 系统默认应用兜底 |

打开方式**完全自动**，不需要任何额外按键——`→` / `l` / `回车` 三个键就能打开一切。
想先看看内容再决定？按 `空格` 快速预览（再按一次返回列表）。

> 无扩展名的文本文件也能识别：自动做二进制嗅探，是纯文本就当文本处理。
> 如果文本文件打不开，说明系统里没有任何编辑器（已自动尝试 `$VISUAL` → `$EDITOR` → vim → vi → nano → code），安装一个或设置 `export EDITOR=你的编辑器` 即可。

## 按键速查

| 按键 | 动作 |
|---|---|
| `↑` `↓` / `k` `j` | 移动选中行（首尾循环） |
| `→` `l` / `回车` | 目录→进入；文件→自动选择方式打开 |
| `←` `h` | 返回上级（**记住上次光标位置**，回到原处） |
| `空格` | 预览文本内容 |
| `.` | 显示/隐藏隐藏文件 |
| `q` `Esc` `Ctrl+C` | 退出 |

`ncd`（跳转目录）模式下：`→` `l` = 深入目录，`回车` = **智能处理当前项**（目录→shell 切过去，文件→按类型打开：文本→编辑器、可执行→终端、其他→默认应用），`p` 或 `q` = 停在当前目录并输出。

## 升级 / 卸载

```bash
# 升级：重跑安装命令即可（自动覆盖旧版）
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh | sh

# 卸载：删二进制 + 移除 ncd 函数
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh -o /tmp/nav-install.sh
NAV_UNINSTALL=1 sh /tmp/nav-install.sh
```

## 平台支持

| 平台 | 状态 | 打开文件方式 |
|---|---|---|
| macOS (arm64/amd64) | ✅ 完整支持 | `open` |
| Linux (arm64/amd64) | ✅ 完整支持 | `xdg-open` |
| Windows (arm64/amd64) | ✅ 支持（建议 Windows Terminal） | `cmd /c start` |

## 给开发者

```bash
./build.sh          # 六平台一次产出（需要 Go ≥1.22）
go test ./...       # 纯逻辑单测
```

```
main.go            参数解析 + TTY 守卫 + 入口
app.go             App 状态机 + 主循环 + 打开策略（suspend/恢复终端）
fs.go              扫描/类型判定（dir|exe|text|media|unknown）/排序
render.go          宽度感知截断 + renderRegion 纯函数 + formatRow
keys.go            按键读取（ESC 三字节序列 + 100ms 超时）
term_*.go          终端层（cbreak：darwin/linux/windows 分平台）
opener_*.go        默认应用打开（open / xdg-open / cmd start）
```

设计要点：UI 走 stderr、stdout 只输出结果路径（`ncd` 命令替换依赖）；cbreak 而非 raw（Ctrl+C 保持正常）；CJK/emoji 宽度感知截断，中文文件名不撕裂。
