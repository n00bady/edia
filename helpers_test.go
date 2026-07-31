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

func TestParseFloatToXDecimals_RoundsCorrectly(t *testing.T) {
	t.Parallel()

	got, err := ParseFloatToXDecimals("3.14159", 3)
	if err != nil {
		t.Fatalf("unexpected error parsing float: %v", err)
	}
	want := 3.142
	if got != want {
		t.Fatalf("ParseFloatToXDecimals() = %v, want %v", got, want)
	}
}

func TestParseFloatToXDecimals_InvalidInputs(t *testing.T) {
	t.Parallel()

	// empty string should return an error
	if _, err := ParseFloatToXDecimals("", 2); err == nil {
		t.Fatalf("expected error for empty string, got nil")
	}

	// too many decimals (negative) should return an error
	if _, err := ParseFloatToXDecimals("1.23", -1); err == nil {
		t.Fatalf("expected error for negative decimals, got nil")
	}
}

func TestTruncateFloatTo2Decimals_TruncatesDown(t *testing.T) {
	t.Parallel()

	got := TruncateFloatTo2Decimals(3.14159)
	want := 3.14
	if got != want {
		t.Fatalf("TruncateFloatTo2Decimals() = %v, want %v", got, want)
	}

	// negative value
	got = TruncateFloatTo2Decimals(-1.239)
	want = -1.23 // int(-1.239*100) => int(-123.9) => -123 -> -1.23
	if got != want {
		t.Fatalf("TruncateFloatTo2Decimals() with negative = %v, want %v", got, want)
	}
}
