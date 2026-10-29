package doctor

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jinrunli/bay/internal/model"
	"github.com/jinrunli/bay/internal/scan"
)

const (
	KindUnknownPurpose = "unknown_purpose"
	KindStalePaused    = "stale_paused"
	KindStaleActive    = "stale_active"
	KindTrash          = "trash"
	KindMissing        = "missing"
)

type Finding struct {
	Kind   string
	Name   string
	Root   string
	Path   string
	Detail string
}

type Report struct {
	Findings []Finding
}

func Build(projects []model.Project, staleAfterDays int, now time.Time) Report {
	var findings []Finding
	cutoff := now.AddDate(0, 0, -staleAfterDays)

	for _, p := range projects {
		base := Finding{Name: p.Name, Root: p.Root, Path: p.Path}

		if !p.PurposeKnown {
			f := base
			f.Kind = KindUnknownPurpose
			f.Detail = "purpose unknown; set a summary"
			findings = append(findings, f)
		}

		lastActive := lastActiveAt(p)
		stale := lastActive.Before(cutoff)

		if p.Status == model.StatusPaused && stale {
			f := base
			f.Kind = KindStalePaused
			f.Detail = fmt.Sprintf("paused; last active %s (>%d days)", lastActive.UTC().Format(time.RFC3339), staleAfterDays)
			findings = append(findings, f)
		}
		if p.Status == model.StatusActive && stale {
			f := base
			f.Kind = KindStaleActive
			f.Detail = fmt.Sprintf("active; last active %s (>%d days)", lastActive.UTC().Format(time.RFC3339), staleAfterDays)
			findings = append(findings, f)
		}
		if p.Status == model.StatusTrash {
			f := base
			f.Kind = KindTrash
			f.Detail = "in trash; clean up manually"
			findings = append(findings, f)
		}
		if p.Missing {
			f := base
			f.Kind = KindMissing
			f.Detail = "path missing on disk"
			findings = append(findings, f)
		}
	}

	return Report{Findings: findings}
}

func lastActiveAt(p model.Project) time.Time {
	var sig scan.Signals
	_ = json.Unmarshal([]byte(p.Signals), &sig)

	var best time.Time
	var ok bool
	for _, s := range []string{sig.GitLastCommitAt, sig.Mtime} {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			continue
		}
		if !ok || t.After(best) {
			best = t
			ok = true
		}
	}
	if ok {
		return best
	}
	return p.LastSeenAt
}

func (r Report) FormatText() string {
	if len(r.Findings) == 0 {
		return "no findings\n"
	}
	var b strings.Builder
	for _, f := range r.Findings {
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%s\n", f.Kind, f.Name, f.Root, f.Path, f.Detail)
	}
	return b.String()
}

func (r Report) FormatMarkdown() string {
	var b strings.Builder
	b.WriteString("# bay doctor\n\n")
	if len(r.Findings) == 0 {
		b.WriteString("No findings.\n")
		return b.String()
	}
	b.WriteString("| Kind | Name | Root | Path | Detail |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")
	for _, f := range r.Findings {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n",
			f.Kind, f.Name, f.Root, f.Path, f.Detail)
	}
	return b.String()
}
