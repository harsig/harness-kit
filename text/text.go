// Package text holds helpers for text handed to a model.
package text

import (
	"fmt"
	"unicode/utf8"
)

// Truncate cuts s to at most limit bytes, on a UTF-8 character boundary, and
// marks the cut so the model knows output is missing.
func Truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	for limit > 0 && !utf8.RuneStart(s[limit]) {
		limit--
	}
	return fmt.Sprintf("%s\n[truncated after %d bytes]", s[:limit], limit)
}
