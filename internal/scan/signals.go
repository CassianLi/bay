package scan

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

type Signals struct {
	DirName         string   `json:"dir_name"`
	ReadmePreview   string   `json:"readme_preview,omitempty"`
	ReadmeEmpty     bool     `json:"readme_empty,omitempty"`
	GitLastCommitAt string   `json:"git_last_commit_at,omitempty"`
	GitSubject      string   `json:"git_subject,omitempty"`
	Mtime           string   `json:"mtime,omitempty"`
	StackTags       []string `json:"stack_tags,omitempty"`
}

var stackProbeFiles = []struct {
	file string
	tag  string
}{
	{"package.json", "node"},
	{"go.mod", "go"},
	{"pyproject.toml", "python"},
	{"Cargo.toml", "rust"},
}

func CollectSignals(dir string) (Signals, error) {
	sig := Signals{DirName: filepath.Base(dir)}

	if fi, err := os.Stat(dir); err == nil {
		sig.Mtime = fi.ModTime().UTC().Format(time.RFC3339)
	}

	readme, empty := readReadme(dir)
	if empty {
		sig.ReadmeEmpty = true
	} else if readme != "" {
		sig.ReadmePreview = readme
	}

	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		if at, subject, ok := gitLastCommit(dir); ok {
			sig.GitLastCommitAt = at
			sig.GitSubject = subject
		}
	}

	for _, p := range stackProbeFiles {
		if _, err := os.Stat(filepath.Join(dir, p.file)); err == nil {
			sig.StackTags = append(sig.StackTags, p.tag)
		}
	}

	return sig, nil
}

func readReadme(dir string) (preview string, empty bool) {
	var data []byte
	for _, name := range []string{"README.md", "README"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err == nil {
			data = b
			break
		}
	}
	if data == nil {
		return "", false
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return "", true
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}
		title := strings.TrimSpace(strings.TrimLeft(line, "#"))
		if title != "" {
			return title, false
		}
	}
	return truncateRunes(trimmed, 500), false
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n])
}

func gitLastCommit(dir string) (at, subject string, ok bool) {
	out, err := exec.Command("git", "-C", dir, "log", "-1", "--format=%cI%n%s").Output()
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "", "", false
	}
	at = parts[0]
	if len(parts) > 1 {
		subject = parts[1]
	}
	return at, subject, true
}
