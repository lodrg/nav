package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func mkEntry(t *testing.T, dir, name string, content []byte, mode os.FileMode) os.DirEntry {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, content, mode); err != nil {
		t.Fatal(err)
	}
	d, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range d {
		if e.Name() == name {
			return e
		}
	}
	t.Fatalf("entry %s not found", name)
	return nil
}

func TestClassifyDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	d, _ := os.ReadDir(dir)
	if got := Classify(dir, d[0]); got != "dir" {
		t.Fatalf("want dir, got %s", got)
	}
}

func TestClassifyTextByExt(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.py", "README.md", "Dockerfile", ".gitignore"} {
		d := mkEntry(t, dir, name, []byte("x"), 0o644)
		if got := Classify(dir, d); got != "text" {
			t.Fatalf("%s: want text, got %s", name, got)
		}
	}
}

func TestClassifyMediaByExt(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"pic.png", "doc.pdf", "a.zip"} {
		d := mkEntry(t, dir, name, []byte("x"), 0o644)
		if got := Classify(dir, d); got != "media" {
			t.Fatalf("%s: want media, got %s", name, got)
		}
	}
}

func TestClassifyExecutable(t *testing.T) {
	dir := t.TempDir()
	d := mkEntry(t, dir, "run.sh", []byte("#!/bin/sh\n"), 0o755)
	if got := Classify(dir, d); got != "exe" {
		t.Fatalf("want exe, got %s", got)
	}
}

func TestClassifySniffTextNoExt(t *testing.T) {
	dir := t.TempDir()
	d := mkEntry(t, dir, "NOTES", []byte("hello world\nno nul\n"), 0o644)
	if got := Classify(dir, d); got != "text" {
		t.Fatalf("want text (sniffed), got %s", got)
	}
}

func TestClassifySniffBinaryNoExt(t *testing.T) {
	dir := t.TempDir()
	d := mkEntry(t, dir, "blob", []byte{0x00, 0x01, 0x02, 0xff}, 0o644)
	if got := Classify(dir, d); got != "unknown" {
		t.Fatalf("want unknown, got %s", got)
	}
}

func TestScanSortedDirsFirstCasefold(t *testing.T) {
	dir := t.TempDir()
	for _, d := range []string{"b_dir", "A_dir"} {
		os.Mkdir(filepath.Join(dir, d), 0o755)
	}
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "B.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte("x"), 0o644)

	items, err := Scan(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	SortEntries(items, SortName)
	// casefold 排序："a.txt" < "B.txt"（大小写不敏感，Finder 风格）
	want := []string{"A_dir", "b_dir", "a.txt", "B.txt"}
	for i, w := range want {
		if items[i].Name != w {
			t.Fatalf("items[%d] = %s, want %s (all: %+v)", i, items[i].Name, w, names(items))
		}
	}
	if !items[0].IsDir || !items[1].IsDir {
		t.Fatal("dirs must come first")
	}
}

func TestSortByTimeAndSize(t *testing.T) {
	// 时间与大小故意相反，才能区分两种排序模式
	items := []Entry{
		{Name: "recent.txt", IsDir: false, SizeN: 10, MTimeN: time.Unix(300, 0)},
		{Name: "mid.txt", IsDir: false, SizeN: 50, MTimeN: time.Unix(200, 0)},
		{Name: "stale.txt", IsDir: false, SizeN: 999, MTimeN: time.Unix(100, 0)},
		{Name: "dir1", IsDir: true},
	}
	SortEntries(items, SortTime)
	if !items[0].IsDir {
		t.Fatal("dirs must come first")
	}
	// 时间新→旧: recent(300) > mid(200) > stale(100)
	if items[1].Name != "recent.txt" || items[2].Name != "mid.txt" || items[3].Name != "stale.txt" {
		t.Fatalf("time sort wrong: %+v", names(items))
	}
	SortEntries(items, SortSize)
	// 大小大→小: stale(999) > mid(50) > recent(10)
	if items[1].Name != "stale.txt" || items[2].Name != "mid.txt" || items[3].Name != "recent.txt" {
		t.Fatalf("size sort wrong: %+v", names(items))
	}
}

func TestScanShowHidden(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte("x"), 0o644)
	items, _ := Scan(dir, true)
	found := false
	for _, e := range items {
		if e.Name == ".hidden" {
			found = true
		}
	}
	if !found {
		t.Fatal(".hidden should be visible when showHidden")
	}
}

func TestScanUnreadableDir(t *testing.T) {
	_, err := Scan("/nonexistent-path-xyz", false)
	if err == nil {
		t.Fatal("want error for nonexistent path")
	}
}

func TestHumanSize(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"}, {1023, "1023 B"}, {1024, "1.0 KB"}, {5 * 1024 * 1024, "5.0 MB"},
	}
	for _, c := range cases {
		if got := HumanSize(c.in); got != c.want {
			t.Fatalf("HumanSize(%d) = %s, want %s", c.in, got, c.want)
		}
	}
}

func names(items []Entry) []string {
	out := make([]string, len(items))
	for i, e := range items {
		out[i] = e.Name
	}
	return out
}
