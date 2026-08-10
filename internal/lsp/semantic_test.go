package lsp

import (
	"sort"
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/ast"
	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func TestSemanticDiagnosticsNilProgram(t *testing.T) {
	t.Parallel()

	errs := semanticDiagnostics(nil, "file:///x.wrg")
	if len(errs) != 0 {
		t.Fatalf("nil program errors = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsEmptyProgram(t *testing.T) {
	t.Parallel()

	program := &ast.ProgramNode{Statements: nil}
	errs := semanticDiagnostics(program, "file:///x.wrg")
	if len(errs) != 0 {
		t.Fatalf("empty program errors = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsUndefinedVariable(t *testing.T) {
	t.Parallel()

	text := "// input x\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 1 {
		t.Fatalf("errs = %d, want 1", len(errs))
	}
	we, ok := errs[0].(*diagnostics.WorngError)
	if !ok {
		t.Fatalf("err type = %T, want *WorngError", errs[0])
	}
	if we.Diag.Code != diagnostics.UndefinedVariable.Code {
		t.Fatalf("code = %d, want %d", we.Diag.Code, diagnostics.UndefinedVariable.Code)
	}
	if strings.TrimSpace(we.Detail) == "" {
		t.Fatal("expected detail on undefined variable error")
	}
	if strings.TrimSpace(we.Hint) == "" {
		t.Fatal("expected hint on undefined variable error")
	}
}

func TestSemanticDiagnosticsDefinedVariableNoErrors(t *testing.T) {
	t.Parallel()

	text := "// a = 1\n// input a\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected error: %v", e)
		}
		t.Fatalf("defined variable errs = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsBuiltinsNotReported(t *testing.T) {
	t.Parallel()

	tests := []string{
		"// input true\n",
		"// input false\n",
		"// input null\n",
		"// export wronglib\n// input define wronglib.len([1,2])\n",
	}
	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			parsed := parseProgram("file:///a.wrg", text)
			errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
			if len(errs) != 0 {
				for _, e := range errs {
					t.Logf("unexpected: %v", e)
				}
				t.Fatalf("builtin errs = %d, want 0", len(errs))
			}
		})
	}
}

func TestSemanticDiagnosticsMultipleUndefined(t *testing.T) {
	t.Parallel()

	text := "// input x\n// input y\n// input z\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 3 {
		t.Fatalf("errs = %d, want 3", len(errs))
	}
	names := make([]string, len(errs))
	for i, e := range errs {
		we := e.(*diagnostics.WorngError)
		names[i] = we.Args[0]
	}
	sort.Strings(names)
	if names[0] != "x" || names[1] != "y" || names[2] != "z" {
		t.Fatalf("names = %v, want [x, y, z]", names)
	}
}

func TestSemanticDiagnosticsOnlyReportsOncePerName(t *testing.T) {
	t.Parallel()

	text := "// input x\n// input x\n// input x\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 1 {
		t.Fatalf("duplicate errs = %d, want 1", len(errs))
	}
}

func TestSemanticDiagnosticsFunctionParams(t *testing.T) {
	t.Parallel()

	text := "// call add(a, b) }\n//   discard a\n// {\n// define add(1, 2)\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected: %v", e)
		}
		t.Fatalf("function param errs = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsForLoopVariable(t *testing.T) {
	t.Parallel()

	text := "// for item in [1, 2, 3] }\n//   input item\n// {\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected: %v", e)
		}
		t.Fatalf("for-loop var errs = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsIfElseScope(t *testing.T) {
	t.Parallel()

	text := "// x = 1\n// if x }\n//   input x\n" +
		"// { else }\n//   y = 2\n" +
		"//   input y\n// {\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected: %v", e)
		}
		t.Fatalf("if-else errs = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsWhileBody(t *testing.T) {
	t.Parallel()

	text := "// i = 0\n// while i }\n//   input i\n// {\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected: %v", e)
		}
		t.Fatalf("while body errs = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsExportDefinesModule(t *testing.T) {
	t.Parallel()

	text := "// export wronglib\n// input define wronglib.len([1,2])\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected: %v", e)
		}
		t.Fatalf("export errs = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsDelDefinesVariable(t *testing.T) {
	t.Parallel()

	text := "// del score\n// input score\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected: %v", e)
		}
		t.Fatalf("del errs = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsScopeStatement(t *testing.T) {
	t.Parallel()

	text := "// global x\n// input x\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected: %v", e)
		}
		t.Fatalf("scope errs = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsTryExceptErrorVar(t *testing.T) {
	t.Parallel()

	text := "// try }\n// input 1\n// { except(err) }\n" +
		"//   input err\n// {\n"
	parsed := parseProgram("file:///a.wrg", text)
	if parsed.program == nil || len(parsed.program.Statements) == 0 {
		t.Skip("parser failed; cannot test semantic")
	}
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	for _, e := range errs {
		we := e.(*diagnostics.WorngError)
		if we.Args[0] == "err" {
			t.Fatalf("except error var 'err' should be defined: %v", e)
		}
	}
}

