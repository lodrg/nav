//go:build darwin

package term

import "golang.org/x/sys/unix"

// Cbreak 进入 cbreak 模式：清 ICANON|ECHO，保留 ISIG（Ctrl+C 仍是 SIGINT）
// 和 OPOST（\n 正常换行）。返回旧状态用于恢复。
// darwin（BSD）的 termios ioctl 请求是 TIOCGETA/TIOCSETA。
func Cbreak(fd int) (any, error) {
	old, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return nil, err
	}
	t := *old
	t.Lflag &^= unix.ICANON | unix.ECHO
	t.Cc[unix.VMIN] = 1
	t.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(fd, unix.TIOCSETA, &t); err != nil {
		return nil, err
	}
	return old, nil
}

func Restore(fd int, old any) {
	if t, ok := old.(*unix.Termios); ok {
		_ = unix.IoctlSetTermios(fd, unix.TIOCSETA, t)
	}
}

func Size(fd int) (int, int) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 || ws.Row == 0 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}
