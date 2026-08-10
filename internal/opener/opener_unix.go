//go:build !windows

package opener

import (
	"os/exec"
	"runtime"
)

// Open 用系统默认应用打开（不阻塞）。
func Open(path string) {
	cmd := "xdg-open"
	if runtime.GOOS == "darwin" {
		cmd = "open"
	}
	_ = exec.Command(cmd, path).Start()
}
