package model

import (
	"fmt"
	"strings"
)

type Status string

const (
	StatusIdea     Status = "idea"
	StatusActive   Status = "active"
	StatusPaused   Status = "paused"
	StatusKept     Status = "kept"
	StatusArchived Status = "archived"
	StatusTrash    Status = "trash"
)

// AllStatuses returns lifecycle statuses in canonical order.
func AllStatuses() []Status {
	return []Status{
		StatusIdea,
		StatusActive,
		StatusPaused,
		StatusKept,
		StatusArchived,
		StatusTrash,
	}
}

// StatusValuesHint is a short pipe-separated list for CLI help and errors.
func StatusValuesHint() string {
	all := AllStatuses()
	parts := make([]string, len(all))
	for i, st := range all {
		parts[i] = string(st)
	}
	return strings.Join(parts, "|")
}

func ParseStatus(s string) (Status, error) {
	st := Status(s)
	if !st.Valid() {
		return "", fmt.Errorf("invalid status %q (want %s)", s, StatusValuesHint())
	}
	return st, nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusIdea, StatusActive, StatusPaused, StatusKept, StatusArchived, StatusTrash:
		return true
	default:
		return false
	}
}
