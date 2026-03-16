package core

import (
	"strings"
	"testing"
)

func FuzzReverse(f *testing.F) {
	f.Add("")
	f.Add("abc")
	f.Add("héllo🎉")

	f.Fuzz(func(t *testing.T, s string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("reverse panicked: %v", r)
			}
		}()

		r := Reverse(s)
		if Reverse(r) != s {
			t.Fatalf("double reverse mismatch: %q -> %q", s, r)
		}
		if len([]rune(r)) != len([]rune(s)) {
			t.Fatalf("rune length changed: in=%d out=%d", len([]rune(s)), len([]rune(r)))
		}
	})
}

func FuzzContains(f *testing.F) {
	f.Add("hello", "ell")
	f.Add("", "")
	f.Add("abc", "d")

	f.Fuzz(func(t *testing.T, s, substr string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("contains panicked: %v", r)
			}
		}()

		got := Contains(s, substr)
		want := strings.Contains(s, substr)
		if got != want {
			t.Fatalf("Contains(%q, %q) = %v, want %v", s, substr, got, want)
		}
	})
}
