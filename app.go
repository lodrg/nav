package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

const (
	footerNormal = "↑↓/jk 移动 · →/l 进入 · ←/h 上级 · e 编辑 · a 应用 · 空格 预览 · q 退出"
	footerPrint  = "↑↓/jk 移动 · →/l 深入 · ⏎ 选中 · p/q 输出 · h/← 上级 · Esc 退出"
)

type App struct {
	cwd        string
	printMode  bool
	once       bool
	entries    []Entry
	err        error
	cursor     int
	top        int
	showHidden bool
	lastHeight int
	preview    []string
	previewOn  bool
	selected   string
	oldState   any
	fd         int
}

func (a *App) load() {
	items, err := scan(a.cwd, a.showHidden)
	a.entries, a.err = items, err
	if len(a.entries) == 0 {
		a.cursor = 0
	} else if a.cursor > len(a.entries)-1 {
		a.cursor = len(a.entries) - 1
	}
	a.top = min(a.top, a.cursor)
}

func (a *App) loadPreview(e Entry) {
	if e.Kind != "text" {
		a.preview = []string{"（非文本文件，按 → 打开 · a 用默认应用 · 空格 返回）"}
		a.previewOn = true
		return
	}
	full := filepath.Join(a.cwd, e.Name)
	data, err := os.ReadFile(full)
	if err != nil {
		a.preview = []string{fmt.Sprintf("（无法读取：%v）", err)}
		a.previewOn = true
		return
	}
	lines := strings.Split(string(data), "\n")
	// 去掉尾随换行产生的空串（与 Python readlines 行为一致）
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > 200 {
		lines = lines[:200]
	}
	a.preview = lines
	a.previewOn = true
}

// formatPath：弹层顶部当前目录行（加粗 + 📂，HOME 缩写为 ~，超宽保留尾部）。
func (a *App) formatPath(width int) string {
	p := a.cwd
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		switch {
		case p == home:
			p = "~"
		case strings.HasPrefix(p, home+string(os.PathSeparator)):
			p = "~" + p[len(home):]
		}
	}
	// 宽度预算：图标(2) + 空格(1) + 条目数后缀
	suffix := fmt.Sprintf(" (%d 项)", len(a.entries))
	avail := width - 4 - dispWidth(suffix)
	if avail < 8 {
		avail = 8
	}
	return "\x1b[1m" + "📂 " + truncateTail(p, avail) + "\x1b[0m" + dim + suffix + reset
}

func (a *App) draw(first bool) {
	width, termH := termSize(a.fd)
	maxH := max(4, termH-2) // 弹层最多占 termH-2 行（路径行 + 内容 + footer）
	n := len(a.entries)

	body := []string{a.formatPath(width)} // 第 0 行：当前目录
	hl := -1
	if a.previewOn {
		for _, ln := range a.preview {
			body = append(body, truncate(ln, width))
		}
	} else if n == 0 {
		msg := "（空目录）"
		if a.err != nil {
			msg = a.err.Error()
		}
		body = append(body, red+truncate(msg, width)+reset)
	} else {
		entryVis := min(n, maxH-2) // 路径行 + footer 占两行
		a.top = min(a.top, a.cursor)
		if a.cursor >= a.top+entryVis {
			a.top = a.cursor - entryVis + 1
		}
		a.top = max(0, min(a.top, n-entryVis))
		for i := a.top; i < a.top+entryVis; i++ {
			body = append(body, formatRow(a.entries[i], width))
		}
		hl = 1 + a.cursor - a.top // 路径行占 index 0
	}

	ftr := footerNormal
	if a.printMode {
		ftr = footerPrint
	}
	lines := append(body[:min(len(body), maxH-1)], dim+truncate(ftr, width)+reset)
	height := len(lines)

	var hlPtr *int
	if hl >= 0 {
		hlPtr = &hl
	}
	out, newH := renderRegion(lines, height, width, a.lastHeight, hlPtr)
	a.lastHeight = newH
	if first {
		out = "\n" + out // 首次绘制：先离开提示符所在行
	}
	ui(out)
}

func (a *App) erasePopup() {
	if a.lastHeight > 0 {
		var b strings.Builder
		b.WriteString(up(a.lastHeight))
		for i := 0; i < a.lastHeight; i++ {
			b.WriteString(clearLn + "\n")
		}
		b.WriteString(up(a.lastHeight))
		ui(b.String())
	}
	a.lastHeight = 0
}

// ---- 打开策略 ----

