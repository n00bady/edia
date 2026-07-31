package main

import (
	"testing"
)

func TestNewEntryWithLabel_ReturnsEntry(t *testing.T) {
	t.Parallel()

	e := newEntryWithLabel("placeholder")
	if e == nil {
		t.Fatalf("expected non-nil *widget.Entry, got nil")
	}
}

func TestNewFilteredEntry_SetsOnChanged(t *testing.T) {
	t.Parallel()

	e := NewFilteredEntry("[^0-9]", "digits only")
	if e == nil {
		t.Fatalf("expected non-nil *widget.Entry, got nil")
	}
	if e.OnChanged == nil {
		t.Fatalf("expected OnChanged handler to be set on filtered entry")
	}
}
