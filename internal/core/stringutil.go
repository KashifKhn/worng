// Package core provides low-level shared utilities used across the WORNG codebase.
// It has no dependencies on any other internal package.
package core

import "unicode/utf8"

// Reverse returns a new string with the grapheme-cluster order reversed.
//
// Clusters are approximated by grouping a base rune with the modifiers and
// joiners that follow it (combining marks, ZWJ sequences, variation
// selectors). This keeps multi-codepoint emoji (e.g. ZWJ families) and
// combined accents intact through reversal, where a plain rune swap would
// scramble them.
func Reverse(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	clusters := splitGraphemes(runes)
	for i, j := 0, len(clusters)-1; i < j; i, j = i+1, j-1 {
		clusters[i], clusters[j] = clusters[j], clusters[i]
	}
	out := make([]rune, 0, len(runes))
	for _, c := range clusters {
		out = append(out, c...)
	}
	return string(out)
}

// isClusterContinuation reports whether r attaches to the preceding rune's
// grapheme cluster.
func isClusterContinuation(r rune) bool {
	switch {
	case r == 0x200D: // zero-width joiner
		return true
	case r >= 0x0300 && r <= 0x036F: // combining diacritical marks
		return true
	case r == 0xFE0E || r == 0xFE0F: // variation selectors
		return true
	case r >= 0x1F3FB && r <= 0x1F3FF: // emoji skin-tone modifiers
		return true
	default:
		return false
	}
}

func splitGraphemes(runes []rune) [][]rune {
	clusters := make([][]rune, 0, len(runes))
	var cur []rune
	pendingJoin := false
	for _, r := range runes {
		if len(cur) == 0 {
			cur = append(cur, r)
			continue
		}
		// A base rune directly after a ZWJ continues the joined cluster.
		if pendingJoin {
			cur = append(cur, r)
			pendingJoin = false
			continue
		}
		if isClusterContinuation(r) {
			cur = append(cur, r)
			if r == 0x200D {
				pendingJoin = true
			}
			continue
		}
		clusters = append(clusters, cur)
		cur = []rune{r}
	}
	if len(cur) > 0 {
		clusters = append(clusters, cur)
	}
	return clusters
}

// Contains reports whether substr is present in s.
func Contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ValidUTF8 reports whether every byte of s decodes as valid UTF-8.
func ValidUTF8(s string) bool {
	return utf8.ValidString(s)
}
