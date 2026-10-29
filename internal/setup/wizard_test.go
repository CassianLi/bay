package setup_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/CassianLi/bay/internal/config"
	"github.com/CassianLi/bay/internal/setup"
)

func TestWizardWritesConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	rootA := filepath.Join(dir, "a")
	rootB := filepath.Join(dir, "b")

	in := strings.NewReader(strings.Join([]string{
		rootA + ", " + rootB, // roots
		"60",                 // stale
		"_pm",                // pm
		"node_modules",       // ignore
		"y",                  // confirm
	}, "\n") + "\n")
	var out strings.Builder

	base := config.Default()
	got, err := setup.Wizard(in, &out, base, cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Roots) != 2 || got.StaleAfterDays != 60 || got.PMDir != "_pm" {
		t.Fatalf("%+v", got)
	}
	if len(got.Ignore) != 1 || got.Ignore[0] != "node_modules" {
		t.Fatalf("ignore=%v", got.Ignore)
	}
	loaded, err := config.LoadOrDefault(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.StaleAfterDays != 60 {
		t.Fatalf("saved %+v", loaded)
	}
}
