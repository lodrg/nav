//go:build windows

package main

import "os/exec"

// openDefault：Windows 用 ShellExecute 语义的 start（第一个空参是窗口标题）。
func openDefault(path string) {
	_ = exec.Command("cmd", "/c", "start", "", path).Start()
}
