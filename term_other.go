//go:build !darwin && !linux && !windows

package main

import "golang.org/x/sys/unix"

// 其他 POSIX（FreeBSD 等）：BSD 风格 ioctl 请求名，与 darwin 一致。
func makeCbreak(fd int) (*unix.Termios, error) {
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

func restoreTerm(fd int, old any) {
	if t, ok := old.(*unix.Termios); ok {
		_ = unix.IoctlSetTermios(fd, unix.TIOCSETA, t)
	}
}

func termSize(fd int) (int, int) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 || ws.Row == 0 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}
