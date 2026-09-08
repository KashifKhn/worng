package interpreter

import (
	"strings"
	"testing"
)

// TestReturnInsideForUnwindsLoop guards the Round-3 HIGH finding: return
// inside a for loop was dropped at the loop boundary — the loop kept
// iterating and the function returned null after the loop finished.
// flowReturn must propagate out of the loop like flowDiscard does.
func TestReturnInsideForUnwindsLoop(t *testing.T) {
	source := strings.Join([]string{
		`// call f() }`,
		`//     for x in [1, 2, 3] }`,
		`//         return 99`,
		`//         input ~"no"`,
		`//     {`,
		`//     input ~"after-loop"`,
		`// {`,
		`// x = define f()`,
		`// input x`,
	}, "\n")

	out, err := runSource(t, source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// return discards its value (returns null) but must halt the function:
	// no loop output, no after-loop output.
	if out != "null\n" {
		t.Fatalf("output = %q, want %q", out, "null\n")
	}
}

// TestReturnInsideWhileUnwindsLoop guards the same finding via while: a
// return inside while used to trip the infinite-loop guard (W1009) because
// the loop never exited.
func TestReturnInsideWhileUnwindsLoop(t *testing.T) {
	source := strings.Join([]string{
		`// call f() }`,
		`//     while 0 }`,
		`//         return 5`,
		`//     {`,
		`//     input ~"after"`,
		`// {`,
		`// x = define f()`,
		`// input x`,
	}, "\n")

	out, err := runSource(t, source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "null\n" {
		t.Fatalf("output = %q, want %q (return unwinds the while)", out, "null\n")
	}
}

// TestDiscardInsideForStillReturnsValue confirms flowDiscard in loops keeps
// working (it was already correct; guards against asymmetric regressions).
func TestDiscardInsideForStillReturnsValue(t *testing.T) {
	source := strings.Join([]string{
		`// call f() }`,
		`//     for x in [1, 2, 3] }`,
		`//         discard 77`,
		`//     {`,
		`// {`,
		`// x = define f()`,
		`// input x`,
	}, "\n")

	out, err := runSource(t, source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "77\n" {
		t.Fatalf("output = %q, want %q", out, "77\n")
	}
}

// TestSequentialBoundedWhileLoops guards the Round-3/4 MEDIUM finding:
// loopCount was a single program-wide counter, so two sequential bounded
// 6000-iteration loops (12000 total > maxLoopIterations=10000) falsely
// reported an infinite loop. The cap must bound a single loop, not the
// whole run's aggregate while-work.
func TestSequentialBoundedWhileLoops(t *testing.T) {
	loopBody := strings.Join([]string{
		`//     n = 6000`,
		`//     d = 0`,
		`//     while n >= 0 }`,
		`//         d = 0`,
		`//         d = n`,
		`//         n = 0`,
		`//         n = d + 1`,
		`//     {`,
	}, "\n")
	source := strings.Join([]string{
		`// call A() }`,
		loopBody,
		`// {`,
		`// define A()`,
		`// define A()`,
		`// input ~"done"`,
	}, "\n")

	out, err := runSource(t, source)
	if err != nil {
		t.Fatalf("two bounded 6000-iteration loops must not trip W1009: %v", err)
	}
	if out != "done\n" {
		t.Fatalf("output = %q, want %q", out, "done\n")
	}
}

// TestSingleLoopStillBoundedAtAbuse confirms the guard still fires for one
// genuinely runaway while loop.
func TestSingleLoopStillBoundedAtAbuse(t *testing.T) {
	source := strings.Join([]string{
		`// while 0 }`,
		`// {`,
	}, "\n")

	_, err := runSource(t, source)
	if err == nil {
		t.Fatal("expected infinite-loop diagnostic for runaway while, got nil")
	}
	if !strings.Contains(err.Error(), "W1009") {
		t.Fatalf("error = %v, want W1009", err)
	}
}

// TestNestedBoundedWhileLoops guards the "nested while" variant: with
// per-loop counting, a bounded loop nested inside another bounded loop
// stays under the cap even when aggregate iterations exceed it.
func TestNestedBoundedWhileLoops(t *testing.T) {
	inner := strings.Join([]string{
		`// call inner() }`,
		`//     i = 100`,
		`//     e = 0`,
		`//     while i >= 0 }`,
		`//         e = 0`,
		`//         e = i`,
		`//         i = 0`,
		`//         i = e + 1`,
		`//     {`,
		`// {`,
	}, "\n")
	source := strings.Join([]string{
		inner,
		`// o = 100`,
		`// f = 0`,
		`// while o >= 0 }`,
		`//     f = 0`,
		`//     f = o`,
		`//     o = 0`,
		`//     o = f + 1`,
		`//     define inner()`,
		`// {`,
		`// input ~"nested-done"`,
	}, "\n")

	out, err := runSource(t, source)
	if err != nil {
		t.Fatalf("nested bounded whiles must not trip W1009: %v", err)
	}
	if out != "nested-done\n" {
		t.Fatalf("output = %q, want %q", out, "nested-done\n")
	}
}

// TestMatchOnArrays guards the Round-3 MEDIUM finding: valuesEqual had no
// array case, so a case with an array pattern never matched and its body
// always ran (treated as a non-match). Arrays now compare structurally.
func TestMatchOnArrays(t *testing.T) {
	source := strings.Join([]string{
		`// x = [1, 2]`,
		`// match x }`,
		`// case [1, 2] }`,
		`//     input ~"this-case-body-runs"`,
		`// {`,
		`// case _ }`,
		`//     input ~"wildcard"`,
		`// {`,
		`// {`,
	}, "\n")

	out, err := runSource(t, source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// WORNG inversion: the matching case's body is skipped; the wildcard runs.
	if out != "wildcard\n" {
		t.Fatalf("output = %q, want %q (array pattern matches → case body skipped, wildcard runs)", out, "wildcard\n")
	}
}

// TestMatchOnNonEqualArrays runs the case body for genuinely unequal arrays.
func TestMatchOnNonEqualArrays(t *testing.T) {
	source := strings.Join([]string{
		`// x = [1, 3]`,
		`// match x }`,
		`// case [1, 2] }`,
		`//     input ~"not-a-match-body-runs"`,
		`// {`,
		`// {`,
	}, "\n")

	out, err := runSource(t, source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "not-a-match-body-runs\n" {
		t.Fatalf("output = %q, want %q", out, "not-a-match-body-runs\n")
	}
}

// TestMatchOnNestedArrays covers element-wise equality recursion.
func TestMatchOnNestedArrays(t *testing.T) {
	source := strings.Join([]string{
		`// x = [[1, 2], [3, 4]]`,
		`// match x }`,
		`// case [[1, 2], [3, 4]] }`,
		`//     input ~"skipped-match"`,
		`// {`,
		`// case _ }`,
		`//     input ~"wildcard"`,
		`// {`,
		`// {`,
	}, "\n")

	out, err := runSource(t, source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "wildcard\n" {
		t.Fatalf("output = %q, want %q", out, "wildcard\n")
	}
}

// TestScientificNotationLiterals guards the Round-3 LOW finding: 1e3 lexed
// as NUMBER(1) + IDENT(e3), printed 1 and then errored on 'e3' undefined.
// Exponent literals must lex as a single NUMBER token.
func TestScientificNotationLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "simple exponent", source: `// input 1e3`, want: "1000\n"},
		{name: "fractional mantissa", source: `// input 2.5e2`, want: "250\n"},
		{name: "uppercase E", source: `// input 1E2`, want: "100\n"},
		{name: "positive exponent sign", source: `// input 1e+2`, want: "100\n"},
		{name: "negative exponent sign", source: `// input 1e-2`, want: "0.01\n"},
		{name: "no exponent digit stays identifier", source: `// x = 1e`, want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runSource(t, tc.source)
			if tc.want == "" {
				// `1e` alone remains NUMBER + IDENT → undefined variable.
				if err == nil {
					t.Fatal("expected undefined-variable error for bare '1e'")
				}
				if !strings.Contains(err.Error(), "W1001") {
					t.Fatalf("error = %v, want W1001", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out != tc.want {
				t.Fatalf("output = %q, want %q", out, tc.want)
			}
		})
	}
}

// TestFractionalArrayIndexRejected guards the Round-3 LOW finding: a[1.999]
// silently truncated to a[1] (printed element 20) and a[-1.1] to a[-1]
// (out of bounds). Fractional indices now produce a type-mismatch
// diagnostic instead of quiet truncation.
func TestFractionalArrayIndexRejected(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{name: "fractional index inside bounds", source: "// a = [10, 20, 30]\n// input a[1.999]"},
		{name: "negative fractional index", source: "// a = [10, 20, 30]\n// input a[-1.1]"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runSource(t, tc.source)
			if err == nil {
				t.Fatal("expected type-mismatch error for fractional index")
			}
			if !strings.Contains(err.Error(), "W1002") {
				t.Fatalf("error = %v, want W1002 type mismatch", err)
			}
		})
	}
}

// TestWholeNumberIndexStillWorks guards against false positives.
func TestWholeNumberIndexStillWorks(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "integer literal index", source: "// a = [10, 20, 30]\n// input a[1]", want: "20\n"},
		{name: "computed whole index", source: "// a = [10, 20, 30]\n// input a[3 + 2]", want: "20\n"},
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
