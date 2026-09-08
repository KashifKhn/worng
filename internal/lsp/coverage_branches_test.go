package lsp

import (
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

// TestOperatorAtTwoCharOperators covers the ==, !=, <=, >=, ** hover paths
// (cursor on either half of the pair) that the LSP coverage threshold run
// flagged as unexercised.
func TestOperatorAtTwoCharOperators(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		line     string
		char     int
		wantOp   string
		wantNull bool
	}{
		{name: "on second = of ==", line: "a == b", char: 3, wantOp: "=="},
		{name: "on first ! of !=", line: "a != b", char: 2, wantOp: "!="},
		{name: "on second char of >=", line: "a >= b", char: 3, wantOp: ">="},
		{name: "on <= pair", line: "a <= b", char: 3, wantOp: "<="},
		{name: "on ** pair", line: "a ** b", char: 3, wantOp: "**"},
		{name: "single plus", line: "a + b", char: 2, wantOp: "+"},
		{name: "single minus", line: "a - b", char: 2, wantOp: "-"},
		{name: "open brace", line: "if x }", char: 5, wantOp: "}"},
		{name: "close brace", line: "{", char: 0, wantOp: "{"},
		{name: "non operator letter", line: "abc", char: 1, wantNull: true},
		{name: "comment marker returns empty", line: "// x", char: 1, wantNull: true},
		{name: "bang-bang marker returns empty", line: "!! x", char: 1, wantNull: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			op, rng := operatorAt(tc.line, lsproto.Position{Line: 0, Character: tc.char})
			if tc.wantNull {
				if op != "" {
					t.Fatalf("operatorAt(%q, %d) = %q, want empty", tc.line, tc.char, op)
				}
				if rng.Start.Line != 0 || rng.End.Line != 0 {
					t.Fatalf("range should be zero, got %+v", rng)
				}
				return
			}
			if op != tc.wantOp {
				t.Fatalf("operatorAt(%q, %d) = %q, want %q", tc.line, tc.char, op, tc.wantOp)
			}
		})
	}
}

// TestOperatorAtEdgePositions covers the guard branches: empty line, cursor
// past end of line (clamped to last char), and negative character.
func TestOperatorAtEdgePositions(t *testing.T) {
	t.Parallel()

	t.Run("empty line", func(t *testing.T) {
		op, _ := operatorAt("", lsproto.Position{Line: 0, Character: 0})
		if op != "" {
			t.Fatalf("empty line should give no operator, got %q", op)
		}
	})

	t.Run("cursor clamped to last char", func(t *testing.T) {
		// "a + b" is 5 runes; character 99 clamps onto the 'b' (non-op).
		op, _ := operatorAt("a + b", lsproto.Position{Line: 0, Character: 99})
		if op != "" {
			t.Fatalf("clamped cursor on 'b' should give no operator, got %q", op)
		}
	})

	t.Run("cursor clamped onto trailing operator", func(t *testing.T) {
		// "1 +" is 3 runes; character 99 clamps onto '+'.
		op, _ := operatorAt("1 +", lsproto.Position{Line: 0, Character: 99})
		if op != "+" {
			t.Fatalf("clamped cursor on '+' should give +, got %q", op)
		}
	})

	t.Run("past-end on single space line", func(t *testing.T) {
		// One rune line; character clamps to 0 which is a space → no op.
		op, _ := operatorAt(" ", lsproto.Position{Line: 0, Character: 5})
		if op != "" {
			t.Fatalf("space line should give no operator, got %q", op)
		}
	})
}

