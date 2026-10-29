package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jinrunli/bay/internal/config"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := config.Config{
		Roots:          []string{"/a", "/b"},
		Ignore:         []string{"node_modules"},
		StaleAfterDays: 30,
		PMDir:          "_pm",
	}
	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := config.LoadOrDefault(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Roots) != 2 || got.Roots[0] != "/a" || got.StaleAfterDays != 30 {
		t.Fatalf("%+v", got)
	}
}

func TestLoadOrDefaultMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nope.yaml")
	got, err := config.LoadOrDefault(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.StaleAfterDays != 90 || got.PMDir != "_pm" {
		t.Fatalf("%+v", got)
	}
	if len(got.Roots) != 2 {
		t.Fatalf("want 2 default roots, got %v", got.Roots)
	}
	_ = os.Getenv // silence if unused; Default() must expand home
}
