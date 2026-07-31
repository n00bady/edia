package helpers_test

import (
	"testing"
)

// NOTE:
// Helper function tests use table-driven style and explicit subtests.
// Keep lightweight unit tests fast and mark heavier ones to skip unless explicitly enabled.

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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Replace the following skip with a call to the real helper when ready, e.g.:
			// got := helpers.FormatDate(tt.args.input)
			// if got != tt.want { t.Fatalf("FormatDate() = %q, want %q", got, tt.want) }
			t.Skip("unit test scaffold: wire to actual helper function and enable")
		})
	}
}

func TestEnsurePathExists(t *testing.T) {
	t.Parallel()
	t.Run("creates and cleans up temporary dir", func(t *testing.T) {
		t.Skip("filesystem helper test scaffold: implement with os.MkdirTemp and cleanup or use afero for in-memory FS")
	})
}
