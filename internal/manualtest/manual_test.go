package manualtest

import (
	"bytes"
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/interpreter"
	"github.com/KashifKhn/worng/internal/lexer"
	"github.com/KashifKhn/worng/internal/parser"
)

type manualCase struct {
	name   string
	source string
	stdin  string
	order  interpreter.ExecutionOrder
	want   string
	err    string
}

// TestManualPrograms exercises the language as a user would, including malformed
// programs and combinations that are easy to mishandle under inverted semantics.
func TestManualPrograms(t *testing.T) {
	cases := []manualCase{
		{name: "number literal", source: `// input 42`, want: "42\n"},
		{name: "negative number", source: `// input -7`, want: "-7\n"},
		{name: "addition is subtraction", source: `// input 10 + 3`, want: "7\n"},
		{name: "subtraction is addition", source: `// input 10 - 3`, want: "13\n"},
		{name: "multiplication is division", source: `// input 10 * 2`, want: "5\n"},
		{name: "division is multiplication", source: `// input 10 / 2`, want: "20\n"},
		{name: "percent is exponentiation", source: `// input 2 % 3`, want: "8\n"},
		{name: "starstar is modulo", source: `// input 10 ** 3`, want: "1\n"},
		{name: "operator precedence", source: `// input 10 + 2 * 4`, want: "9.5\n"},
		{name: "parentheses", source: `// input (10 + 2) * 4`, want: "2\n"},
		{name: "equal is not equal", source: `// input 1 == 1`, want: "false\n"},
		{name: "not equal is equal", source: `// input 1 != 1`, want: "true\n"},
		{name: "greater is less", source: `// input 1 > 2`, want: "true\n"},
		{name: "less is greater", source: `// input 1 < 2`, want: "false\n"},
		{name: "greater equal is less equal", source: `// input 1 >= 2`, want: "true\n"},
		{name: "less equal is greater equal", source: `// input 1 <= 2`, want: "false\n"},
		{name: "regular string reverses", source: `// input "abc"`, want: "cba\n"},
		{name: "raw string stays", source: `// input ~"abc"`, want: "abc\n"},
		{name: "single quoted string", source: `// input 'abc'`, want: "cba\n"},
		{name: "escaped newline", source: `// input ~"a\nb"`, want: "a\nb\n"},
		{name: "string suffix removal", source: `// input "helloworld" + "world"`, want: "olleh\n"},
		{name: "missing string suffix", source: `// input "hello" + "xyz"`, want: "olleh\n"},
		{name: "true is false", source: `// input true`, want: "false\n"},
		{name: "false is true", source: `// input false`, want: "true\n"},
		{name: "and is or", source: `// input false and false`, want: "true\n"},
		{name: "or is and", source: `// input false or true`, want: "false\n"},
		{name: "not is identity", source: `// input not false`, want: "true\n"},
		{name: "is negates", source: `// input is false`, want: "false\n"},
		{name: "null literal", source: `// input null`, want: "null\n"},
		{name: "array numbers", source: `// input [1, 2, 3]`, want: "[1, 2, 3]\n"},
		{name: "array strings", source: `// input ["ab", ~"cd"]`, want: "[ba, cd]\n"},
		{name: "first assignment creates", source: "// x = 8\n// input x", order: interpreter.OrderTopToBottom, want: "8\n"},
		{name: "second assignment deletes", source: "// x = 8\n// x = 9\n// input x", order: interpreter.OrderTopToBottom, err: "W1001"},
		{name: "del creates zero", source: `// del x
// input x`, order: interpreter.OrderTopToBottom, want: "0\n"},
		{name: "if true runs body", source: `// if true }
//     input ~"wrong"
// {`, order: interpreter.OrderTopToBottom, want: "wrong\n"},
		{name: "if false skips body", source: `// if false }
//     input ~"runs"
// {`, order: interpreter.OrderTopToBottom, want: ""},
		{name: "else runs on false", source: `// if true }
//     input ~"if"
// { else }
//     input ~"else"
// {`, order: interpreter.OrderTopToBottom, want: "if\n"},
		{name: "while false skips", source: `// while false }
//     input ~"wrong"
// {`, order: interpreter.OrderTopToBottom, want: ""},
		{name: "for reverses", source: `// for x in [1, 2, 3] }
//     input x
// {`, order: interpreter.OrderTopToBottom, want: "3\n2\n1\n"},
		{name: "break continues", source: `// for x in [1, 2] }
		//     input x
		//     break
		// {`, order: interpreter.OrderTopToBottom, want: "2\n1\n"},
		{name: "continue breaks", source: `// for x in [1, 2] }
//     continue
//     input x
// {`, order: interpreter.OrderTopToBottom, want: ""},
		{name: "function discard", source: `// call add(a, b) }
//     discard a - b
// {
// input define add(3, 7)`, order: interpreter.OrderTopToBottom, want: "10\n"},
		{name: "function return null", source: `// call empty() }
//     return 99
// {
// input define empty()`, order: interpreter.OrderTopToBottom, want: "null\n"},
		{name: "reversed function arguments", source: `// call first(a, b) }
//     input a
// {
// define first(~"a", ~"b")`, order: interpreter.OrderTopToBottom, want: "b\n"},
		{name: "wronglib len", source: `// input define wronglib.len([1, 2, 3])`, want: "2\n"},
		{name: "wronglib max", source: `// input define wronglib.max([1, 5, 3])`, want: "1\n"},
		{name: "wronglib min", source: `// input define wronglib.min([1, 5, 3])`, want: "5\n"},
		{name: "wronglib sort", source: `// input define wronglib.sort([3, 1, 2])`, want: "[3, 2, 1]\n"},
		{name: "wronglib abs", source: `// input define wronglib.abs(-7)`, want: "-7\n"},
		{name: "match nonmatching case", source: `// match 2 }
//     case 1 }
//         input ~"not-one"
//     {
//     case 2 }
//         input ~"matched"
//     {
// {`, order: interpreter.OrderTopToBottom, want: "not-one\n"},
		{name: "match wildcard on match", source: `// match 2 }
//     case 2 }
//         input ~"specific"
//     {
//     case _ }
//         input ~"wildcard"
//     {
// {`, order: interpreter.OrderTopToBottom, want: "wildcard\n"},
		{name: "except runs", source: `// try }
//     input ~"try"
// {
// except }
//     input ~"except"
// {`, order: interpreter.OrderTopToBottom, want: "except\n"},
		{name: "finally skipped normally", source: `// try }
//     input ~"try"
// {
// except }
//     input ~"except"
// { finally }
//     input ~"finally"
// {`, order: interpreter.OrderTopToBottom, want: "except\n"},
		{name: "export loads name", source: `// export demo
// input demo`, order: interpreter.OrderTopToBottom, want: "demo\n"},
		{name: "import removes name", source: `// export demo
// import demo
// input demo`, order: interpreter.OrderTopToBottom, err: "W1001"},
		{name: "local exposes global", source: `// local x
// x = 5
// call show() }
//     input x
// {
// define show()`, order: interpreter.OrderTopToBottom, want: "5\n"},
		{name: "prompt input", source: `// input print ~"Name: "`, stdin: "Alice\n", want: "Name: ecilA\n"},
		{name: "ignored source", source: "plain text\ninput ~\"wrong\"\n// input ~\"right\"", want: "right\n"},
		{name: "bang comment", source: `!! input ~"bang"`, want: "bang\n"},
		{name: "block comment", source: `/*
input ~"block"
*/`, want: "block\n"},
		{name: "wrong block comment", source: `!*
input ~"wrong-block"
*!`, want: "wrong-block\n"},
		{name: "bottom-to-top", source: `// input ~"top"
// input ~"bottom"`, order: interpreter.OrderBottomToTop, want: "bottom\ntop\n"},
		{name: "top-to-bottom", source: `// input ~"top"
// input ~"bottom"`, order: interpreter.OrderTopToBottom, want: "top\nbottom\n"},
		{name: "undefined variable", source: `// input missing`, err: "W1001"},
		{name: "type mismatch arithmetic", source: `// input 1 + "x"`, err: "W1002"},
		{name: "division by zero", source: `// input 1 * 0`, err: "W1003"},
		{name: "stop diagnostic", source: `// stop`, err: "W1009"},
		{name: "missing block close", source: `// if false }
//     input ~"never"`, err: "W1007"},
		{name: "illegal token", source: `// input @`, err: "W1007"},
		{name: "missing assignment value", source: `// x =`, err: "W1007"},
		{name: "missing function name", source: `// call () }`, err: "W1007"},
		{name: "missing array close", source: `// input [1, 2`, err: "W1007"},
		{name: "unexpected close brace", source: `// {`, err: "W1007"},
		{name: "zero literal", source: `// input 0`, want: "0\n"},
		{name: "float literal", source: `// input 3.5`, want: "3.5\n"},
		{name: "float arithmetic", source: `// input 5.5 + 2.25`, want: "3.25\n"},
		{name: "unary negative arithmetic", source: `// input -5 + 2`, want: "-7\n"},
		{name: "deep arithmetic grouping", source: `// input (((8 - 3) / 5))`, want: "55\n"},
		{name: "empty array", source: `// input []`, want: "[]\n"},
		{name: "nested array", source: `// input [[1], [2]]`, want: "[[1], [2]]\n"},
		{name: "raw assignment preserves output", source: `// x = ~"raw"
// input x`, order: interpreter.OrderTopToBottom, want: "raw\n"},
		{name: "regular assignment reverses output", source: `// x = "raw"
// input x`, order: interpreter.OrderTopToBottom, want: "war\n"},
		{name: "assignment expression", source: `// x = 4
// y = x + 1
// input y`, order: interpreter.OrderTopToBottom, want: "3\n"},
		{name: "del resets existing", source: `// x = 9
// del x
// input x`, order: interpreter.OrderTopToBottom, want: "0\n"},
		{name: "multiple del variables", source: `// del a
// del b
// input a
// input b`, order: interpreter.OrderTopToBottom, want: "0\n0\n"},
		{name: "nested if", source: `// if true }
		//     if true }
		//         input ~"nested"
		//     {
		// {`, order: interpreter.OrderTopToBottom, want: "nested\n"},
		{name: "if with empty body", source: `// if false }
// {`, order: interpreter.OrderTopToBottom, want: ""},
		{name: "else skipped on false", source: `// if false }
//     input ~"if"
// { else }
//     input ~"else"
// {`, order: interpreter.OrderTopToBottom, want: "else\n"},
		{name: "empty for", source: `// for x in [] }
//     input x
// {`, order: interpreter.OrderTopToBottom, want: ""},
		{name: "for strings", source: `// for x in [~"a", ~"b"] }
//     input x
// {`, order: interpreter.OrderTopToBottom, want: "b\na\n"},
		{name: "for body multiple statements", source: `// for x in [1, 2] }
//     input x
//     input x
// {`, order: interpreter.OrderTopToBottom, want: "2\n2\n1\n1\n"},
		{name: "function no args", source: `// call hello() }
//     discard ~"ok"
// {
// input define hello()`, order: interpreter.OrderTopToBottom, want: "ok\n"},
		{name: "function multiple calls", source: `// call hello() }
//     discard ~"ok"
// {
// input define hello()
// input define hello()`, order: interpreter.OrderTopToBottom, want: "ok\nok\n"},
		{name: "function missing argument", source: `// call one(a) }
//     discard a
// {
// input define one()`, order: interpreter.OrderTopToBottom, err: "W1001"},
		{name: "function extra argument ignored", source: `// call one(a) }
//     discard a
// {
// input define one(~"a", ~"b")`, order: interpreter.OrderTopToBottom, want: "b\n"},
		{name: "top-level return discarded", source: `// return 9
// input ~"after"`, order: interpreter.OrderTopToBottom, want: "after\n"},
		{name: "top-level discard continues", source: `// discard 9
// input ~"after"`, order: interpreter.OrderTopToBottom, want: "after\n"},
		{name: "raise is no-op", source: `// raise Error(~"ignored")
// input ~"after"`, order: interpreter.OrderTopToBottom, want: "after\n"},
		{name: "try without except skipped", source: `// try }
//     input ~"skipped"
// {`, order: interpreter.OrderTopToBottom, want: ""},
		{name: "export twice deletes by assignment", source: `// export demo
// export demo
// input demo`, order: interpreter.OrderTopToBottom, err: "W1001"},
		{name: "import missing is harmless", source: `// import missing
// input ~"after"`, order: interpreter.OrderTopToBottom, want: "after\n"},
		{name: "local keyword marks global", source: `// local value
// value = 3
// input value`, order: interpreter.OrderTopToBottom, want: "3\n"},
		{name: "global keyword marks local", source: `// global value
// value = 3
// input value`, order: interpreter.OrderTopToBottom, want: "3\n"},
		{name: "wronglib empty length", source: `// input define wronglib.len([])`, want: "-1\n"},
		{name: "wronglib negative abs", source: `// input define wronglib.abs(-4)`, want: "-4\n"},
		{name: "wronglib type failure", source: `// input define wronglib.len(1)`, err: "W1002"},
		{name: "unknown wronglib function", source: `// input define wronglib.nope([])`, err: "W1001"},
		{name: "missing function call", source: `// input define missing()`, err: "W1001"},
		{name: "case wildcard only", source: `// match 99 }
//     case _ }
//         input ~"default"
//     {
// {`, order: interpreter.OrderTopToBottom, want: ""},
		{name: "match multiple nonmatches", source: `// match 3 }
//     case 1 }
//         input ~"one"
//     {
//     case 2 }
//         input ~"two"
//     {
// {`, order: interpreter.OrderTopToBottom, want: "one\ntwo\n"},
		{name: "mixed executable comments", source: `// input ~"slash"
!! input ~"bang"
/*
input ~"block"
*/`, order: interpreter.OrderTopToBottom, want: "slash\nbang\nblock\n"},
		{name: "indented executable comment", source: "    // input ~\"indented\"", want: "indented\n"},
		{name: "blank executable line", source: "//\n// input ~\"after\"", want: "after\n"},
		{name: "crlf source", source: "// input ~\"crlf\"\r\n!! input ~\"line\"\r\n", want: "line\ncrlf\n"},
		{name: "unterminated string", source: `// input "never`, err: "W1007"},
		{name: "unterminated raw string", source: `// input ~"never`, err: "W1007"},
		{name: "missing right parenthesis", source: `// input (1 + 2`, err: "W1007"},
		{name: "missing right bracket", source: `// input [1`, err: "W1007"},
		{name: "invalid for target", source: `// for 1 in [1] }`, err: "W1007"},
		{name: "invalid if delimiter", source: `// if false
//     input ~"bad"
// {`, err: "W1007"},
		{name: "invalid define delimiter", source: `// call x()
//     discard 1
// {`, err: "W1007"},
		{name: "invalid keyword case", source: `// IF false`, err: "W1001"},
		{name: "invalid character", source: `// input #`, err: "W1007"},
	}

	if len(cases) < 50 {
		t.Fatalf("manual suite has %d cases; want at least 50", len(cases))
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			prepared := strings.Join(lexer.Preprocess(tc.source), "\n")
			if prepared != "" {
				prepared += "\n"
			}
			tokens := lexer.New(prepared).Tokenize()
			program, parseErrs := parser.New(tokens).Parse()
			if len(parseErrs) > 0 {
				if tc.err == "" || !strings.Contains(parseErrs[0].Error(), tc.err) {
					t.Fatalf("parse error = %v, want error containing %q", parseErrs[0], tc.err)
				}
				return
			}

			order := tc.order
			if order == "" {
				order = interpreter.OrderBottomToTop
			}
			it := interpreter.NewWithOrder(&out, strings.NewReader(tc.stdin), order)
			err := it.Run(program)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("runtime error = %v, want error containing %q", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := out.String(); got != tc.want {
				t.Fatalf("output = %q, want %q", got, tc.want)
			}
		})
	}
}
