package lexer

import (
	"reflect"
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/diagnostics"
)

func TestPreprocessNestedBlockCommentMarkerIsInert(t *testing.T) {
	t.Parallel()

	// Per SPEC §4.5 block comments do not nest: an inner /* or !* inside a
	// block comment is inert content, never a new block opener. It must not
	// be emitted as an executable line (which previously triggered W1012).
	source := "/*\ninput ~\"run\"\n/* nested opener\ninput ~\"also-run\"\n*/\nignored\n"

	got, err := Preprocess(source)
	if err != nil {
		t.Fatalf("Preprocess() error: %v", err)
	}
	want := []string{`input ~"run"`, `input ~"also-run"`}

	assertLinesEqual(t, got, want)
}

func TestPreprocessNestedWBlockCommentMarkerIsInert(t *testing.T) {
	t.Parallel()

	source := "!*\ninput 1\n!* nested opener\n*!\n"

	got, err := Preprocess(source)
	if err != nil {
		t.Fatalf("Preprocess() error: %v", err)
	}
	want := []string{"input 1"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessBlockCommentOpenerElsewhereStillOpens(t *testing.T) {
	t.Parallel()

	source := "/*\ninput 1\n*/\n// input 2\n"

	got, err := Preprocess(source)
	if err != nil {
		t.Fatalf("Preprocess() error: %v", err)
	}
	want := []string{"input 1", "input 2"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Preprocess() = %v, want %v", got, want)
	}
}

func TestPreprocessUnclosedBlockCommentReturnsW1012(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		opener string
	}{
		{name: "slash-star opener on own line", source: "ignore\n/*\nunclosed code\n", opener: "/*"},
		{name: "slash-star opener with content", source: "/* unclosed", opener: "/*"},
		{name: "bang-star opener on own line", source: "ignore\n!*\nunclosed code\n", opener: "!*"},
		{name: "bang-star opener with content", source: "!* unclosed", opener: "!*"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Preprocess(tc.source)
			if err == nil {
				t.Fatalf("Preprocess() = %v, want error for unclosed block comment", got)
			}
			we, ok := err.(*diagnostics.WorngError)
			if !ok {
				t.Fatalf("error type = %T, want *diagnostics.WorngError", err)
			}
			if we.Diag.Code != diagnostics.UnterminatedBlockComment.Code {
				t.Fatalf("diag code = %d, want %d (W1012)", we.Diag.Code, diagnostics.UnterminatedBlockComment.Code)
			}
			if we.Found != tc.opener {
				t.Fatalf("Found = %q, want %q", we.Found, tc.opener)
			}
		})
	}
}

func TestPreprocessUnclosedBlockCommentStillYieldsLines(t *testing.T) {
	t.Parallel()

	// Lines inside an unclosed block are still extracted (preserve current
	// extraction behaviour); the error signals the missing close marker.
	got, err := Preprocess("ignore\n/*\na\nb\n")
	if err == nil {
		t.Fatal("expected unterminated block comment error")
	}
	if strings.Join(got, ",") != "a,b" {
		t.Fatalf("Preprocess() = %v, want [a, b]", got)
	}
}
