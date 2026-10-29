package doctor_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jinrunli/bay/internal/doctor"
	"github.com/jinrunli/bay/internal/model"
	"github.com/jinrunli/bay/internal/scan"
)

func TestBuildFindings(t *testing.T) {
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	staleDays := 90
	old := now.AddDate(0, 0, -100).Format(time.RFC3339)
	fresh := now.AddDate(0, 0, -10).Format(time.RFC3339)

	projects := []model.Project{
		{
			Name: "unk", Root: "/r", Path: "/r/unk",
			Status: model.StatusActive, PurposeKnown: false,
			Signals: mustSignals(t, scan.Signals{Mtime: fresh}),
			LastSeenAt: now,
		},
		{
			Name: "old-paused", Root: "/r", Path: "/r/old-paused",
			Status: model.StatusPaused, PurposeKnown: true,
			Signals:    mustSignals(t, scan.Signals{GitLastCommitAt: old, Mtime: old}),
			LastSeenAt: now,
		},
		{
			Name: "old-active", Root: "/r", Path: "/r/old-active",
			Status: model.StatusActive, PurposeKnown: true,
			Signals:    mustSignals(t, scan.Signals{Mtime: old}),
			LastSeenAt: now,
		},
		{
			Name: "in-trash", Root: "/r", Path: "/r/_pm/trash/in-trash",
			Status: model.StatusTrash, PurposeKnown: true,
			Signals:    mustSignals(t, scan.Signals{Mtime: fresh}),
			LastSeenAt: now,
		},
		{
			Name: "gone", Root: "/r", Path: "/r/gone",
			Status: model.StatusActive, PurposeKnown: true, Missing: true,
			Signals:    mustSignals(t, scan.Signals{Mtime: fresh}),
			LastSeenAt: now,
		},
		{
			Name: "fresh-paused", Root: "/r", Path: "/r/fresh-paused",
			Status: model.StatusPaused, PurposeKnown: true,
			Signals:    mustSignals(t, scan.Signals{Mtime: fresh}),
			LastSeenAt: now,
		},
	}

	rep := doctor.Build(projects, staleDays, now)
	kindsByName := map[string][]string{}
	for _, f := range rep.Findings {
		kindsByName[f.Name] = append(kindsByName[f.Name], f.Kind)
		if f.Root == "" || f.Path == "" {
			t.Errorf("%s finding missing Root/Path: %+v", f.Name, f)
		}
	}

	assertKinds(t, kindsByName["unk"], "unknown_purpose")
	assertKinds(t, kindsByName["old-paused"], "stale_paused")
	assertKinds(t, kindsByName["old-active"], "stale_active")
	assertKinds(t, kindsByName["in-trash"], "trash")
	assertKinds(t, kindsByName["gone"], "missing")
	if _, ok := kindsByName["fresh-paused"]; ok {
		t.Fatalf("fresh paused must not get findings, got %v", kindsByName["fresh-paused"])
	}
}

func TestLastActivePrefersNewerSignal(t *testing.T) {
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	old := now.AddDate(0, 0, -100).Format(time.RFC3339)
	fresh := now.AddDate(0, 0, -5).Format(time.RFC3339)

	// git old, mtime fresh → not stale
	p := model.Project{
		Name: "mixed", Root: "/r", Path: "/r/mixed",
		Status: model.StatusPaused, PurposeKnown: true,
		Signals:    mustSignals(t, scan.Signals{GitLastCommitAt: old, Mtime: fresh}),
		LastSeenAt: now.AddDate(0, 0, -200),
	}
	rep := doctor.Build([]model.Project{p}, 90, now)
	if len(rep.Findings) != 0 {
		t.Fatalf("expected no stale when newer signal is fresh, got %+v", rep.Findings)
	}

	// neither signal parseable → LastSeenAt (stale)
	p2 := model.Project{
		Name: "fallback", Root: "/r", Path: "/r/fallback",
		Status: model.StatusActive, PurposeKnown: true,
		Signals:    `{"git_last_commit_at":"not-a-date","mtime":""}`,
		LastSeenAt: now.AddDate(0, 0, -120),
	}
	rep2 := doctor.Build([]model.Project{p2}, 90, now)
	assertKinds(t, kindsOf(rep2, "fallback"), "stale_active")
}

func TestFormatTextAndMarkdown(t *testing.T) {
	rep := doctor.Report{Findings: []doctor.Finding{
		{Kind: "unknown_purpose", Name: "a", Root: "/r", Path: "/r/a", Detail: "purpose unknown"},
		{Kind: "trash", Name: "b", Root: "/r", Path: "/r/b", Detail: "in trash"},
	}}
	txt := rep.FormatText()
	md := rep.FormatMarkdown()
	for _, s := range []string{"unknown_purpose", "a", "trash", "b"} {
		if !strings.Contains(txt, s) {
			t.Errorf("FormatText missing %q:\n%s", s, txt)
		}
		if !strings.Contains(md, s) {
			t.Errorf("FormatMarkdown missing %q:\n%s", s, md)
		}
	}
	if !strings.Contains(md, "#") && !strings.Contains(md, "|") && !strings.Contains(md, "-") {
		t.Errorf("FormatMarkdown looks non-markdown: %q", md)
	}
}

func mustSignals(t *testing.T, s scan.Signals) string {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func kindsOf(rep doctor.Report, name string) []string {
	var out []string
	for _, f := range rep.Findings {
		if f.Name == name {
			out = append(out, f.Kind)
		}
	}
	return out
}

func assertKinds(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("kinds got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("kinds[%d] got %q want %q (full %v)", i, got[i], want[i], got)
		}
	}
}
