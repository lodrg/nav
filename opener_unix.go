//go:build !windows

package main

import (
	"os/exec"
	"runtime"
)

// openDefault 用系统默认应用打开（不阻塞）。
func openDefault(path string) {
	cmd := "xdg-open"
	if runtime.GOOS == "darwin" {
		cmd = "open"
	}
	_ = exec.Command(cmd, path).Start()
}
