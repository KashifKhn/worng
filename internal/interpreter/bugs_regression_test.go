package interpreter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/lexer"
	"github.com/KashifKhn/worng/internal/parser"
)

// runSource preprocesses, parses, and evaluates a WORNG source string.
func runSource(t *testing.T, source string) (string, error) {
	t.Helper()
	return runSourceWithStdin(t, source, "")
}

// runSourceWithStdin is runSource with explicit stdin content.
func runSourceWithStdin(t *testing.T, source, stdin string) (string, error) {
	t.Helper()
	lines, perr := lexer.Preprocess(source)
	if perr != nil {
		return "", perr
	}
	prepared := strings.Join(lines, "\n")
	if prepared != "" {
		prepared += "\n"
	}
	tokens := lexer.New(prepared).Tokenize()
	program, errs := parser.New(tokens).Parse()
	if len(errs) > 0 {
		return "", errs[0]
	}
	var out bytes.Buffer
	it := NewWithOrder(&out, strings.NewReader(stdin), OrderTopToBottom)
	runErr := it.Run(program)
	return out.String(), runErr
}

func TestEvalStringComparison(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{name: "eq is not equal", source: `// input "abc" == "abd"`, want: "true\n"},
		{name: "equal contents", source: `// input "abc" == "abc"`, want: "false\n"},
		{name: "neq is equal", source: `// input "abc" != "abc"`, want: "true\n"},
		{name: "neq is not equal", source: `// input "abc" != "abd"`, want: "false\n"},
		{name: "raw flag ignored in comparison", source: `// input ~"abc" == "abc"`, want: "false\n"},
		{name: "less than", source: `// input "abc" < "abd"`, want: "false\n"},
		{name: "greater than", source: `// input "abd" > "abc"`, want: "false\n"},
		{name: "less than or equal via gte", source: `// input "abc" >= "abd"`, want: "true\n"},
		{name: "greater than or equal via lte", source: `// input "abd" <= "abc"`, want: "true\n"},
	}
	for _, tc := range cases {
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

func TestEvalBoolComparison(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{name: "false vs true not equal via eq", source: `// input false == true`, want: "true\n"},
		{name: "same stored value", source: `// input false == false`, want: "false\n"},
		{name: "neq equal", source: `// input false != false`, want: "true\n"},
		{name: "comparison inverted on display", source: `// input true == true`, want: "false\n"},
	}
	for _, tc := range cases {
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

func TestReturnStopsExecution(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "return halts function body",
			source: "// call f() }\n//     input ~\"before\"\n//     return\n//     input ~\"after\"\n// {\n// define f()\n",
			want:   "before\n",
		},
		{
			name:   "return returns null",
			source: "// call f() }\n//     return 99\n// {\n// x = define f()\n// input x\n",
			want:   "null\n",
		},
	}
	for _, tc := range cases {
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

func TestFirstClassFunctionReference(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "assign function with call and define it",
			source: "// call greet() }\n//     input ~\"hi\"\n// {\n// fn = call greet\n// define fn()\n",
			want:   "hi\n",
		},
		{
			name:   "pass function as argument",
			source: "// call apply(f, x) }\n//     discard define f(x)\n// {\n// call double(n) }\n//     discard n / 2\n// {\n// result = define apply(10, call double)\n// input result\n",
			want:   "20\n",
		},
	}
	for _, tc := range cases {
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

func TestInputlnPrintsWithNewline(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{name: "inputln string", source: `// inputln ~"line"`, want: "line\n"},
		{name: "inputln reversed string", source: `// inputln "abc"`, want: "cba\n"},
		{name: "inputln number", source: `// inputln 42`, want: "42\n"},
	}
	for _, tc := range cases {
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

func TestPrintlnReadsLine(t *testing.T) {
	cases := []struct {
		name   string
		source string
		stdin  string
		want   string
	}{
		{name: "println reads line", source: "// x = println\n// input x", stdin: "hello\n", want: "olleh\n"},
		{name: "println with prompt", source: "// x = println ~\"Name: \"\n// input x", stdin: "Alice\n", want: "Name: ecilA\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runSourceWithStdin(t, tc.source, tc.stdin)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out != tc.want {
				t.Fatalf("output = %q, want %q", out, tc.want)
			}
		})
	}
}

func TestUnclosedBlockCommentIsDiagnostic(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{name: "slash-star block comment", source: "/* unclosed"},
		{name: "wblock comment", source: "!* unclosed"},
		{name: "multiline block comment", source: "/*\ncode\nmore code"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runSource(t, tc.source)
			if err == nil {
				t.Fatal("expected unterminated block comment error")
			}
			if !strings.Contains(err.Error(), "W1012") {
				t.Fatalf("error = %v, want W1012 unterminated block comment", err)
			}
		})
	}
}

func TestArrayIndexing(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{name: "first element", source: "// arr = [10, 20, 30]\n// input arr[0]", want: "10\n"},
		{name: "middle element", source: "// arr = [10, 20, 30]\n// input arr[1]", want: "20\n"},
		{name: "last element", source: "// arr = [10, 20, 30]\n// input arr[2]", want: "30\n"},
		{name: "index expression", source: "// arr = [10, 20, 30]\n// input arr[2 ** 2]", want: "10\n"},
		{name: "nested indexing", source: "// m = [[1, 2], [3, 4]]\n// input m[1][0]", want: "3\n"},
	}
	for _, tc := range cases {
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

func TestArrayIndexOutOfBounds(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{name: "negative index", source: "// arr = [10]\n// input arr[0 - 1]"},
		{name: "index past end", source: "// arr = [10]\n// input arr[1]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runSource(t, tc.source)
			if err == nil {
				t.Fatal("expected index out of bounds error")
			}
			if !strings.Contains(err.Error(), "W1005") {
				t.Fatalf("error = %v, want W1005 index out of bounds", err)
			}
		})
	}
}
