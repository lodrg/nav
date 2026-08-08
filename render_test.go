package main

import (
	"strings"
	"testing"
)

func TestDispWidth(t *testing.T) {
	cases := []struct {
		s    string
		want int
	}{
		{"abc", 3},
		{"中文", 4},
		{"📁", 2},
		{"⚙️", 2}, // ⚙(1) + VS16(0)
		{"a中b", 4},
	}
	for _, c := range cases {
		if got := dispWidth(c.s); got != c.want {
			t.Fatalf("dispWidth(%q) = %d, want %d", c.s, got, c.want)
		}
	}
}

func TestTruncateFits(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Fatalf("want hello, got %q", got)
	}
}

func TestTruncateCutsWithEllipsis(t *testing.T) {
	got := truncate("hello world", 6)
	if got != "hello…" {
		t.Fatalf("want hello…, got %q", got)
	}
	if dispWidth(got) > 6 {
		t.Fatalf("width %d > 6", dispWidth(got))
	}
}

func TestTruncateCJKNoOverflow(t *testing.T) {
	got := truncate("你好世界你好", 7)
	if dispWidth(got) > 7 {
		t.Fatalf("width %d > 7", dispWidth(got))
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("want ellipsis suffix, got %q", got)
	}
}

func TestRenderRegionBasic(t *testing.T) {
	out, h := renderRegion([]string{"aa", "bb"}, 2, 20, 0, nil)
	if h != 2 {
		t.Fatalf("height = %d, want 2", h)
	}
	if !strings.Contains(out, clearLn) {
		t.Fatal("missing clear-line sequences")
	}
	if !strings.Contains(out, "aa") || !strings.Contains(out, "bb") {
		t.Fatalf("missing rows: %q", out)
	}
}

func TestRenderRegionHighlight(t *testing.T) {
	hl := 1
	out, _ := renderRegion([]string{"aa", "bb"}, 2, 20, 0, &hl)
	if strings.Contains(out, reverse+"aa"+reset) {
		t.Fatal("row 0 should not be highlighted")
	}
	if !strings.Contains(out, reverse+"bb"+reset) {
		t.Fatalf("row 1 should be highlighted: %q", out)
	}
}

func TestRenderRegionMovesUpBeforeRedraw(t *testing.T) {
	out, h := renderRegion([]string{"aa"}, 1, 20, 3, nil)
	if !strings.HasPrefix(out, "\x1b[3A") {
		t.Fatalf("want move-up 3 prefix, got %q", out)
	}
	if h != 1 {
		t.Fatalf("height = %d, want 1", h)
	}
}

func TestRenderRegionShrinkReturnsCursor(t *testing.T) {
	out, _ := renderRegion([]string{"aa"}, 1, 20, 3, nil)
	if !strings.HasSuffix(out, "\x1b[2A") {
		t.Fatalf("want trailing move-up 2, got %q", out)
	}
}

func TestRenderRegionFewerLinesThanHeight(t *testing.T) {
	out, h := renderRegion([]string{"only"}, 5, 20, 0, nil)
	if h != 5 {
		t.Fatalf("height = %d, want 5", h)
	}
	if strings.Count(out, clearLn) != 5 {
		t.Fatalf("want 5 cleared lines, got %d", strings.Count(out, clearLn))
	}
}

func TestFormatRowWidthRespected(t *testing.T) {
	row := formatRow(Entry{
		Name: "很长的中文文件名.txt", Kind: "text",
		Size: "1.0 KB", MTime: "2026-01-01 00:00",
	}, 40)
	if dispWidth(row) > 40 {
		t.Fatalf("row width %d > 40: %q", dispWidth(row), row)
	}
	if !strings.Contains(row, "…") {
		t.Fatal("long name should be truncated")
	}
}
