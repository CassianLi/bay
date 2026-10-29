package scan_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jinrunli/bay/internal/config"
	"github.com/jinrunli/bay/internal/model"
	"github.com/jinrunli/bay/internal/scan"
	"github.com/jinrunli/bay/internal/store"
)

func TestRunDiscoversDailyAndPMAndLegacy(t *testing.T) {
	root := t.TempDir()
	mustMk := func(parts ...string) {
		p := filepath.Join(append([]string{root}, parts...)...)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mustMk("alpha")
	os.WriteFile(filepath.Join(root, "alpha", "README.md"), []byte("# Alpha tool\n\nDoes X.\n"), 0o644)
	mustMk("_pm", "incubator", "beta")
	mustMk("_archive", "gamma")
	mustMk("_pm") // container only

	db := filepath.Join(t.TempDir(), "t.db")
	st, err := store.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	cfg := config.Config{Roots: []string{root}, PMDir: "_pm", StaleAfterDays: 90}
	res, err := scan.Run(cfg, st)
	if err != nil {
		t.Fatal(err)
	}
	if res.Added != 3 {
		t.Fatalf("added=%d want 3", res.Added)
	}
	alpha, _ := st.GetByName("alpha", root)
	if alpha.Status != model.StatusPaused || alpha.PurposeKnown {
		t.Fatalf("%+v", alpha)
	}
	beta, _ := st.GetByName("beta", root)
	if beta.Status != model.StatusIdea {
		t.Fatalf("%+v", beta)
	}
	gamma, _ := st.GetByName("gamma", root)
	if gamma.Status != model.StatusArchived {
		t.Fatalf("%+v", gamma)
	}
}
