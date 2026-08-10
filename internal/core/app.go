// Package core App 状态机：主循环、按键处理、打开策略。
package core

import (
	"fmt"
	iofs "io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"nav/internal/fs"
	"nav/internal/opener"
	"nav/internal/render"
	"nav/internal/term"
)

const (
	footerNormal = "↑↓/jk 移动 · →/l ⏎ 打开 · ←/h 上级 · Tab 排序 · 空格 预览 · q 退出"
	footerPrint  = "↑↓/jk 移动 · →/l 深入 · ⏎ 选中 · p/q 输出 · h/← 上级 · Esc 退出"
)

type historyEntry struct {
	cwd    string
	cursor int
	top    int
}

// App 是导航器的应用状态。
type App struct {
	cwd        string
	printMode  bool
	once       bool
	entries    []fs.Entry
	err        error
	cursor     int
	top        int
	showHidden bool
	sortMode   int // 0 名称 / 1 时间 / 2 大小（Tab 循环）
	lastHeight int
	preview    []string
	previewOn  bool
	selected   string
	oldState   any
	fd         int
	hist       []historyEntry // 浏览历史栈：进入目录前的位置，h/← 返回时恢复
}

// New 创建 App。
func New(cwd string, printMode, once bool) *App {
	return &App{cwd: cwd, printMode: printMode, once: once}
}

// SetFD 设置终端 fd（--open 模式等非 Run 路径使用）。
func (a *App) SetFD(fd int) { a.fd = fd }

// enterDir 进入目录：先记住当前位置（返回时可恢复），再切换并从头浏览。
func (a *App) enterDir(full string) {
	a.hist = append(a.hist, historyEntry{a.cwd, a.cursor, a.top})
	if len(a.hist) > 100 {
		a.hist = a.hist[1:] // 限制栈深
	}
	a.cwd = full
	a.load()
	a.cursor, a.top = 0, 0
}

func (a *App) load() {
	a.entries, a.err = fs.Scan(a.cwd, a.showHidden)
	if a.err == nil {
		fs.SortEntries(a.entries, a.sortMode)
	}
	if len(a.entries) == 0 {
		a.cursor = 0
	} else if a.cursor > len(a.entries)-1 {
		a.cursor = len(a.entries) - 1
	}
	a.top = min(a.top, a.cursor)
}

func (a *App) loadPreview(e fs.Entry) {
	if e.Kind != "text" {
		a.preview = []string{"（非文本文件，按 → 打开 · 空格 返回）"}
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
	// 宽度预算：图标(2) + 空格(1) + 条目数+排序后缀
	suffix := fmt.Sprintf(" (%d 项 · %s)", len(a.entries), fs.SortLabels[a.sortMode])
	avail := width - 4 - render.DispWidth(suffix)
	if avail < 8 {
		avail = 8
	}
	return "\x1b[1m" + "📂 " + render.TruncateTail(p, avail) + "\x1b[0m" + render.Dim + suffix + render.Reset
}

func (a *App) draw(first bool) {
	width, termH := term.Size(a.fd)
	maxH := max(4, termH-2) // 弹层最多占 termH-2 行（路径行 + 内容 + footer）
	n := len(a.entries)

	body := []string{a.formatPath(width)} // 第 0 行：当前目录
	hl := -1
	if a.previewOn {
		for _, ln := range a.preview {
			body = append(body, render.Truncate(ln, width))
		}
	} else if n == 0 {
		msg := "（空目录）"
		if a.err != nil {
			msg = a.err.Error()
		}
		body = append(body, render.Red+render.Truncate(msg, width)+render.Reset)
	} else {
		entryVis := min(n, maxH-2) // 路径行 + footer 占两行
		a.top = min(a.top, a.cursor)
		if a.cursor >= a.top+entryVis {
			a.top = a.cursor - entryVis + 1
		}
		a.top = max(0, min(a.top, n-entryVis))
		for i := a.top; i < a.top+entryVis; i++ {
			body = append(body, render.FormatRow(a.entries[i], width))
		}
		hl = 1 + a.cursor - a.top // 路径行占 index 0
	}

	ftr := footerNormal
	if a.printMode {
		ftr = footerPrint
	}
	lines := append(body[:min(len(body), maxH-1)], render.Dim+render.Truncate(ftr, width)+render.Reset)
	height := len(lines)

	var hlPtr *int
	if hl >= 0 {
		hlPtr = &hl
	}
	out, newH := render.RenderRegion(lines, height, width, a.lastHeight, hlPtr)
	a.lastHeight = newH
	if first {
		out = "\n" + out // 首次绘制：先离开提示符所在行
	}
	render.UI(out)
}

func (a *App) erasePopup() {
	if a.lastHeight > 0 {
		var b strings.Builder
		b.WriteString(render.Up(a.lastHeight))
		for i := 0; i < a.lastHeight; i++ {
			b.WriteString(render.ClearLn + "\n")
		}
		b.WriteString(render.Up(a.lastHeight))
		render.UI(b.String())
	}
	a.lastHeight = 0
}

// ---- 打开策略 ----

// pickEditor 编辑器回退链：$VISUAL → $EDITOR → vim → vi → nano → code → notepad
func pickEditor() string {
	var candidates []string
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if v := os.Getenv(env); v != "" {
			candidates = append(candidates, v)
		}
	}
	candidates = append(candidates, "vim", "vi", "nano", "code", "notepad")
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			return c
		}
	}
	return ""
}

