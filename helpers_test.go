package main

import (
	"errors"
	"testing"
)

func TestParseFloatToXDecimals(t *testing.T) {
	tests := []struct {
		n    string
		d    int
		want float64
		err  error
	}{
		{"1.2341454312324432123", 5, 1.23414, errors.New("string too long")},
		{"2.2", 3, 2.200, nil},
		{"5", 3, 5.000, nil},
		{"lol", 6, 0, errors.New("cannot parse float: strconv.ParseFloat: parsing \"lol\": invalid syntax")},
	}

	for _, tc := range tests {
		got, err := ParseFloatToXDecimals(tc.n, tc.d)

		if err != nil && err.Error() != tc.err.Error() {
			t.Errorf("expected error %v, got %v", tc.err, err)
		}

		if err == nil && got != tc.want {
			t.Errorf("expected %v, got %v", tc.want, got)
		}
	}
}

func TestIsValidLatitude(t *testing.T) {
	tests := []struct {
		f    float64
		want bool
		err  error
	}{
		{40.49591, true, nil},
		{-199.23102, false, nil},
		{282, false, nil},
		{1, true, nil},
	}

	for _, tc := range tests {
		got, err := IsValidLatitude(tc.f)

		if err != nil && err.Error() != tc.err.Error() {
			t.Errorf("expected error %v, got %v", tc.err, err)
		}

		if err == nil && got != tc.want {
			t.Errorf("expected %v, got %v", tc.want, got)
		}
	}
}

func TestIsValidLongitude(t *testing.T) {
	tests := []struct {
		f    float64
		want bool
		err  error
	}{
		{23.069222, true, nil},
		{-199.23102, false, nil},
		{-180, true, nil},
		{180, true, nil},
	}

	for _, tc := range tests {
		got, err := IsValidLongitude(tc.f)

		if err != nil && err.Error() != tc.err.Error() {
			t.Errorf("expected error %v, got %v", tc.err, err)
		}

		if err == nil && got != tc.want {
			t.Errorf("expected %v, got %v", tc.want, got)
		}
	}
}

func TestIsNegative(t *testing.T) {
	tests := []struct {
		f    float64
		want bool
		err  error
	}{
		{-2, true, nil},
		{-45892, true, nil},
		{421, false, nil},
		{1.239102, false, nil},
	}

	for _, tc := range tests {
		got, err := IsNegative(tc.f)

		if err != nil && err.Error() != tc.err.Error() {
			t.Errorf("expected error %v, got %v", tc.err, err)
		}

		if err == nil && got != tc.want {
			t.Errorf("expected %v, got %v", tc.want, got)
		}
	}
}

func TestTruncateFloatTo2Decimals(t *testing.T) {
	tests := []struct {
		f float64
		want float64
	}{
		{1, 1.00},
		{45.39102932, 45.39},
		{-32.239581, -32.23},
		{1.2, 1.20},
	}

	for _, tc := range tests {
		got := TruncateFloatTo2Decimals(tc.f)
		if got != tc.want {
			t.Errorf("expected %v, got %v", tc.want, got)
		}
	}
}

// Maybe I should add some tests for the db queries
// I'll probably need mocking stuff and I don't feel like doing it just for that... 
// besides nobody else gonna develop this...
// Anyway the tests are waste of time unless there are a ton of devs working on the same codebase.
