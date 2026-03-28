package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/vfs"
)

func TestPrintDiagnosticsPrettyIncludesSnippetAndCaret(t *testing.T) {
	t.Parallel()

	mem := vfs.NewMemFS()
	if err := mem.WriteFile("bad.wrg", []byte("x =\n")); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	err := diagnostics.NewExpectedToken(
		diagnostics.Position{File: "bad.wrg", Line: 1, Column: 4, EndLine: 1, EndColumn: 4},
		[]string{"identifier"},
		"<eof>",
	)

	var out bytes.Buffer
	printDiagnostics(&out, diagnostics.NewErrorList([]error{err}), mem, "bad.wrg", false)
	got := out.String()

	if !strings.Contains(got, "bad.wrg:1:4: [W1007]") {
		t.Fatalf("output missing location/code: %q", got)
	}
	if !strings.Contains(got, " 1 | x =") {
		t.Fatalf("output missing source snippet: %q", got)
	}
	if !strings.Contains(got, "^") {
		t.Fatalf("output missing caret marker: %q", got)
	}
}

func TestPrintDiagnosticsJSONIncludesStructuredFields(t *testing.T) {
	t.Parallel()

	err := diagnostics.NewTypeMismatch(
		diagnostics.Position{File: "bad.wrg", Line: 3, Column: 2, EndLine: 3, EndColumn: 5},
		[]string{"number"},
		"string",
		"binary operation",
	)

	var out bytes.Buffer
	printDiagnostics(&out, diagnostics.NewErrorList([]error{err}), vfs.NewMemFS(), "bad.wrg", true)
	got := out.String()

	if !strings.Contains(got, `"code": 1002`) {
		t.Fatalf("json missing code: %q", got)
	}
	if !strings.Contains(got, `"key": "type_mismatch"`) {
		t.Fatalf("json missing key: %q", got)
	}
	if !strings.Contains(got, `"expected": [`) {
		t.Fatalf("json missing expected: %q", got)
	}
	if !strings.Contains(got, `"found": "string"`) {
		t.Fatalf("json missing found: %q", got)
	}
}
