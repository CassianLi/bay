package main

import "testing"

func TestDisplayWidthCJK(t *testing.T) {
	if got := displayWidth("ab"); got != 2 {
		t.Fatalf("ascii: %d", got)
	}
	if got := displayWidth("中文"); got != 4 {
		t.Fatalf("cjk: %d", got)
	}
	if got := displayWidth("A中"); got != 3 {
		t.Fatalf("mixed: %d", got)
	}
}

func TestTruncateDisplay(t *testing.T) {
	got := truncateDisplay("跨境电商清关证据链生成系统", 10)
	if displayWidth(got) > 10 {
		t.Fatalf("too wide %q width=%d", got, displayWidth(got))
	}
	if got[len(got)-len("…"):] != "…" && !containsEllipsis(got) {
		t.Fatalf("expected ellipsis, got %q", got)
	}
}

func TestPadDisplayAlignsMixed(t *testing.T) {
	a := padDisplay("Aster", 10)
	b := padDisplay("中文名", 10)
	if displayWidth(a) != 10 || displayWidth(b) != 10 {
		t.Fatalf("pad widths %d %d", displayWidth(a), displayWidth(b))
	}
}

func containsEllipsis(s string) bool {
	for _, r := range s {
		if r == '…' {
			return true
		}
	}
	return false
}
