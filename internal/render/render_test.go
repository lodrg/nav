package render

import (
	"strings"
	"testing"

	"nav/internal/fs"
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
		if got := DispWidth(c.s); got != c.want {
			t.Fatalf("DispWidth(%q) = %d, want %d", c.s, got, c.want)
		}
	}
}

func TestTruncateFits(t *testing.T) {
	if got := Truncate("hello", 10); got != "hello" {
		t.Fatalf("want hello, got %q", got)
	}
}

func TestTruncateCutsWithEllipsis(t *testing.T) {
	got := Truncate("hello world", 6)
	if got != "hello…" {
		t.Fatalf("want hello…, got %q", got)
	}
	if DispWidth(got) > 6 {
		t.Fatalf("width %d > 6", DispWidth(got))
	}
}

func TestTruncateCJKNoOverflow(t *testing.T) {
	got := Truncate("你好世界你好", 7)
	if DispWidth(got) > 7 {
		t.Fatalf("width %d > 7", DispWidth(got))
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("want ellipsis suffix, got %q", got)
	}
}

func TestRenderRegionBasic(t *testing.T) {
	out, h := RenderRegion([]string{"aa", "bb"}, 2, 20, 0, nil)
	if h != 2 {
		t.Fatalf("height = %d, want 2", h)
	}
	if !strings.Contains(out, ClearLn) {
		t.Fatal("missing clear-line sequences")
	}
	if !strings.Contains(out, "aa") || !strings.Contains(out, "bb") {
		t.Fatalf("missing rows: %q", out)
	}
}

func TestRenderRegionHighlight(t *testing.T) {
	hl := 1
	out, _ := RenderRegion([]string{"aa", "bb"}, 2, 20, 0, &hl)
	if strings.Contains(out, Reverse+"aa"+Reset) {
		t.Fatal("row 0 should not be highlighted")
	}
	if !strings.Contains(out, Reverse+"bb"+Reset) {
		t.Fatalf("row 1 should be highlighted: %q", out)
	}
}

func TestRenderRegionMovesUpBeforeRedraw(t *testing.T) {
	out, h := RenderRegion([]string{"aa"}, 1, 20, 3, nil)
	if !strings.HasPrefix(out, "\x1b[3A") {
		t.Fatalf("want move-up 3 prefix, got %q", out)
	}
	if h != 1 {
		t.Fatalf("height = %d, want 1", h)
	}
}

func TestRenderRegionShrinkReturnsCursor(t *testing.T) {
	out, _ := RenderRegion([]string{"aa"}, 1, 20, 3, nil)
	if !strings.HasSuffix(out, "\x1b[2A") {
		t.Fatalf("want trailing move-up 2, got %q", out)
	}
}

func TestRenderRegionFewerLinesThanHeight(t *testing.T) {
	out, h := RenderRegion([]string{"only"}, 5, 20, 0, nil)
	if h != 5 {
		t.Fatalf("height = %d, want 5", h)
	}
	if strings.Count(out, ClearLn) != 5 {
		t.Fatalf("want 5 cleared lines, got %d", strings.Count(out, ClearLn))
	}
}

func TestTruncateTailKeepsTail(t *testing.T) {
	got := TruncateTail("/very/long/中文/路径/to/nav-go", 16)
	if !strings.HasSuffix(got, "nav-go") {
		t.Fatalf("should keep tail, got %q", got)
	}
	if !strings.HasPrefix(got, "…") {
		t.Fatalf("should prefix ellipsis, got %q", got)
	}
	if DispWidth(got) > 16 {
		t.Fatalf("width %d > 16", DispWidth(got))
	}
}

func TestTruncateTailFits(t *testing.T) {
	if got := TruncateTail("/short", 20); got != "/short" {
		t.Fatalf("want /short, got %q", got)
	}
}

func TestRuneWidth(t *testing.T) {
	if runeWidth('中') != 2 || runeWidth('a') != 1 || runeWidth(0xFE0F) != 0 {
		t.Fatal("runeWidth basic cases wrong")
	}
}

func TestFormatRowWidthRespected(t *testing.T) {
	row := FormatRow(fs.Entry{
		Name: "很长的中文文件名.txt", Kind: "text",
		Size: "1.0 KB", MTime: "2026-01-01 00:00",
	}, 40)
	if DispWidth(row) > 40 {
		t.Fatalf("row width %d > 40: %q", DispWidth(row), row)
	}
	if !strings.Contains(row, "…") {
		t.Fatal("long name should be truncated")
	}
}