func (a *App) openEntry(idx int, method string) bool {
	e := a.entries[idx]
	full := filepath.Join(a.cwd, e.Name)
	if _, err := os.Lstat(full); err != nil {
		a.load()
		return false
	}
	if e.IsDir {
		if method == "default" { // a 键在目录上 → Finder/文件管理器
			openDefault(full)
			return true
		}
		a.cwd = full
		a.load()
		a.cursor, a.top = 0, 0
		return false
	}
	if (method == "auto" || method == "editor") && (e.Kind == "text" || method == "editor") {
		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
		if editor == "" {
			if runtime.GOOS == "windows" {
				editor = "notepad"
			} else {
				editor = "vim"
			}
		}
		a.suspendRun([]string{editor, full}, false)
		return true
	}
	if method == "auto" && e.Kind == "exe" {
		a.suspendRun([]string{full}, true)
		return true
	}
	openDefault(full)
	return true
}

func (a *App) suspendRun(argv []string, waitKey bool) {
	a.erasePopup()
	restoreTerm(a.fd, a.oldState)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	rc := 0
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			rc = ee.ExitCode()
		}
	}
	a.oldState, _ = makeCbreak(a.fd)
	if waitKey {
		msg := "按任意键返回…"
		if rc != 0 {
			msg = fmt.Sprintf("（退出码 %d）按任意键返回…", rc)
		}
		ui(msg)
		buf := make([]byte, 1)
		os.Stdin.Read(buf)
	}
	a.lastHeight = 0
}

// ---- 按键处理（纯逻辑，可单测） ----

// handleKey 返回动作: redraw | quit | print_path | opened_file
func (a *App) handleKey(key string) string {
	n := len(a.entries)
	switch key {
	case "q", "esc", "ctrl-c":
		return "quit"
	case "up", "k":
		if n > 0 {
			a.cursor = (a.cursor - 1 + n) % n
		}
		a.previewOn = false
		return "redraw"
	case "down", "j":
		if n > 0 {
			a.cursor = (a.cursor + 1) % n
		}
		a.previewOn = false
		return "redraw"
	case "left", "h":
		parent := filepath.Dir(a.cwd)
		if parent != a.cwd {
			a.cwd = parent
			a.load()
			a.cursor, a.top = 0, 0
		}
		a.previewOn = false
		return "redraw"
	case "right", "l":
		if n == 0 {
			return "redraw"
		}
		if a.printMode {
			// --print 模式：→ 只负责深入目录（⏎ 才是选中）
			if a.entries[a.cursor].IsDir {
				a.cwd = filepath.Join(a.cwd, a.entries[a.cursor].Name)
				a.load()
				a.cursor, a.top = 0, 0
			}
			return "redraw"
		}
		a.previewOn = false
		if a.openEntry(a.cursor, "auto") {
			return "opened_file"
		}
		return "redraw"
	case "enter":
		if n == 0 {
			return "redraw"
		}
		if a.printMode {
			// --print 模式：⏎ 选中当前项（目录或文件），输出其绝对路径并退出
			a.selected = filepath.Join(a.cwd, a.entries[a.cursor].Name)
			return "print_path"
		}
		a.previewOn = false
		if a.openEntry(a.cursor, "auto") {
			return "opened_file"
		}
		return "redraw"
	case ".":
		a.showHidden = !a.showHidden
		a.load()
		return "redraw"
	case " ":
		if a.previewOn {
			a.previewOn = false
		} else if n > 0 {
			a.loadPreview(a.entries[a.cursor])
		}
		return "redraw"
	case "e":
		if n > 0 {
			a.previewOn = false
			a.openEntry(a.cursor, "editor")
			return "opened_file"
		}
		return "redraw"
	case "a":
		if n > 0 {
			a.previewOn = false
			a.openEntry(a.cursor, "default")
			return "opened_file"
		}
		return "redraw"
	case "p":
		if a.printMode {
			a.selected = a.cwd
			return "print_path"
		}
	}
	return "redraw"
}

// ---- 主循环 ----

func (a *App) run(fd int) int {
	a.fd = fd
	old, err := makeCbreak(fd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "nav: 无法进入终端原始模式:", err)
		return 2
	}
	a.oldState = old
	defer restoreTerm(fd, a.oldState)

	code := 0
	a.load()
	a.draw(true)

	// 信号：恢复终端；--print 模式输出当前目录（Ctrl+C 语义与 Python 版一致）
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case s := <-sigCh:
				restoreTerm(fd, a.oldState)
				if a.printMode {
					fmt.Println(a.cwd)
				}
				switch s {
				case os.Interrupt:
					os.Exit(130)
				default:
					os.Exit(143)
				}
			case <-done:
				return
			}
		}
	}()
	defer close(done)

loop:
	for {
		key := readKey()
		switch a.handleKey(key) {
		case "quit":
			break loop
		case "print_path":
			break loop
		case "opened_file":
			if a.once {
				break loop
			}
		}
		a.draw(false)
	}
	a.erasePopup()
	// --print 模式：任何方式退出都输出一个路径（⏎ 选中项 或 当前目录）
	if a.printMode && a.selected == "" {
		a.selected = a.cwd
	}
	if a.selected != "" {
		fmt.Println(a.selected)
	}
	return code
}
