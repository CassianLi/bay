package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
	"unicode"
)

func newTable(w io.Writer) *tabwriter.Writer {
	// ASCII key/value tables (show/scan); CJK-heavy tables use writeAligned.
	return tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.UTC().Format(time.RFC3339)
}

func formatPurpose(known bool) string {
	if known {
		return "known"
	}
	return "unknown"
}

// runeDisplayWidth approximates terminal columns for a rune (CJK/emoji = 2).
func runeDisplayWidth(r rune) int {
	switch {
	case r == 0 || r == '\n' || r == '\r':
		return 0
	case unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r):
		return 0
	case r < 0x1100:
		return 1
	case r <= 0x115F, // Hangul Jamo init
		r == 0x2329, r == 0x232A,
		r >= 0x2E80 && r <= 0xA4CF && r != 0x303F,
		r >= 0xAC00 && r <= 0xD7A3, // Hangul syllables
		r >= 0xF900 && r <= 0xFAFF,
		r >= 0xFE10 && r <= 0xFE19,
		r >= 0xFE30 && r <= 0xFE6F,
		r >= 0xFF00 && r <= 0xFF60,
		r >= 0xFFE0 && r <= 0xFFE6,
		r >= 0x1F300 && r <= 0x1FAFF, // emoji / symbols (approx)
		r >= 0x20000 && r <= 0x3FFFD:
		return 2
	default:
		return 1
	}
}

func displayWidth(s string) int {
	n := 0
	for _, r := range s {
		n += runeDisplayWidth(r)
	}
	return n
}

// truncateDisplay shortens s to at most max terminal columns, appending "…" when clipped.
func truncateDisplay(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if max <= 0 || s == "" {
		return s
	}
	if displayWidth(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	var b strings.Builder
	w := 0
	ellipsis := displayWidth("…")
	for _, r := range s {
		rw := runeDisplayWidth(r)
		if w+rw+ellipsis > max {
			b.WriteRune('…')
			return b.String()
		}
		b.WriteRune(r)
		w += rw
	}
	return b.String()
}

func padDisplay(s string, width int) string {
	dw := displayWidth(s)
	if dw >= width {
		return s
	}
	return s + strings.Repeat(" ", width-dw)
}

// writeAligned prints columns padded by display width. Last column is not padded.
// gap is spaces between columns.
func writeAligned(w io.Writer, rows [][]string, gap int) error {
	if len(rows) == 0 {
		return nil
	}
	cols := len(rows[0])
	widths := make([]int, cols)
	for _, row := range rows {
		for i := 0; i < cols-1 && i < len(row); i++ {
			if dw := displayWidth(row[i]); dw > widths[i] {
				widths[i] = dw
			}
		}
	}
	if gap < 1 {
		gap = 2
	}
	sep := strings.Repeat(" ", gap)
	for _, row := range rows {
		for i := 0; i < cols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			if i > 0 {
				if _, err := io.WriteString(w, sep); err != nil {
					return err
				}
			}
			if i < cols-1 {
				cell = padDisplay(cell, widths[i])
			}
			if _, err := io.WriteString(w, cell); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	return nil
}

// listSummary prefers project summary; falls back to signals readme_preview.
// max is terminal display columns (not rune count).
func listSummary(summary, signalsJSON string, max int) string {
	if s := strings.TrimSpace(summary); s != "" {
		return truncateDisplay(s, max)
	}
	var sig struct {
		ReadmePreview string `json:"readme_preview"`
	}
	if err := json.Unmarshal([]byte(signalsJSON), &sig); err == nil {
		if s := strings.TrimSpace(sig.ReadmePreview); s != "" {
			return truncateDisplay(s, max)
		}
	}
	return "-"
}