// showError 在弹层位置显示错误并等待按键（不静默失败）。
func (a *App) showError(msg string) {
	a.erasePopup()
	render.UI(msg + "\n按任意键返回…")
	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	a.lastHeight = 0
}

// fileInfoEntry：把 os.FileInfo 适配成 os.DirEntry，复用 fs.Classify 类型判定。
type fileInfoEntry struct {
	os.FileInfo
}

func (e fileInfoEntry) Type() iofs.FileMode        { return e.FileInfo.Mode().Type() }
func (e fileInfoEntry) Info() (os.FileInfo, error) { return e.FileInfo, nil }

// OpenFileMode：--open 子命令入口（ncd 复用）。按类型打开文件并返回退出码。
func OpenFileMode(cwd, full string, fd int) int {
	app := New(cwd, false, false)
	app.fd = fd
	if old, err := term.Cbreak(fd); err == nil {
		app.oldState = old
		defer term.Restore(fd, old)
	} // 非 TTY（管道调用）时降级：不进入 cbreak，直接运行
	if !app.openFile(full) {
		return 1
	}
	return 0
}

// openFile：按类型直接打开文件（--open 子命令）。
// 文本→编辑器（回退链）；可执行→终端运行；其他→默认应用。
func (a *App) openFile(full string) bool {
	info, err := os.Stat(full)
	if err != nil {
		a.showError(fmt.Sprintf("无法打开 %s: %v", full, err))
		return false
	}
	if info.IsDir() {
		opener.Open(full)
		return true
	}
	kind := fs.Classify(filepath.Dir(full), fileInfoEntry{info})
	switch kind {
	case "text":
		editor := pickEditor()
		if editor == "" {
			a.showError("未找到编辑器（已尝试 vim/vi/nano/code），请设置 $EDITOR 环境变量")
			return false
		}
		a.suspendRun([]string{editor, full}, false)
		return true
	case "exe":
		a.suspendRun([]string{full}, true)
		return true
	default:
		opener.Open(full)
		return true
	}
}

func (a *App) openEntry(idx int, method string) bool {
	e := a.entries[idx]
	full := filepath.Join(a.cwd, e.Name)
	if _, err := os.Lstat(full); err != nil {
		a.load()
		return false
	}
	if e.IsDir {
		if method == "default" { // 目录用文件管理器打开（预留 API）
			opener.Open(full)
			return true
		}
		a.enterDir(full)
		return false
	}
	if (method == "auto" || method == "editor") && (e.Kind == "text" || method == "editor") {
		editor := pickEditor()
		if editor == "" {
			a.showError("未找到编辑器（已尝试 vim/vi/nano/code），请设置 $EDITOR 环境变量")
			return true
		}
		a.suspendRun([]string{editor, full}, false)
		return true
	}
	if method == "auto" && e.Kind == "exe" {
		a.suspendRun([]string{full}, true)
		return true
	}
	opener.Open(full)
	return true
}

func (a *App) suspendRun(argv []string, waitKey bool) {
	a.erasePopup()
	term.Restore(a.fd, a.oldState)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	a.oldState, _ = term.Cbreak(a.fd)
	msg := ""
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() != 0 {
			msg = fmt.Sprintf("（退出码 %d）", ee.ExitCode())
		} else if !ok {
			msg = fmt.Sprintf("（启动失败: %v）", err)
		}
	}
	if waitKey || msg != "" {
		render.UI(msg + "按任意键返回…")
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
		if len(a.hist) > 0 {
			// 弹出栈顶：恢复到进入当前目录前的位置（记住上级的光标）
			h := a.hist[len(a.hist)-1]
			a.hist = a.hist[:len(a.hist)-1]
			a.cwd = h.cwd
			a.cursor, a.top = h.cursor, h.top
			a.load()
		} else {
			parent := filepath.Dir(a.cwd)
			if parent != a.cwd {
				a.cwd = parent
				a.load()
				a.cursor, a.top = 0, 0
			}
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
				a.enterDir(filepath.Join(a.cwd, a.entries[a.cursor].Name))
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
	case "tab":
		a.sortMode = (a.sortMode + 1) % len(fs.SortLabels)
		keep := ""
		if n > 0 {
			keep = a.entries[a.cursor].Name
		}
		a.load()
		if keep != "" { // 重排后尽量保持光标停留在同一项
			for i := range a.entries {
				if a.entries[i].Name == keep {
					a.cursor, a.top = i, 0
					break
				}
			}
		}
		a.previewOn = false
		return "redraw"
	case " ":
		if a.previewOn {
			a.previewOn = false
		} else if n > 0 {
			a.loadPreview(a.entries[a.cursor])
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

// Run 启动交互主循环，返回进程退出码。
func (a *App) Run(fd int) int {
	a.fd = fd
	old, err := term.Cbreak(fd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "nav: 无法进入终端原始模式:", err)
		return 2
	}
	a.oldState = old
	defer term.Restore(fd, a.oldState)

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
				term.Restore(fd, a.oldState)
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