func TestSemanticDiagnosticsUndefinedModuleInCall(t *testing.T) {
	t.Parallel()

	text := "// input define missinglib.func()\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	found := false
	for _, e := range errs {
		we := e.(*diagnostics.WorngError)
		if we.Args[0] == "missinglib" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected undefined module 'missinglib' in qualified call, got %v", errs)
	}
}

func TestSemanticDiagnosticsMatchCaseBodies(t *testing.T) {
	t.Parallel()

	text := "// match [1, 2] }\n//   case [1, 2] }\n" +
		"//     input 0\n//   {\n" +
		"//   case _ }\n//     r = 1\n//     input r\n//   {\n// {\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected: %v", e)
		}
		t.Fatalf("match case errs = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsBinaryAndUnaryExpressions(t *testing.T) {
	t.Parallel()

	text := "// a = 1\n// b = 2\n// input a + b\n// input not a\n// input is a\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected: %v", e)
		}
		t.Fatalf("binary/unary errs = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsArrayLiterals(t *testing.T) {
	t.Parallel()

	text := "// x = 1\n// y = 2\n// input [x, y, 3]\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected: %v", e)
		}
		t.Fatalf("array literal errs = %d, want 0", len(errs))
	}
}

func TestSemanticDiagnosticsUndefinedInArrayLiteral(t *testing.T) {
	t.Parallel()

	text := "// input [z]\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 1 {
		t.Fatalf("errs = %d, want 1", len(errs))
	}
	we := errs[0].(*diagnostics.WorngError)
	if we.Args[0] != "z" {
		t.Fatalf("name = %q, want z", we.Args[0])
	}
}

func TestInferExprTypeTableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr ast.Expression
		want string
	}{
		{name: "nil", expr: nil, want: "unknown"},
		{name: "number", expr: &ast.NumberLiteral{}, want: "number"},
		{name: "string", expr: &ast.StringLiteral{}, want: "string"},
		{name: "bool", expr: &ast.BoolLiteral{}, want: "bool"},
		{name: "null", expr: &ast.NullLiteral{}, want: "null"},
		{name: "array", expr: &ast.ArrayLiteral{}, want: "array"},
		{name: "funcCall", expr: &ast.FuncCallNode{}, want: "function"},
		{name: "binary", expr: &ast.BinaryNode{}, want: "expression"},
		{name: "unary", expr: &ast.UnaryNode{}, want: "expression"},
		{name: "ident (unknown)", expr: &ast.IdentNode{}, want: "unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := inferExprType(tc.expr)
			if got != tc.want {
				t.Fatalf("inferExprType = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSpecrefAnchorTableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  string
		want string
	}{
		{key: "syntax_error", want: "execution-model"},
		{key: "undefined_variable", want: "execution-model"},
		{key: "type_mismatch", want: "numbers"},
		{key: "division_by_zero", want: "numbers"},
		{key: "stack_overflow", want: "execution-model"},
		{key: "index_out_of_bounds", want: "execution-model"},
		{key: "module_not_found", want: "execution-model"},
		{key: "file_not_found", want: "execution-model"},
		{key: "infinite_loop", want: "inversion-rules"},
		{key: "nonexistent_key", want: "execution-model"},
		{key: "", want: ""},
	}
	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()
			got := specrefAnchor(tc.key)
			if got != tc.want {
				t.Fatalf("specrefAnchor(%q) = %q, want %q", tc.key, got, tc.want)
			}
		})
	}
}

func TestSpecRefURLTableDriven(t *testing.T) {
	t.Parallel()

	base := "https://github.com/KashifKhn/worng/blob/main/docs/SPEC.md"

	tests := []struct {
		name string
		key  string
		want string
	}{
		{name: "empty key returns base", key: "", want: base},
		{name: "whitespace key returns base", key: "   ", want: base},
		{name: "known key includes anchor", key: "type_mismatch", want: base + "#numbers"},
		{name: "unknown key includes default", key: "xyz", want: base + "#execution-model"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := specRefURL(tc.key)
			if got != tc.want {
				t.Fatalf("specRefURL(%q) = %q, want %q", tc.key, got, tc.want)
			}
		})
	}
}

func TestDiagRangeEndColumnEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pos  diagnostics.Position
	}{
		{name: "zero end column falls back", pos: diagnostics.Position{Line: 1, Column: 5, EndLine: 1, EndColumn: 0}},
		{name: "negative end line", pos: diagnostics.Position{Line: 1, Column: 3, EndLine: -1, EndColumn: 5}},
		{name: "end line less than start", pos: diagnostics.Position{Line: 5, Column: 3, EndLine: 2, EndColumn: 7}},
		{name: "same line same column", pos: diagnostics.Position{Line: 1, Column: 3, EndLine: 1, EndColumn: 3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rng := diagRange(tc.pos)
			if rng.End.Line < rng.Start.Line {
				t.Fatalf("end line < start: %#v", rng)
			}
			if rng.End.Line == rng.Start.Line && rng.End.Character <= rng.Start.Character {
				t.Fatalf("end char <= start char: %#v", rng)
			}
		})
	}
}

func TestKeywordDocsCoverage(t *testing.T) {
	t.Parallel()

	s := NewServer()
	// All keywords() should have hover docs.
	for _, kw := range keywords() {
		if _, ok := s.keywordDoc[kw]; !ok {
			t.Errorf("keyword %q has no hover doc", kw)
		}
	}
}

func TestOperatorDocsCoverage(t *testing.T) {
	t.Parallel()

	s := NewServer()
	expectedOps := []string{"+", "-", "*", "/", "%", "**", "==", "!=", ">", "<", ">=", "<=", "{", "}"}
	for _, op := range expectedOps {
		if _, ok := s.operatorDoc[op]; !ok {
			t.Errorf("operator %q has no hover doc", op)
		}
	}
}

func TestWronglibDocsCoverage(t *testing.T) {
	t.Parallel()

	s := NewServer()
	expected := []string{"len", "max", "min", "sort", "abs"}
	for _, name := range expected {
		if _, ok := s.wronglibDoc[name]; !ok {
			t.Errorf("wronglib.%s has no hover doc", name)
		}
	}
}

func TestRenderHoverDocRendersAllFields(t *testing.T) {
	t.Parallel()

	d := hoverDoc{
		Title:   "`+` operator",
		Written: "Addition.",
		Actual:  "Subtraction.",
		Gotcha:  "String + is not concatenation.",
		Example: "// input 10 + 3",
		SpecRef: "docs/SPEC.md",
	}
	got := renderHoverDoc(d)
	for _, want := range []string{d.Title, d.Written, d.Actual, d.Gotcha, d.Example, d.SpecRef} {
		if !strings.Contains(got, want) {
			t.Errorf("renderHoverDoc missing %q in: %s", want, got[:min(len(got), 200)])
		}
	}
}

func TestRenderHoverDocMinimal(t *testing.T) {
	t.Parallel()

	d := hoverDoc{Title: "Minimal", Actual: "Does something."}
	got := renderHoverDoc(d)
	if !strings.Contains(got, "Minimal") || !strings.Contains(got, "Does something") {
		t.Fatalf("renderHoverDoc minimal = %q", got)
	}
	if strings.Contains(got, "Gotcha") || strings.Contains(got, "Example") {
		t.Fatalf("minimal doc should not have Gotcha/Example: %s", got)
	}
}

func TestHoverOnUnknownIdentifier(t *testing.T) {
	t.Parallel()

	s := NewServer()
	s.docs["file:///id.wrg"] = &document{
		uri:  "file:///id.wrg",
		text: "// some_unknown_var\n",
	}
	h := s.hover(lsproto.TextDocumentPositionParams{
		TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///id.wrg"},
		Position:     lsproto.Position{Line: 0, Character: 3},
	})
	if !strings.Contains(h.Contents.Value, "some_unknown_var") {
		t.Fatalf("hover = %q, want some_unknown_var", h.Contents.Value)
	}
}

func TestDiagnosticMessageWithDetailAndHint(t *testing.T) {
	t.Parallel()

	we := diagnostics.New(diagnostics.SyntaxError, diagnostics.Position{Line: 1, Column: 1})
	we.Detail = "something broke"
	we.Hint = "try this instead"
	msg := diagnosticMessage(we)
	if !strings.Contains(msg, we.Message()) {
		t.Fatalf("message missing base: %q", msg)
	}
	if !strings.Contains(msg, we.Detail) {
		t.Fatalf("message missing detail: %q", msg)
	}
	if !strings.Contains(msg, we.Hint) {
		t.Fatalf("message missing hint: %q", msg)
	}
}

func TestDiagnosticDataWithExpected(t *testing.T) {
	t.Parallel()

	we := diagnostics.New(diagnostics.SyntaxError, diagnostics.Position{Line: 1, Column: 1})
	we.Expected = []string{"}"}
	we.Found = "<eof>"
	data := diagnosticData(we)
	if _, ok := data["expected"]; !ok {
		t.Fatal("data missing expected")
	}
	if len(data["expected"].([]string)) != 1 || data["expected"].([]string)[0] != "}" {
		t.Fatalf("expected = %v, want [}]", data["expected"])
	}
	if data["found"] != "<eof>" {
		t.Fatalf("found = %v, want <eof>", data["found"])
	}
}

