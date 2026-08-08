package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 构造不碰终端的 App 实例
func testApp(t *testing.T, path string, printMode, once bool) *App {
	t.Helper()
	app := &App{cwd: path, printMode: printMode, once: once}
	app.entries = []Entry{
		{Name: "a", IsDir: false, Kind: "text", Size: "1 B", MTime: "2026-01-01 00:00"},
		{Name: "b", IsDir: false, Kind: "media", Size: "2 B", MTime: "2026-01-01 00:00"},
		{Name: "sub", IsDir: true, Kind: "dir", Size: "—", MTime: "2026-01-01 00:00"},
	}
	return app
}

func TestUpDownWrap(t *testing.T) {
	app := testApp(t, "/tmp", false, false)
	if got := app.handleKey("down"); got != "redraw" || app.cursor != 1 {
		t.Fatalf("down: cursor=%d action=%s", app.cursor, got)
	}
	app.handleKey("down")
	if app.cursor != 2 {
		t.Fatalf("cursor=%d, want 2", app.cursor)
	}
	app.handleKey("down")
	if app.cursor != 0 {
		t.Fatalf("wrap: cursor=%d, want 0", app.cursor)
	}
	app.handleKey("up")
	if app.cursor != 2 {
		t.Fatalf("up: cursor=%d, want 2", app.cursor)
	}
}

func TestQuitKeys(t *testing.T) {
	app := testApp(t, "/tmp", false, false)
	for _, k := range []string{"q", "esc", "ctrl-c"} {
		if got := app.handleKey(k); got != "quit" {
			t.Fatalf("%s: action=%s, want quit", k, got)
		}
	}
}

func TestPrintModeSelectFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644)
	app := &App{cwd: dir, printMode: true}
	app.load()
	app.cursor = 0 // a.txt
	if got := app.handleKey("enter"); got != "print_path" {
		t.Fatalf("action=%s, want print_path", got)
	}
	if app.selected != filepath.Join(dir, "a.txt") {
		t.Fatalf("selected=%s", app.selected)
	}
}

func TestPrintModeSelectDir(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	app := &App{cwd: dir, printMode: true}
	app.load()
	app.cursor = 0 // sub 目录排第一
	if got := app.handleKey("enter"); got != "print_path" {
		t.Fatalf("action=%s, want print_path", got)
	}
	if app.selected != filepath.Join(dir, "sub") {
		t.Fatalf("selected=%s", app.selected)
	}
}

func TestPrintModeRightEntersDir(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	app := &App{cwd: dir, printMode: true}
	app.load()
	app.cursor = 0
	if got := app.handleKey("right"); got != "redraw" {
		t.Fatalf("action=%s, want redraw", got)
	}
	if app.cwd != filepath.Join(dir, "sub") {
		t.Fatalf("cwd=%s", app.cwd)
	}
	if app.selected != "" {
		t.Fatal("selected should be empty")
	}
}

func TestPrintModeRightOnFileNoop(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644)
	app := &App{cwd: dir, printMode: true}
	app.load()
	app.cursor = 0
	app.handleKey("right")
	if app.selected != "" {
		t.Fatal("→ on file in print mode must not select")
	}
}

func TestPrintModePPrintsCwd(t *testing.T) {
	app := testApp(t, "/tmp/xyz", true, false)
	if got := app.handleKey("p"); got != "print_path" {
		t.Fatalf("action=%s, want print_path", got)
	}
	if app.selected != "/tmp/xyz" {
		t.Fatalf("selected=%s", app.selected)
	}
}

func TestLeftGoesParent(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0o755)
	app := &App{cwd: sub}
	app.load()
	app.handleKey("left")
	if app.cwd != dir {
		t.Fatalf("cwd=%s, want %s", app.cwd, dir)
	}
}

func TestToggleHidden(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".h"), []byte("x"), 0o644)
	app := &App{cwd: dir}
	app.load()
	for _, e := range app.entries {
		if e.Name == ".h" {
			t.Fatal(".h should be hidden by default")
		}
	}
	app.handleKey(".")
	found := false
	for _, e := range app.entries {
		if e.Name == ".h" {
			found = true
		}
	}
	if !found {
		t.Fatal(".h should be visible after toggle")
	}
}

func TestPreviewToggle(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.py"), []byte("print(1)\nprint(2)\n"), 0o644)
	app := &App{cwd: dir}
	app.load()
	for i, e := range app.entries {
		if e.Name == "a.py" {
			app.cursor = i
		}
	}
	app.handleKey(" ")
	if !app.previewOn || len(app.preview) != 2 || app.preview[0] != "print(1)" {
		t.Fatalf("preview=%v", app.preview)
	}
	app.handleKey(" ")
	if app.previewOn {
		t.Fatal("preview should be off after second space")
	}
}

func TestPreviewNonTextNotice(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pic.png"), []byte("x"), 0o644)
	app := &App{cwd: dir}
	app.load()
	for i, e := range app.entries {
		if e.Name == "pic.png" {
			app.cursor = i
		}
	}
	app.handleKey(" ")
	if !app.previewOn {
		t.Fatal("preview should show notice for non-text")
	}
}

func TestHjklNavigation(t *testing.T) {
	app := testApp(t, "/tmp", false, false)
	// j = down, k = up
	app.handleKey("j")
	if app.cursor != 1 {
		t.Fatalf("j: cursor=%d, want 1", app.cursor)
	}
	app.handleKey("j")
	if app.cursor != 2 {
		t.Fatalf("j: cursor=%d, want 2", app.cursor)
	}
	app.handleKey("k")
	if app.cursor != 1 {
		t.Fatalf("k: cursor=%d, want 1", app.cursor)
	}
}

func TestHjklLeftRight(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0o755)
	// l = right（进入目录）
	app := &App{cwd: dir, printMode: true}
	app.load()
	app.cursor = 0 // sub
	if got := app.handleKey("l"); got != "redraw" {
		t.Fatalf("l: action=%s", got)
	}
	if app.cwd != sub {
		t.Fatalf("l should enter dir, cwd=%s", app.cwd)
	}
	// h = left（返回上级）
	app.handleKey("h")
	if app.cwd != dir {
		t.Fatalf("h should go parent, cwd=%s", app.cwd)
	}
}

func TestFormatPathHomeAbbrev(t *testing.T) {
	home, _ := os.UserHomeDir()
	app := &App{cwd: home, entries: []Entry{{Name: "x"}}}
	line := app.formatPath(100)
	if !strings.Contains(line, "~") {
		t.Fatalf("home should abbreviate to ~: %q", line)
	}
	if !strings.Contains(line, "1 项") {
		t.Fatalf("should show entry count: %q", line)
	}
}

func TestEnterDirChangesCwd(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	app := &App{cwd: dir}
	app.load()
	idx := -1
	for i, e := range app.entries {
		if e.IsDir {
			idx = i
		}
	}
	subName := app.entries[idx].Name
	if opened := app.openEntry(idx, "auto"); opened {
		t.Fatal("entering dir must not count as opened_file")
	}
	if app.cwd != filepath.Join(dir, subName) {
		t.Fatalf("cwd=%s", app.cwd)
	}
}
