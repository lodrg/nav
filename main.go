package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nav/internal/core"
)

var version = "dev" // 构建时注入：go build -ldflags "-X main.version=<tag>"（见 build.sh / workflow）

const usageTemplate = `nav %s — 内联弹出式目录导航器（Go 版，单二进制）

用法:
  nav [路径]           弹出导航器（默认当前目录）
  nav --print [路径]   选择后打印绝对路径并退出 → cd "$(nav --print)" / ncd
                       （⏎ 选中当前项 · → 深入目录 · p/q 结束并输出当前目录）
  nav --once [路径]    打开一个文件后立即退出
  nav --open <文件>    按类型直接打开文件（文本→编辑器，可执行→终端，其他→默认应用）

按键: ↑↓/jk 移动 · →/l/⏎ 打开 · ←/h 上级 · Tab 排序 · 空格 预览 · . 隐藏文件 · q/Esc 退出
`

func main() {
	path := "."
	printMode, once, openMode := false, false, false
	for _, arg := range os.Args[1:] {
		switch {
		case arg == "-h" || arg == "--help":
			fmt.Printf(usageTemplate, version)
			return
		case arg == "--version":
			fmt.Printf("nav %s\n", version)
			return
		case arg == "--print":
			printMode = true
		case arg == "--once":
			once = true
		case arg == "--open":
			openMode = true
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "nav: 未知选项: %s\n（用 --help 查看用法）\n", arg)
			os.Exit(2)
		default:
			path = arg
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nav: 无法解析路径: %s\n", path)
		os.Exit(2)
	}
	st, err := os.Stat(abs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nav: 路径不存在: %s\n", path)
		os.Exit(2)
	}

	fd := int(os.Stdin.Fd())
	if openMode {
		if st.IsDir() {
			fmt.Fprintln(os.Stderr, "nav: --open 需要文件路径（不是目录）")
			os.Exit(2)
		}
		os.Exit(core.OpenFileMode(filepath.Dir(abs), abs, fd))
	}

	// 非 TTY 守卫（交互模式）
	if fi, err := os.Stdin.Stat(); err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		fmt.Fprintln(os.Stderr, "nav: 标准输入不是终端（TTY），无法交互")
		os.Exit(2)
	}
	if !st.IsDir() {
		fmt.Fprintf(os.Stderr, "nav: 不是目录: %s\n", path)
		os.Exit(2)
	}

	go core.KeyReaderLoop()
	app := core.New(abs, printMode, once)
	os.Exit(app.Run(fd))
}
