package main

import (
	"os"
	"time"
)

var keyCh = make(chan byte, 32)

// keyReaderLoop：单 goroutine 串行读 fd0（cbreak 模式），
// 避免 bufio 并发读同一 fd 的问题。
func keyReaderLoop() {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if n == 1 {
			keyCh <- buf[0]
		}
		if err != nil {
			return
		}
	}
}

// readKey 读取一个按键（阻塞）。方向键是 ESC [ A/B/C/D 三字节序列，
// 兼容 OA 应用光标模式；单独 Esc 只发一字节，用 100ms 超时区分。
func readKey() string {
	b := <-keyCh
	if b == 0x1b {
		select {
		case b2 := <-keyCh:
			b3 := <-keyCh
			switch {
			case (b2 == '[' || b2 == 'O') && b3 == 'A':
				return "up"
			case (b2 == '[' || b2 == 'O') && b3 == 'B':
				return "down"
			case (b2 == '[' || b2 == 'O') && b3 == 'C':
				return "right"
			case (b2 == '[' || b2 == 'O') && b3 == 'D':
				return "left"
			}
			return "?"
		case <-time.After(100 * time.Millisecond):
			return "esc"
		}
	}
	switch b {
	case '\r', '\n':
		return "enter"
	case 0x03:
		return "ctrl-c" // ISIG 保留时不会出现；防御性处理
	}
	return string(b)
}
