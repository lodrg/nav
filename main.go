package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// version 构建时注入：go build -ldflags "-X main.version=<tag>"（见 build.sh / workflow）
var version = "dev"

const usageTemplate = `nav %s — 内联弹出式目录导航器（Go 版，单二进制）

用法:
  nav [路径]           弹出导航器（默认当前目录）
  nav --print [路径]   选择后打印绝对路径并退出 → cd "$(nav --print)" / ncd
                       （⏎ 选中当前项 · → 深入目录 · p/q 结束并输出当前目录）
  nav --once [路径]    打开一个文件后立即退出

按键: ↑/↓ 移动 · →/Enter 打开/进入 · ← 上级 · e 编辑器 · a 默认应用
      空格 预览 · . 隐藏文件 · q/Esc 退出 · p (--print) 输出当前路径
`

func main() {
	path := "."
	printMode, once := false, false
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
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "nav: 未知选项: %s\n（用 --help 查看用法）\n", arg)
			os.Exit(2)
		default:
			path = arg
		}
	}
	// 非 TTY 守卫
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		fmt.Fprintln(os.Stderr, "nav: 标准输入不是终端（TTY），无法交互")
		os.Exit(2)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nav: 无法解析路径: %s\n", path)
		os.Exit(2)
	}
	st, err := os.Stat(abs)
	if err != nil || !st.IsDir() {
		fmt.Fprintf(os.Stderr, "nav: 不是目录: %s\n", path)
		os.Exit(2)
	}

	go keyReaderLoop()
	app := &App{cwd: abs, printMode: printMode, once: once}
	os.Exit(app.run(int(os.Stdin.Fd())))
}