func TestDiagnosticDataMinimal(t *testing.T) {
	t.Parallel()

	we := diagnostics.New(diagnostics.SyntaxError, diagnostics.Position{Line: 1, Column: 1})
	data := diagnosticData(we)
	if data["key"] != "syntax_error" {
		t.Fatalf("key = %v, want syntax_error", data["key"])
	}
	if _, ok := data["hint"]; ok {
		t.Fatal("no hint should mean no hint key in data")
	}
	if _, ok := data["detail"]; ok {
		t.Fatal("no detail should mean no detail key in data")
	}
}

func TestWronglibNamespaceHover(t *testing.T) {
	t.Parallel()

	s := NewServer()
	result := wronglibNamespaceHover(s)
	if !strings.Contains(result, "wronglib") {
		t.Fatalf("wronglib namespace hover = %q, want wronglib", result)
	}
	for _, name := range []string{"len", "max", "min", "sort", "abs"} {
		if !strings.Contains(result, name) {
			t.Fatalf("wronglib namespace hover missing %q: %s", name, result)
		}
	}
}

func TestCollectDefinesNestedStructures(t *testing.T) {
	t.Parallel()

	text := "// call outer(x) }\n" +
		"//   w = 1\n" +
		"//   call inner(y) }\n" +
		"//     input w\n" +
		"//     input x\n" +
		"//     input y\n" +
		"//   {\n" +
		"// {\n" +
		"// define outer(99)\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			we := e.(*diagnostics.WorngError)
			t.Logf("unexpected: name=%s detail=%s", we.Args[0], we.Detail)
		}
		t.Fatalf("nested structures errs = %d, want 0", len(errs))
	}
}

func TestCheckExprReferencesPrintNode(t *testing.T) {
	t.Parallel()

	text := "// input print x\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 1 {
		t.Fatalf("print node errs = %d, want 1", len(errs))
	}
	we := errs[0].(*diagnostics.WorngError)
	if we.Args[0] != "x" {
		t.Fatalf("name = %q, want x", we.Args[0])
	}
}

func TestCheckExprReferencesDiscardAndReturnNodes(t *testing.T) {
	t.Parallel()

	text := "// call test(a) }\n" +
		"//   discard a\n" +
		"//   return a\n" +
		"// {\n"
	parsed := parseProgram("file:///a.wrg", text)
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Logf("unexpected: %v", e)
		}
		t.Fatalf("discard/return errs = %d, want 0", len(errs))
	}
}

func TestCheckExprReferencesRaiseNode(t *testing.T) {
	t.Parallel()

	text := "// raise E(~\"msg\")\n"
	parsed := parseProgram("file:///a.wrg", text)
	if parsed.program == nil || len(parsed.program.Statements) == 0 {
		t.Skip("parser failed to produce raise node")
	}
	errs := semanticDiagnostics(parsed.program, "file:///a.wrg")
	// E is a raise error name, not a variable reference — should not flag
	found := false
	for _, e := range errs {
		we := e.(*diagnostics.WorngError)
		if we.Args[0] == "E" {
			found = true
		}
	}
	if found {
		t.Fatal("raise error name should not be flagged as undefined variable")
	}
}

func TestHoverOnDefinedVariableShowsType(t *testing.T) {
	t.Parallel()

	s := NewServer()
	s.docs["file:///vt.wrg"] = &document{
		uri:  "file:///vt.wrg",
		text: "// x = 42\n// input x\n",
	}
	s.reindexDoc("file:///vt.wrg", "// x = 42\n// input x\n")

	h := s.hover(lsproto.TextDocumentPositionParams{
		TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///vt.wrg"},
		Position:     lsproto.Position{Line: 1, Character: 9},
	})
	if !strings.Contains(h.Contents.Value, "number") {
		t.Fatalf("hover = %q, want 'number'", h.Contents.Value)
	}
}

func TestHoverOnStringVariableShowsStringType(t *testing.T) {
	t.Parallel()

	s := NewServer()
	s.docs["file:///vs.wrg"] = &document{
		uri:  "file:///vs.wrg",
		text: "// name = ~\"hello\"\n// input name\n",
	}
	s.reindexDoc("file:///vs.wrg", "// name = ~\"hello\"\n// input name\n")

	h := s.hover(lsproto.TextDocumentPositionParams{
		TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///vs.wrg"},
		Position:     lsproto.Position{Line: 1, Character: 9},
	})
	if !strings.Contains(h.Contents.Value, "string") {
		t.Fatalf("hover = %q, want 'string'", h.Contents.Value)
	}
}
