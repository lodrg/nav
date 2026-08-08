package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	reset   = "\x1b[0m"
	dim     = "\x1b[2m"
	red     = "\x1b[31m"
	reverse = "\x1b[7m"
	clearLn = "\x1b[2K"
)

const (
	iconDir   = "📁"
	iconFile  = "📄"
	iconExe   = "⚙️"
	iconMedia = "🖼️"
)

// isWideRune：East Asian Width W/F 的宽字符区间（经典 wcwidth 表，
// 覆盖 CJK/Hangul/emoji 象形文字；与 Python east_asian_width in "WF" 语义一致）。
func isWideRune(r rune) bool {
	switch {
	case r >= 0x1100 && r <= 0x115F, // Hangul Jamo
		r >= 0x2E80 && r <= 0x303E,   // CJK 部首..CJK 符号
		r >= 0x3041 && r <= 0x33FF,   // 平假名..CJK 兼容
		r >= 0x3400 && r <= 0x4DBF,   // CJK 扩展 A
		r >= 0x4E00 && r <= 0x9FFF,   // CJK 统一表意
		r >= 0xA000 && r <= 0xA4CF,   // 彝文
		r >= 0xAC00 && r <= 0xD7A3,   // 谚文音节
		r >= 0xF900 && r <= 0xFAFF,   // CJK 兼容表意
		r >= 0xFE30 && r <= 0xFE4F,   // CJK 兼容形式
		r >= 0xFF00 && r <= 0xFF60,   // 全角形式
		r >= 0xFFE0 && r <= 0xFFE6,   // 全角符号
		r >= 0x1F300 && r <= 0x1FAFF, // emoji 象形文字
		r >= 0x20000 && r <= 0x3FFFD: // CJK 扩展 B+
		return true
	}
	return false
}

// dispWidth：CJK/emoji 按 2 列；ZWJ/VS16 零宽；
// 后跟 VS16 的字符按 emoji 呈现计 2 列（⚙️ 在终端里占 2 列）。
func dispWidth(s string) int {
	runes := []rune(s)
	w := 0
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == 0x200D || r == 0xFE0F { // ZWJ / VS16
			continue
		}
		rw := 1
		if isWideRune(r) || (i+1 < len(runes) && runes[i+1] == 0xFE0F) {
			rw = 2
		}
		w += rw
	}
	return w
}

// truncate 宽度感知截断：不撕裂 CJK，超宽补 …。
func truncate(s string, w int) string {
	if dispWidth(s) <= w {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		// 用 dispWidth 累加（而非逐 rune 宽度），保证 VS16 组合宽度一致
		if dispWidth(b.String()+string(r)) > w-1 {
			break
		}
		b.WriteRune(r)
	}
	return b.String() + "…"
}

func up(n int) string {
	if n > 0 {
		return fmt.Sprintf("\x1b[%dA", n)
	}
	return ""
}

// renderRegion 在终端内联区域重绘（纯函数）。lines 必须是已按宽度截断的行。
// 返回 (转义串, 新高度)。算法：上移 lastHeight 行 → 清空并重写
// max(lastHeight, height) 行 → 若旧区域更高则把光标移回新区域底部。
func renderRegion(lines []string, height, width, lastHeight int, highlight *int) (string, int) {
	total := max(lastHeight, height)
	var b strings.Builder
	b.WriteString(up(lastHeight))
	for i := 0; i < total; i++ {
		b.WriteString(clearLn)
		if i < height && i < len(lines) {
			row := lines[i]
			if highlight != nil && i == *highlight {
				row = reverse + row + reset
			}
			b.WriteString(row)
		}
		b.WriteString("\n")
	}
	if lastHeight > height {
		b.WriteString(up(lastHeight - height))
	}
	return b.String(), height
}

func formatRow(e Entry, width int) string {
	icon := iconFile
	switch {
	case e.IsDir:
		icon = iconDir
	case e.Kind == "exe":
		icon = iconExe
	case e.Kind == "media":
		icon = iconMedia
	}
	// icon(2) + 空格 + size(9) + 空格 + mtime(16) + 空格 = 30
	nameW := width - 30
	if nameW < 8 {
		nameW = 8
	}
	name := truncate(e.Name, nameW)
	pad := nameW - dispWidth(name)
	if pad < 0 {
		pad = 0
	}
	size := e.Size
	if e.IsDir {
		size = "—"
	}
	return fmt.Sprintf("%s %s%s %9s %s", icon, name, strings.Repeat(" ", pad), size, e.MTime)
}

// ui：UI 输出统一走 stderr，保证 stdout 只承载结果路径
// （ncd / cd "$(nav --print)" 依赖命令替换捕获 stdout）。
func ui(s string) {
	os.Stderr.WriteString(s)
}
