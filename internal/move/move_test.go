package move_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jinrunli/bay/internal/model"
	"github.com/jinrunli/bay/internal/move"
)

func TestPlanAndApplyArchive(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "foo")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	plan, err := move.PlanMove(root, "foo", "_pm", src, model.StatusArchived)
	if err != nil || !plan.NeedMove {
		t.Fatalf("%+v %v", plan, err)
	}
	if err := move.Apply(plan); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(plan.To); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("src still exists")
	}
}

func TestApplyRefusesExistingTarget(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "foo")
	dst := filepath.Join(root, "_pm", "trash", "foo")
	os.MkdirAll(src, 0o755)
	os.MkdirAll(dst, 0o755)
	plan, _ := move.PlanMove(root, "foo", "_pm", src, model.StatusTrash)
	if err := move.Apply(plan); err == nil {
		t.Fatal("expected error")
	}
}