// TestDiagRangeBranches covers the clamping branches of diagRange: zero
// positions, end-before-start, and missing end coordinates.
func TestDiagRangeBranches(t *testing.T) {
	t.Parallel()

	t.Run("zero position clamps to 0:0", func(t *testing.T) {
		r := diagRange(diagnostics.Position{})
		if r.Start.Line != 0 || r.Start.Character != 0 {
			t.Fatalf("start = %+v, want 0:0", r.Start)
		}
		// No end info: end falls back to start line, char+1.
		if r.End.Line != 0 || r.End.Character != 1 {
			t.Fatalf("end = %+v, want 0:1", r.End)
		}
	})

	t.Run("negative line and column clamp to zero", func(t *testing.T) {
		r := diagRange(diagnostics.Position{Line: -5, Column: -3, EndLine: -9, EndColumn: -7})
		if r.Start.Line != 0 || r.Start.Character != 0 {
			t.Fatalf("start = %+v, want 0:0", r.Start)
		}
		if r.End.Character != 1 {
			t.Fatalf("end char = %d, want 1 (start+1 fallback)", r.End.Character)
		}
	})

	t.Run("end line before start line clamps up", func(t *testing.T) {
		r := diagRange(diagnostics.Position{Line: 5, Column: 3, EndLine: 2, EndColumn: 9})
		if r.End.Line != r.Start.Line {
			t.Fatalf("end line = %d, want %d (clamped to start)", r.End.Line, r.Start.Line)
		}
	})

	t.Run("same line end char at or before start expands to start+1", func(t *testing.T) {
		r := diagRange(diagnostics.Position{Line: 3, Column: 4, EndLine: 3, EndColumn: 2})
		if r.End.Character != r.Start.Character+1 {
			t.Fatalf("end char = %d, want %d", r.End.Character, r.Start.Character+1)
		}
	})

	t.Run("full envelope passes through", func(t *testing.T) {
		r := diagRange(diagnostics.Position{Line: 2, Column: 5, EndLine: 2, EndColumn: 9})
		if r.Start.Line != 1 || r.Start.Character != 4 || r.End.Character != 9 {
			t.Fatalf("range = %+v, want 1:4-1:9", r)
		}
	})
}

// TestSpecRefURLCoversKeys exercises every diagnostic-key anchor mapping and
// the empty-key and unknown-key fallbacks.
func TestSpecRefURLCoversKeys(t *testing.T) {
	t.Parallel()

	if got := specRefURL(""); !strings.HasSuffix(got, "docs/SPEC.md") || strings.Contains(got, "#") {
		t.Fatalf("empty key = %q, want bare SPEC url", got)
	}
	if got := specRefURL("syntax_error"); !strings.HasSuffix(got, "#execution-model") {
		t.Fatalf("syntax_error = %q", got)
	}
	if got := specRefURL("type_mismatch"); !strings.HasSuffix(got, "#numbers") {
		t.Fatalf("type_mismatch = %q", got)
	}
	if got := specRefURL("infinite_loop"); !strings.HasSuffix(got, "#inversion-rules") {
		t.Fatalf("infinite_loop = %q", got)
	}
	if got := specRefURL("arity_mismatch"); !strings.HasSuffix(got, "#functions") {
		t.Fatalf("arity_mismatch = %q", got)
	}
	if got := specRefURL("invalid_number"); !strings.HasSuffix(got, "#numbers") {
		t.Fatalf("invalid_number = %q", got)
	}
	// Unknown keys fall back to the execution-model anchor.
	if got := specRefURL("totally_unknown"); !strings.HasSuffix(got, "#execution-model") {
		t.Fatalf("unknown key = %q, want execution-model fallback", got)
	}
	if specrefAnchor("") != "" {
		t.Fatal("empty anchor should be empty string")
	}
}

