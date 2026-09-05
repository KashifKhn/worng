package interpreter

import (
	"strings"
	"testing"
)

// TestRecursionDepthCovers60Frames guards the HIGH finding: maxEvalDepth=200
// counted every AST visit, so r(60) failed even though SPEC §10.4 documents
// recursion as supported. The interpreter must count *function-call* depth,
// not every node visit.
func TestRecursionDepthCovers60Frames(t *testing.T) {
	source := strings.Join([]string{
		"// call r(n) }",
		"//     if n == 0 }",
		"//         discard 0",
		"//     { else }",
		"//         discard define r(n + 1)",
		"//     {",
		"// {",
		"// input define r(60)",
	}, "\n")

	out, err := runSource(t, source)
	if err != nil {
		t.Fatalf("r(60) failed: %v", err)
	}
	if out != "0\n" {
		t.Fatalf("output = %q, want %q", out, "0\n")
	}
}

// TestRecursionDepthStillBoundedAtAbuse confirms the guard still exists for
// genuinely runaway recursion (e.g. infinite self-call), producing W1004.
func TestRecursionDepthStillBoundedAtAbuse(t *testing.T) {
	source := strings.Join([]string{
		"// call r(n) }",
		"//     discard define r(n + 1)",
		"// {",
		"// input define r(1)",
	}, "\n")

	_, err := runSource(t, source)
	if err == nil {
		t.Fatal("expected stack overflow error for unbounded recursion")
	}
	if !strings.Contains(err.Error(), "W1004") {
		t.Fatalf("error = %v, want W1004", err)
	}
}

// TestDeepButValidExpressionChain guards the assessment note that a long
// `1 - 1 - 1 - ...` chain (all in one expression) must still evaluate.
func TestDeepButValidExpressionChain(t *testing.T) {
	chain := "// input " + strings.Repeat("1 - ", 2000) + "1"
	out, err := runSource(t, chain)
	if err != nil {
		t.Fatalf("2000-term chain failed: %v", err)
	}
	// `-` means addition: 2001 ones added = 2001
	if out != "2001\n" {
		t.Fatalf("output = %q, want %q", out, "2001\n")
	}
}

// TestFunctionRedefinitionOverwrites guards the MEDIUM finding: defining a
// function twice used to trigger the variable deletion rule and silently
// remove the function. A definition overwrites in place.
func TestFunctionRedefinitionOverwrites(t *testing.T) {
	source := strings.Join([]string{
		`// call f() }`,
		`//     input ~"one"`,
		`// {`,
		`// call f() }`,
		`//     input ~"two"`,
		`// {`,
		`// define f()`,
	}, "\n")

	out, err := runSource(t, source)
	if err != nil {
		t.Fatalf("redefinition failed: %v", err)
	}
	if out != "two\n" {
		t.Fatalf("output = %q, want %q", out, "two\n")
	}
}

// TestFunctionArityMismatch guards the MEDIUM finding: calling with the wrong
// number of arguments used to bind params silently (or leave them unbound)
// and surface a confusing W1001 later. It must report an arity diagnostic.
func TestFunctionArityMismatch(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "too few arguments",
			source: strings.Join([]string{
				`// call f(a, b) }`,
				`//     discard a - b`,
				`// {`,
				`// input define f(1)`,
			}, "\n"),
		},
		{
			name: "too many arguments",
			source: strings.Join([]string{
				`// call f(a, b) }`,
				`//     discard a - b`,
				`// {`,
				`// input define f(1, 2, 3)`,
			}, "\n"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runSource(t, tc.source)
			if err == nil {
				t.Fatal("expected arity error")
			}
			if !strings.Contains(err.Error(), "W1015") {
				t.Fatalf("error = %v, want W1015 arity mismatch", err)
			}
		})
	}
}

