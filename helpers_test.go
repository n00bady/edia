package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatDate(t *testing.T) {
	t.Parallel()

	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty input",
			args: args{input: ""},
			want: "",
		},
		{
			name: "iso date",
			args: args{input: "2026-07-31"},
			want: "2026-07-31", // adjust expected outcome to the repository's helper behavior
		},
	}

	for _, tt := range tests {
		t := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatDate(tt.args.input)
			if got != tt.want {
				t.Fatalf("FormatDate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnsurePathExists(t *testing.T) {
	t.Parallel()
	t.Run("creates and cleans up temporary dir", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "ensurepath_test")
		if err != nil {
			t.Fatalf("MkTemp: %v", err)
		}
		// Clean up whole temp dir at the end of the subtest
		defer os.RemoveAll(tmpDir)

		path := filepath.Join(tmpDir, "subdir", "nested")
		if err := EnsurePathExists(path); err != nil {
			t.Fatalf("EnsurePathExists failed: %v", err)
		}
		if fi, err := os.Stat(path); err != nil {
			t.Fatalf("expected path to exist, stat error: %v", err)
		} else if !fi.IsDir() {
			t.Fatalf("expected a directory at %s, but it's not a dir", path)
		}
	})
}
