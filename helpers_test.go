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

func TestTruncateFloatTo2Decimals(t *testing.T) {
	tests := []struct {
		f    float64
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