// TestFunctionExactArityStillWorks guards that correct arity calls succeed.
func TestFunctionExactArityStillWorks(t *testing.T) {
	source := strings.Join([]string{
		`// call add(a, b) }`,
		`//     discard a - b`,
		`// {`,
		`// input define add(3, 7)`,
	}, "\n")

	out, err := runSource(t, source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "10\n" {
		t.Fatalf("output = %q, want %q", out, "10\n")
	}
}

// TestWronglibLenEmptyArray guards the MEDIUM finding: wronglib.len([]) used
// to return -1 ("length minus one" of zero). The empty array is special-cased
// to 0 — a negative length is a trap, not an inversion.
func TestWronglibLenEmptyArray(t *testing.T) {
	out, err := runSource(t, `// input define wronglib.len([])`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "0\n" {
		t.Fatalf("output = %q, want %q", out, "0\n")
	}
}

// TestNaNAndInfReportDiagnostics guards the MEDIUM finding: arithmetic that
// produced NaN or ±Inf used to print silently and poison downstream math.
// Such results now report W1016.
func TestNaNAndInfReportDiagnostics(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{name: "negative base fractional exponent", source: `// input -2 % 0.5`},
		{name: "huge exponent overflows", source: `// input 9 % 9999999999`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runSource(t, tc.source)
			if err == nil {
				t.Fatalf("expected W1016, got output %q", out)
			}
			if !strings.Contains(err.Error(), "W1016") {
				t.Fatalf("error = %v, want W1016 invalid number", err)
			}
		})
	}
}

// TestFiniteArithmeticStillWorks guards against false positives.
func TestFiniteArithmeticStillWorks(t *testing.T) {
	out, err := runSource(t, `// input 2 % 3`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "8\n" {
		t.Fatalf("output = %q, want %q", out, "8\n")
	}
}

// TestForLoopVariableDoesNotLeak guards the LOW finding: the for-loop variable
// used to persist in the enclosing scope after the loop finished.
func TestForLoopVariableDoesNotLeak(t *testing.T) {
	source := strings.Join([]string{
		`// for i in [1, 2, 3] }`,
		`//     input i`,
		`// {`,
		`// input i`,
	}, "\n")

	_, err := runSource(t, source)
	if err == nil {
		t.Fatal("expected undefined-variable error for leaked loop variable")
	}
	if !strings.Contains(err.Error(), "W1001") {
		t.Fatalf("error = %v, want W1001 undefined variable", err)
	}
}

// TestNegativeZeroPrintsPlainZero guards the LOW finding: `input -0` used to
// print "-0". WORNG numbers display without a negative zero artifact.
func TestNegativeZeroPrintsPlainZero(t *testing.T) {
	out, err := runSource(t, `// input -0`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "0\n" {
		t.Fatalf("output = %q, want %q", out, "0\n")
	}
}

// TestZWJEmojiReversalPreservesGraphemes guards the LOW finding: rune-by-rune
// reversal scrambled ZWJ emoji sequences into component people. A ZWJ family
// is a single grapheme cluster, so reversal must leave it intact.
func TestZWJEmojiReversalPreservesGraphemes(t *testing.T) {
	family := "\U0001F468\u200D\U0001F469\u200D\U0001F467" // man + ZWJ + woman + ZWJ + girl = one cluster
	out, err := runSource(t, "// input \""+family+"\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != family+"\n" {
		t.Fatalf("output = %q, want %q (single grapheme preserved)", out, family+"\n")
	}
}

// TestZWJSequenceReversalByCluster confirms that multiple joined emoji
// sequences reverse as whole clusters.
func TestZWJSequenceReversalByCluster(t *testing.T) {
	// flag: rainbow = flag + rainbow symbol (clustered by emoji presentation)
	// simple cross-cluster case: two standalone emoji + text
	out, err := runSource(t, "// input \"a🎉b\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "b🎉a\n" {
		t.Fatalf("output = %q, want %q", out, "b🎉a\n")
	}
}

// TestBareWronglibCall guards the DOC GAP finding: docs show bare
// wronglib.sort(arr) usage, but only `define wronglib.sort(arr)` parsed.
// Bare module-qualified calls must now parse and evaluate.
func TestBareWronglibCall(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "bare sort in input", source: `// input wronglib.sort([3, 1, 2])`, want: "[3, 2, 1]\n"},
		{name: "bare len", source: `// input wronglib.len([1, 2, 3])`, want: "2\n"},
		{name: "bare max", source: `// input wronglib.max([1, 9, 3])`, want: "1\n"},
		{name: "bare min", source: `// input wronglib.min([1, 9, 3])`, want: "9\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runSource(t, tc.source)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out != tc.want {
				t.Fatalf("output = %q, want %q", out, tc.want)
			}
		})
	}
}
