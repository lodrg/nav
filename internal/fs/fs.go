// Package fs 目录扫描、类型判定与排序。
package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Entry 是目录列表中的一个条目。
type Entry struct {
	Name   string
	IsDir  bool
	Kind   string // dir | exe | text | media | unknown
	Size   string
	SizeN  int64 // 原始字节数（排序用）
	MTime  string
	MTimeN time.Time // 原始修改时间（排序用）
}

const (
	SortName = iota // 名称（casefold，Finder 风格）
	SortTime        // 修改时间（新→旧）
	SortSize        // 大小（大→小）
)

var SortLabels = []string{"名称", "时间", "大小"}

var textExts = map[string]bool{
	".py": true, ".pyw": true, ".md": true, ".txt": true, ".json": true,
	".toml": true, ".yaml": true, ".yml": true, ".ini": true, ".cfg": true,
	".conf": true, ".c": true, ".h": true, ".cpp": true, ".hpp": true,
	".cc": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
	".sh": true, ".bash": true, ".zsh": true, ".csv": true, ".tsv": true,
	".xml": true, ".html": true, ".htm": true, ".css": true, ".scss": true,
	".less": true, ".go": true, ".rs": true, ".rb": true, ".java": true,
	".kt": true, ".swift": true, ".sql": true, ".log": true, ".rst": true,
	".tex": true, ".lock": true, ".gitignore": true, ".env": true,
	".dockerfile": true,
}

var mediaExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".bmp": true, ".tiff": true, ".svg": true, ".ico": true, ".heic": true,
	".pdf": true, ".mp4": true, ".mov": true, ".mkv": true, ".avi": true,
	".webm": true, ".mp3": true, ".wav": true, ".flac": true, ".m4a": true,
	".aac": true, ".ogg": true, ".doc": true, ".docx": true, ".xls": true,
	".xlsx": true, ".ppt": true, ".pptx": true, ".zip": true, ".tar": true,
	".gz": true, ".bz2": true, ".xz": true, ".7z": true, ".dmg": true,
	".pkg": true, ".app": true, ".pages": true, ".numbers": true,
	".key": true, ".epub": true,
}

// exactTextNames：无扩展名/点开头的文本判定（扩展名缺失时做二进制嗅探兜底）。
var exactTextNames = map[string]bool{
	".env": true, ".gitignore": true, ".gitattributes": true,
	".dockerignore": true, "Makefile": true, "Dockerfile": true,
	"LICENSE": true, "README": true,
}

func sniffText(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 8192)
	n, _ := f.Read(buf)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return false
		}
	}
	return true
}

// Classify 返回条目的类型：dir | exe | text | media | unknown。
// dir 是被扫描目录的路径（嗅探无扩展名文件时需要完整路径）。
func Classify(dir string, d os.DirEntry) string {
	if d.IsDir() {
		return "dir"
	}
	if info, err := d.Info(); err == nil && info.Mode()&0111 != 0 {
		return "exe"
	}
	name := d.Name()
	ext := strings.ToLower(filepath.Ext(name))
	if textExts[ext] || exactTextNames[name] {
		return "text"
	}
	if mediaExts[ext] {
		return "media"
	}
	if ext == "" {
		if sniffText(filepath.Join(dir, name)) {
			return "text"
		}
		return "unknown"
	}
	return "unknown"
}

func HumanSize(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	v := float64(n)
	for _, u := range []string{"KB", "MB", "GB", "TB"} {
		v /= 1024
		if v < 1024 {
			return fmt.Sprintf("%.1f %s", v, u)
		}
	}
	return fmt.Sprintf("%.1f PB", v)
}

// Scan 扫描目录（排序由 SortEntries 按模式处理，在调用方 load 中执行）。
func Scan(path string, showHidden bool) ([]Entry, error) {
	des, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(des))
	for _, d := range des {
		name := d.Name()
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}
		var size, mtime string
		var sizeN int64
		var mtimeN time.Time
		if info, err := d.Info(); err == nil {
			mtimeN = info.ModTime()
			if d.IsDir() {
				size = "—"
			} else {
				sizeN = info.Size()
				size = HumanSize(sizeN)
			}
			mtime = mtimeN.Format("2006-01-02 15:04")
		} else {
			size, mtime = "?", ""
		}
		out = append(out, Entry{
			Name:   name,
			IsDir:  d.IsDir(),
			Kind:   Classify(path, d),
			Size:   size,
			SizeN:  sizeN,
			MTime:  mtime,
			MTimeN: mtimeN,
		})
	}
	return out, nil
}

// SortEntries 按模式排序：目录始终在前。
// SortName: casefold 名称升序；SortTime: 修改时间新→旧；SortSize: 大小大→小。
// 时间/大小相同时回退到名称比较，保证顺序稳定。
func SortEntries(items []Entry, mode int) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		switch mode {
		case SortTime:
			if !items[i].MTimeN.Equal(items[j].MTimeN) {
				return items[i].MTimeN.After(items[j].MTimeN)
			}
		case SortSize:
			if items[i].SizeN != items[j].SizeN {
				return items[i].SizeN > items[j].SizeN
			}
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
}
