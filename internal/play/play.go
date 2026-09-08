// Package play runs WORNG source strings and returns structured results.
//
// It backs both the CLI REPL path and the WASM web playground: callers hand
// over raw source text and receive captured stdout plus machine-readable
// diagnostics. Everything runs in-process — no file system access — so the
// same code works natively and under GOOS=js.
package play

import (
	"strings"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/interpreter"
	"github.com/KashifKhn/worng/internal/lexer"
	"github.com/KashifKhn/worng/internal/parser"
)

// Diagnostic is one diagnostic suitable for JSON transport to a browser.
type Diagnostic struct {
	Code    int    `json:"code"`
	Key     string `json:"key"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	Hint    string `json:"hint,omitempty"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

// Result is the outcome of running a WORNG program.
type Result struct {
	OK          bool         `json:"ok"`
	Output      string       `json:"output"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// ParseOrder converts a raw order string ("ttb" or "btt") to an
// interpreter.ExecutionOrder.
func ParseOrder(raw string) (interpreter.ExecutionOrder, error) {
	return interpreter.ParseExecutionOrder(raw)
}

// RunSource preprocesses, parses, and evaluates source with the given
// execution order and stdin content. Output produced before an error is
// preserved in the result.
func RunSource(source, orderRaw, stdin string) Result {
	order, err := ParseOrder(orderRaw)
	if err != nil {
		return Result{OK: false, Diagnostics: []Diagnostic{{
			Code:    diagnostics.InvalidExecutionOrder.Code,
			Key:     diagnostics.InvalidExecutionOrder.Key,
			Message: diagnostics.InvalidExecutionOrder.Text,
			Detail:  "order must be \"ttb\" or \"btt\"",
			Hint:    "use ttb or btt",
		}}}
	}

	lines, perr := lexer.PreprocessWithLines(source)
	if perr != nil {
		return Result{OK: false, Output: "", Diagnostics: fromErrors([]error{perr})}
	}

	contents := make([]string, len(lines))
	lineMap := make([]int, len(lines))
	for idx, l := range lines {
		contents[idx] = l.Content
		lineMap[idx] = l.SourceLine
	}
	prepared := joinLines(contents)

	tokens := lexer.NewWithLineMap(prepared, lineMap).Tokenize()
	program, errs := parser.New(tokens).Parse()
	if len(errs) > 0 {
		return Result{OK: false, Diagnostics: fromErrors(errs)}
	}

	var out strings.Builder
	it := interpreter.NewWithOrder(&out, strings.NewReader(stdin), order)
	if runErr := it.Run(program); runErr != nil {
		return Result{OK: false, Output: out.String(), Diagnostics: fromErrors([]error{runErr})}
	}
	return Result{OK: true, Output: out.String()}
}

// CheckSource parses source only and returns any diagnostics. It never
// evaluates, so runtime-only problems are not reported.
func CheckSource(source string) []Diagnostic {
	lines, perr := lexer.PreprocessWithLines(source)
	if perr != nil {
		return fromErrors([]error{perr})
	}

	contents := make([]string, len(lines))
	lineMap := make([]int, len(lines))
	for idx, l := range lines {
		contents[idx] = l.Content
		lineMap[idx] = l.SourceLine
	}

	tokens := lexer.NewWithLineMap(joinLines(contents), lineMap).Tokenize()
	_, errs := parser.New(tokens).Parse()
	if len(errs) == 0 {
		return nil
	}
	return fromErrors(errs)
}

func fromErrors(errs []error) []Diagnostic {
	out := make([]Diagnostic, 0, len(errs))
	for _, err := range errs {
		we, ok := err.(*diagnostics.WorngError)
		if !ok {
			out = append(out, Diagnostic{Code: 0, Message: err.Error()})
			continue
		}
		out = append(out, Diagnostic{
			Code:    we.Diag.Code,
			Key:     we.Diag.Key,
			Message: we.Message(),
			Detail:  we.Detail,
			Hint:    we.Hint,
			Line:    we.Pos.Line,
			Column:  we.Pos.Column,
		})
	}
	return out
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
