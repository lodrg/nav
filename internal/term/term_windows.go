//go:build windows

package term

import "golang.org/x/sys/windows"

// Cbreak：Windows 控制台模式，清 line-input/echo（无 ISIG 概念，
// Ctrl+C 由控制台产生 CTRL_C_EVENT，默认行为即终止）。
func Cbreak(fd int) (any, error) {
	h := windows.Handle(fd)
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return nil, err
	}
	old := mode
	mode &^= windows.ENABLE_LINE_INPUT | windows.ENABLE_ECHO_INPUT
	if err := windows.SetConsoleMode(h, mode); err != nil {
		return nil, err
	}
	return &old, nil
}

func Restore(fd int, old any) {
	if m, ok := old.(*uint32); ok {
		_ = windows.SetConsoleMode(windows.Handle(fd), *m)
	}
}

func Size(fd int) (int, int) {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(fd), &info); err != nil {
		return 80, 24
	}
	return int(info.Window.Right - info.Window.Left + 1),
		int(info.Window.Bottom - info.Window.Top + 1)
}
