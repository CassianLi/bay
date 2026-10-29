package store_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/CassianLi/bay/internal/model"
	"github.com/CassianLi/bay/internal/store"
)

func TestUpsertScanDoesNotClobberSummaryOrLockedStatus(t *testing.T) {
	db := filepath.Join(t.TempDir(), "ledger.db")
	s, err := store.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	now := time.Now().UTC()
	p := model.Project{
		Name: "foo", Root: "/r", Path: "/r/foo", Status: model.StatusPaused,
		Signals: `{"n":1}`, LastSeenAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.UpsertScan(p); err != nil {
		t.Fatal(err)
	}
	p.Summary = "manual"
	p.PurposeKnown = true
	p.Status = model.StatusActive
	p.UserSetStatus = true
	if err := s.UpdateMeta(p); err != nil {
		t.Fatal(err)
	}
	p2 := p
	p2.Summary = "should-not-write"
	p2.Status = model.StatusTrash
	p2.UserSetStatus = false
	p2.Signals = `{"n":2}`
	p2.Path = "/r/foo"
	if err := s.UpsertScan(p2); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetByName("foo", "/r")
	if err != nil {
		t.Fatal(err)
	}
	if got.Summary != "manual" || got.Status != model.StatusActive || !got.PurposeKnown {
		t.Fatalf("%+v", got)
	}
	if got.Signals != `{"n":2}` {
		t.Fatalf("signals not updated: %s", got.Signals)
	}
}

func TestListFiltersAndAmbiguous(t *testing.T) {
	db := filepath.Join(t.TempDir(), "ledger.db")
	s, err := store.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Now().UTC()
	_ = s.UpsertScan(model.Project{Name: "dup", Root: "/a", Path: "/a/dup", Status: model.StatusPaused, LastSeenAt: now, CreatedAt: now, UpdatedAt: now})
	_ = s.UpsertScan(model.Project{Name: "dup", Root: "/b", Path: "/b/dup", Status: model.StatusKept, PurposeKnown: false, LastSeenAt: now, CreatedAt: now, UpdatedAt: now})
	if _, err := s.GetByName("dup", ""); err != store.ErrAmbiguous {
		t.Fatalf("want ambiguous, got %v", err)
	}
	st := model.StatusKept
	list, err := s.List(model.ListFilter{Status: &st})
	if err != nil || len(list) != 1 {
		t.Fatalf("%v %d", err, len(list))
	}
}

func TestMarkMissingExcept(t *testing.T) {
	db := filepath.Join(t.TempDir(), "ledger.db")
	s, err := store.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	now := time.Now().UTC()
	if err := s.UpsertScan(model.Project{Name: "keep", Root: "/r", Path: "/r/keep", Status: model.StatusActive, LastSeenAt: now, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertScan(model.Project{Name: "gone", Root: "/r", Path: "/r/gone", Status: model.StatusPaused, LastSeenAt: now, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertScan(model.Project{Name: "other", Root: "/x", Path: "/x/other", Status: model.StatusActive, LastSeenAt: now, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}

	n, err := s.MarkMissingExcept("/r", []string{"keep"})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("want 1 marked, got %d", n)
	}

	gone, err := s.GetByName("gone", "/r")
	if err != nil {
		t.Fatal(err)
	}
	if !gone.Missing {
		t.Fatalf("gone should be missing: %+v", gone)
	}
	keep, err := s.GetByName("keep", "/r")
	if err != nil {
		t.Fatal(err)
	}
	if keep.Missing {
		t.Fatalf("keep should not be missing: %+v", keep)
	}
	other, err := s.GetByName("other", "/x")
	if err != nil {
		t.Fatal(err)
	}
	if other.Missing {
		t.Fatalf("other root should be untouched: %+v", other)
	}
}