// TestWordAtEdges covers word extraction guards: empty line, cursor on a
// non-word rune, cursor past end (clamped), and multi-byte UTF-16 positions.
func TestWordAtEdges(t *testing.T) {
	t.Parallel()

	t.Run("empty line", func(t *testing.T) {
		w, _ := wordAt("", lsproto.Position{Line: 0, Character: 0})
		if w != "" {
			t.Fatalf("wordAt on empty line = %q, want empty", w)
		}
	})

	t.Run("nonexistent line", func(t *testing.T) {
		w, _ := wordAt("x = 1", lsproto.Position{Line: 9, Character: 0})
		if w != "" {
			t.Fatalf("wordAt on missing line = %q, want empty", w)
		}
	})

	t.Run("cursor on space is not a word", func(t *testing.T) {
		w, _ := wordAt("a b", lsproto.Position{Line: 0, Character: 1})
		if w != "" {
			t.Fatalf("wordAt on space = %q, want empty", w)
		}
	})

	t.Run("cursor clamped to last word char", func(t *testing.T) {
		w, _ := wordAt("abc", lsproto.Position{Line: 0, Character: 99})
		if w != "abc" {
			t.Fatalf("wordAt clamped = %q, want abc", w)
		}
	})

	t.Run("negative character clamps to first word", func(t *testing.T) {
		w, _ := wordAt("abc def", lsproto.Position{Line: 0, Character: -4})
		if w != "abc" {
			t.Fatalf("wordAt negative = %q, want abc", w)
		}
	})

	t.Run("dotted qualified name is one word", func(t *testing.T) {
		w, _ := wordAt("wronglib.sort", lsproto.Position{Line: 0, Character: 4})
		if w != "wronglib.sort" {
			t.Fatalf("wordAt dotted = %q, want wronglib.sort", w)
		}
	})

	t.Run("utf16 surrogate pair offset", func(t *testing.T) {
		// "𝕒bc": 𝕒 is one rune but two UTF-16 units.
		w, _ := wordAt("𝕒bc", lsproto.Position{Line: 0, Character: 2})
		if w != "𝕒bc" {
			t.Fatalf("wordAt surrogate = %q, want 𝕒bc", w)
		}
	})
}

// TestUTF16ConversionHelpers covers the boundary branches of the UTF-16
// index conversion helpers used by hover/rename positions.
func TestUTF16ConversionHelpers(t *testing.T) {
	t.Parallel()

	t.Run("char to rune index boundaries", func(t *testing.T) {
		if utf16CharToRuneIndex("ab", 0) != 0 {
			t.Fatal("char 0 should be rune 0")
		}
		if utf16CharToRuneIndex("ab", -3) != 0 {
			t.Fatal("negative char should clamp to 0")
		}
		if utf16CharToRuneIndex("ab", 2) != 2 {
			t.Fatal("char 2 should be rune 2")
		}
		if utf16CharToRuneIndex("ab", 99) != 2 {
			t.Fatal("past-end char should clamp to len")
		}
		// 𝕒 occupies 2 UTF-16 units: char 1 lands mid-rune → rune index 0.
		if utf16CharToRuneIndex("𝕒b", 1) != 0 {
			t.Fatal("mid-surrogate char should return rune 0")
		}
		if utf16CharToRuneIndex("𝕒b", 3) != 2 {
			t.Fatal("char 3 should be rune 2")
		}
	})

	t.Run("rune index to utf16 boundaries", func(t *testing.T) {
		if runeIndexToUTF16([]rune("ab"), 0) != 0 {
			t.Fatal("idx 0 should be 0 units")
		}
		if runeIndexToUTF16([]rune("ab"), 99) != 2 {
			t.Fatal("past-end idx should clamp to total units")
		}
		if runeIndexToUTF16([]rune("𝕒b"), 2) != 3 {
			t.Fatal("idx 2 should be 3 units (surrogate + 1)")
		}
	})

	t.Run("char to byte index boundaries", func(t *testing.T) {
		if utf16CharToByteIndex("ab", 0) != 0 {
			t.Fatal("char 0 should be byte 0")
		}
		if utf16CharToByteIndex("ab", -2) != 0 {
			t.Fatal("negative char should be byte 0")
		}
		if utf16CharToByteIndex("ab", 99) != 2 {
			t.Fatal("past-end char should be len(line)")
		}
		if utf16CharToByteIndex("𝕒b", 2) != 4 {
			t.Fatal("char 2 should be byte 4 (after 4-byte rune)")
		}
		// Mid-surrogate char 1 lands inside 𝕒 → byte index of its start.
		if utf16CharToByteIndex("𝕒b", 1) != 0 {
			t.Fatal("mid-surrogate char should return rune start byte")
		}
	})
}
