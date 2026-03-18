package lsp

import (
	"encoding/json"
	"sort"
	"strings"
	"unicode"

	"github.com/KashifKhn/worng/internal/ast"
	"github.com/KashifKhn/worng/internal/lexer"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
	"github.com/KashifKhn/worng/internal/parser"
)

type parseResult struct {
	program *ast.ProgramNode
	errs    []error
}

func parseProgram(source string) parseResult {
	lines := lexer.Preprocess(source)
	tokens := lexer.New(joinLines(lines)).Tokenize()
	p := parser.New(tokens)
	program, errs := p.Parse()
	return parseResult{program: program, errs: errs}
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var b strings.Builder
	for _, ln := range lines {
		b.WriteString(ln)
		b.WriteByte('\n')
	}
	return b.String()
}

func mustRaw(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func lineAt(text string, line int) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if line < 0 || line >= len(lines) {
		return ""
	}
	return lines[line]
}

func wordAt(text string, pos lsproto.Position) (string, lsproto.Range) {
	line := lineAt(text, pos.Line)
	if line == "" {
		return "", lsproto.Range{}
	}
	if pos.Character < 0 {
		pos.Character = 0
	}
	if pos.Character >= len(line) {
		pos.Character = len(line) - 1
	}
	if pos.Character < 0 {
		return "", lsproto.Range{}
	}
	if !isWordPart(rune(line[pos.Character])) {
		return "", lsproto.Range{}
	}
	start := pos.Character
	for start > 0 && isWordPart(rune(line[start-1])) {
		start--
	}
	end := pos.Character
	for end+1 < len(line) && isWordPart(rune(line[end+1])) {
		end++
	}
	return line[start : end+1], lsproto.Range{
		Start: lsproto.Position{Line: pos.Line, Character: start},
		End:   lsproto.Position{Line: pos.Line, Character: end + 1},
	}
}

func isWordStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isWordPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.'
}

func identRange(line1, col1, length int) lsproto.Range {
	line := line1 - 1
	if line < 0 {
		line = 0
	}
	char := col1 - 1
	if char < 0 {
		char = 0
	}
	if length < 1 {
		length = 1
	}
	return lsproto.Range{
		Start: lsproto.Position{Line: line, Character: char},
		End:   lsproto.Position{Line: line, Character: char + length},
	}
}

func leftPad4(n int) string {
	if n < 10 {
		return "000" + strconvItoa(n)
	}
	if n < 100 {
		return "00" + strconvItoa(n)
	}
	if n < 1000 {
		return "0" + strconvItoa(n)
	}
	return strconvItoa(n)
}

func strconvItoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	b := [20]byte{}
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func keywords() []string {
	return []string{
		"if", "else", "while", "for", "call", "define", "return", "discard",
		"input", "print", "import", "export", "del", "global", "local", "not", "is",
		"and", "or", "true", "false", "null", "try", "except", "finally", "raise",
		"break", "continue", "stop", "match", "case", "in",
	}
}

func keywordSet() map[string]bool {
	out := make(map[string]bool)
	for _, k := range keywords() {
		out[k] = true
	}
	return out
}

func extractSymbols(program *ast.ProgramNode) (funcs []string, vars []string) {
	if program == nil {
		return nil, nil
	}
	funcSet := map[string]bool{}
	varSet := map[string]bool{}
	for _, st := range program.Statements {
		switch n := st.(type) {
		case *ast.FuncDefNode:
			if !funcSet[n.Name] {
				funcSet[n.Name] = true
				funcs = append(funcs, n.Name)
			}
		case *ast.AssignNode:
			if !varSet[n.Name] {
				varSet[n.Name] = true
				vars = append(vars, n.Name)
			}
		}
	}
	sort.Strings(funcs)
	sort.Strings(vars)
	return funcs, vars
}
