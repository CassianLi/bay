package model

import "time"

type Project struct {
	ID            int64
	Name          string
	Root          string
	Path          string
	Status        Status
	PurposeKnown  bool
	Summary       string
	Signals       string
	UserSetStatus bool
	Missing       bool
	LastSeenAt    time.Time
	LastMovedAt   time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ListFilter struct {
	Status      *Status
	UnknownOnly bool
	Root        string
	Query       string
}
