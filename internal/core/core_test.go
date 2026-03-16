package core

import "testing"

// ─── Reverse ────────────────────────────────────────────────────────────────

func TestReverseASCII(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "single_char", input: "a", want: "a"},
		{name: "two_chars", input: "ab", want: "ba"},
		{name: "odd_length", input: "hello", want: "olleh"},
		{name: "even_length", input: "abcd", want: "dcba"},
		{name: "palindrome", input: "racecar", want: "racecar"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Reverse(tc.input)
			if got != tc.want {
				t.Fatalf("Reverse(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestReverseUnicode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "multibyte_runes", input: "héllo", want: "olléh"},
		{name: "emoji", input: "ab🎉cd", want: "dc🎉ba"},
		{name: "japanese", input: "日本語", want: "語本日"},
		{name: "combining_rune", input: "a\u0301b", want: "b\u0301a"}, // á as separate combining char
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Reverse(tc.input)
			if got != tc.want {
				t.Fatalf("Reverse(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestReverseIsItsOwnInverse(t *testing.T) {
	t.Parallel()

	inputs := []string{"", "a", "hello world", "日本語", "🎉🎊"}
	for _, s := range inputs {
		if got := Reverse(Reverse(s)); got != s {
			t.Fatalf("Reverse(Reverse(%q)) = %q, want identity", s, got)
		}
	}
}

// ─── Contains ────────────────────────────────────────────────────────────────

func TestContainsBasic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{name: "empty_substr_always_true", s: "hello", substr: "", want: true},
		{name: "empty_s_empty_substr", s: "", substr: "", want: true},
		{name: "empty_s_nonempty_substr", s: "", substr: "x", want: false},
		{name: "exact_match", s: "hello", substr: "hello", want: true},
		{name: "prefix_match", s: "hello", substr: "hel", want: true},
		{name: "suffix_match", s: "hello", substr: "llo", want: true},
		{name: "mid_match", s: "hello", substr: "ell", want: true},
		{name: "single_char_present", s: "hello", substr: "o", want: true},
		{name: "single_char_absent", s: "hello", substr: "z", want: false},
		{name: "substr_longer_than_s", s: "hi", substr: "hello", want: false},
		{name: "not_found", s: "hello", substr: "xyz", want: false},
		{name: "case_sensitive", s: "Hello", substr: "hello", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Contains(tc.s, tc.substr)
			if got != tc.want {
				t.Fatalf("Contains(%q, %q) = %v, want %v", tc.s, tc.substr, got, tc.want)
			}
		})
	}
}

func TestContainsAtBoundaries(t *testing.T) {
	t.Parallel()

	// substr found at the very last possible position
	s := "abcde"
	if !Contains(s, "cde") {
		t.Fatalf("Contains(%q, %q) = false, want true", s, "cde")
	}
	if Contains(s, "cdf") {
		t.Fatalf("Contains(%q, %q) = true, want false", s, "cdf")
	}
}

// ─── Stack ────────────────────────────────────────────────────────────────────

func TestStackPushAndPop(t *testing.T) {
	t.Parallel()

	var s Stack[int]

	if s.Len() != 0 {
		t.Fatalf("new stack Len = %d, want 0", s.Len())
	}

	s.Push(1)
	s.Push(2)
	s.Push(3)

	if s.Len() != 3 {
		t.Fatalf("Len after 3 pushes = %d, want 3", s.Len())
	}

	v, ok := s.Pop()
	if !ok || v != 3 {
		t.Fatalf("Pop() = (%v, %v), want (3, true)", v, ok)
	}
	v, ok = s.Pop()
	if !ok || v != 2 {
		t.Fatalf("Pop() = (%v, %v), want (2, true)", v, ok)
	}
	v, ok = s.Pop()
	if !ok || v != 1 {
		t.Fatalf("Pop() = (%v, %v), want (1, true)", v, ok)
	}

	if s.Len() != 0 {
		t.Fatalf("Len after all pops = %d, want 0", s.Len())
	}
}

func TestStackPopEmpty(t *testing.T) {
	t.Parallel()

	var s Stack[string]
	v, ok := s.Pop()
	if ok {
		t.Fatalf("Pop on empty stack ok = true, want false")
	}
	if v != "" {
		t.Fatalf("Pop on empty stack value = %q, want zero value", v)
	}
}

func TestStackPeek(t *testing.T) {
	t.Parallel()

	var s Stack[float64]

	// Peek on empty
	v, ok := s.Peek()
	if ok {
		t.Fatalf("Peek on empty ok = true, want false")
	}
	if v != 0 {
		t.Fatalf("Peek on empty value = %v, want zero", v)
	}

	s.Push(3.14)
	s.Push(2.71)

	top, ok := s.Peek()
	if !ok || top != 2.71 {
		t.Fatalf("Peek() = (%v, %v), want (2.71, true)", top, ok)
	}

	// Peek does not remove the element
	if s.Len() != 2 {
		t.Fatalf("Len after Peek = %d, want 2", s.Len())
	}
}

func TestStackPeekDoesNotMutate(t *testing.T) {
	t.Parallel()

	var s Stack[int]
	s.Push(10)

	for range 5 {
		v, ok := s.Peek()
		if !ok || v != 10 {
			t.Fatalf("Peek() = (%v, %v), want (10, true)", v, ok)
		}
	}
	if s.Len() != 1 {
		t.Fatalf("Len after 5 Peeks = %d, want 1", s.Len())
	}
}

func TestStackWithStringType(t *testing.T) {
	t.Parallel()

	var s Stack[string]
	s.Push("first")
	s.Push("second")

	v, ok := s.Pop()
	if !ok || v != "second" {
		t.Fatalf("Pop() = (%q, %v), want (\"second\", true)", v, ok)
	}

	v, ok = s.Peek()
	if !ok || v != "first" {
		t.Fatalf("Peek() = (%q, %v), want (\"first\", true)", v, ok)
	}
}

func TestStackLenIncrementsAndDecrements(t *testing.T) {
	t.Parallel()

	var s Stack[bool]
	for i := range 10 {
		s.Push(i%2 == 0)
		if s.Len() != i+1 {
			t.Fatalf("Len after %d pushes = %d, want %d", i+1, s.Len(), i+1)
		}
	}
	for i := 9; i >= 0; i-- {
		s.Pop()
		if s.Len() != i {
			t.Fatalf("Len after pop = %d, want %d", s.Len(), i)
		}
	}
}
