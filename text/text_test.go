package text

import "testing"

func TestTruncate(t *testing.T) {
	tests := []struct {
		s     string
		limit int
		want  string
	}{
		{s: "abc", limit: 3, want: "abc"},
		{s: "abcd", limit: 3, want: "abc\n[truncated after 3 bytes]"},
		{s: "aé", limit: 2, want: "a\n[truncated after 1 bytes]"}, // é is 2 bytes; never split it
		{s: "", limit: 0, want: ""},
	}
	for _, tt := range tests {
		if got := Truncate(tt.s, tt.limit); got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.limit, got, tt.want)
		}
	}
}
