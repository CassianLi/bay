package paths_test

import (
	"path/filepath"
	"testing"

	"github.com/jinrunli/bay/internal/model"
	"github.com/jinrunli/bay/internal/paths"
)

func TestExpectedPath(t *testing.T) {
	root := "/Users/u/CursorProjects"
	if got := paths.ExpectedPath(root, "foo", "_pm", model.StatusIdea); got != filepath.Join(root, "_pm", "incubator", "foo") {
		t.Fatalf("idea: %s", got)
	}
	if got := paths.ExpectedPath(root, "foo", "_pm", model.StatusActive); got != filepath.Join(root, "foo") {
		t.Fatalf("active: %s", got)
	}
	if got := paths.ExpectedPath(root, "foo", "_pm", model.StatusArchived); got != filepath.Join(root, "_pm", "archive", "foo") {
		t.Fatalf("archived: %s", got)
	}
	if got := paths.ExpectedPath(root, "foo", "_pm", model.StatusTrash); got != filepath.Join(root, "_pm", "trash", "foo") {
		t.Fatalf("trash: %s", got)
	}
}

func TestInferStatusLegacyAndPM(t *testing.T) {
	root := "/Users/u/CursorProjects"
	cases := map[string]model.Status{
		filepath.Join(root, "foo"):                    model.StatusPaused,
		filepath.Join(root, "_pm", "incubator", "foo"): model.StatusIdea,
		filepath.Join(root, "_incubator", "foo"):       model.StatusIdea,
		filepath.Join(root, "_pm", "archive", "foo"):   model.StatusArchived,
		filepath.Join(root, "_archive", "foo"):         model.StatusArchived,
		filepath.Join(root, "_pm", "trash", "foo"):     model.StatusTrash,
	}
	for p, want := range cases {
		if got := paths.InferStatus(root, p, "_pm"); got != want {
			t.Fatalf("%s: got %s want %s", p, got, want)
		}
	}
}

func TestShouldSkipDirName(t *testing.T) {
	if !paths.ShouldSkipDirName("_pm", "_pm", nil) {
		t.Fatal("skip _pm")
	}
	if !paths.ShouldSkipDirName(".git", "_pm", nil) {
		t.Fatal("skip dot")
	}
	if !paths.ShouldSkipDirName("vendor-ish", "_pm", []string{"vendor-ish"}) {
		t.Fatal("skip ignore")
	}
	if paths.ShouldSkipDirName("bay", "_pm", nil) {
		t.Fatal("keep bay")
	}
}
