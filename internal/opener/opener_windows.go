//go:build windows

package opener

import "os/exec"

// Open：Windows 用 ShellExecute 语义的 start（第一个空参是窗口标题）。
func Open(path string) {
	_ = exec.Command("cmd", "/c", "start", "", path).Start()
}
