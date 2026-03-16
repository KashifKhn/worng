package golden

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/interpreter"
	"github.com/KashifKhn/worng/internal/lexer"
	"github.com/KashifKhn/worng/internal/parser"
	"github.com/KashifKhn/worng/internal/vfs"
)

func TestGolden(t *testing.T) {
	t.Parallel()

	fixtureRoot := filepath.Join("..", "..", "testdata")
	inputs, err := filepath.Glob(filepath.Join(fixtureRoot, "*", "input.wrg"))
	if err != nil {
		t.Fatalf("glob input fixtures: %v", err)
	}
	if len(inputs) == 0 {
		t.Fatal("no golden fixtures found under testdata/*/input.wrg")
	}
	sort.Strings(inputs)

	for _, inputPath := range inputs {
		caseDir := filepath.Dir(inputPath)
		caseName := filepath.Base(caseDir)

		t.Run(caseName, func(t *testing.T) {
			t.Parallel()

			inputBytes, err := osReadFile(inputPath)
			if err != nil {
				t.Fatalf("read %s: %v", inputPath, err)
			}

			expectedPath := filepath.Join(caseDir, "expected.txt")
			expectedBytes, err := osReadFile(expectedPath)
			if err != nil {
				t.Fatalf("read %s: %v", expectedPath, err)
			}

			stdinText := readOptionalFile(t, filepath.Join(caseDir, "stdin.txt"))
			expectedErr := readOptionalFile(t, filepath.Join(caseDir, "expected_error.txt"))

			order := interpreter.OrderBottomToTop
			if orderText := strings.TrimSpace(readOptionalFile(t, filepath.Join(caseDir, "order.txt"))); orderText != "" {
				parsed, err := interpreter.ParseExecutionOrder(orderText)
				if err != nil {
					t.Fatalf("invalid order in %s/order.txt: %v", caseDir, err)
				}
				order = parsed
			}

			mem := vfs.NewMemFS()
			if err := mem.WriteFile("input.wrg", inputBytes); err != nil {
				t.Fatalf("memfs write input: %v", err)
			}

			actual, err := runGoldenCase(mem, "input.wrg", order, stdinText)
			expectErr := strings.TrimSpace(expectedErr)
			if expectErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", expectErr)
				}
				if !strings.Contains(err.Error(), expectErr) {
					t.Fatalf("error mismatch\nactual: %q\nwant substring: %q", err.Error(), expectErr)
				}
			} else if err != nil {
				t.Fatalf("run case %s: %v", caseName, err)
			}

			expected := string(expectedBytes)
			if actual != expected {
				_ = mem.WriteFile("actual.txt", []byte(actual))
				t.Fatalf("golden mismatch (%s)\nactual: %q\nexpected: %q", caseName, actual, expected)
			}
		})
	}
}

func runGoldenCase(mem vfs.FS, inputPath string, order interpreter.ExecutionOrder, stdin string) (string, error) {
	data, err := mem.ReadFile(inputPath)
	if err != nil {
		return "", err
	}

	prepared := joinExecutableLines(lexer.Preprocess(string(data)))
	tokens := lexer.New(prepared).Tokenize()
	p := parser.New(tokens)
	program, errs := p.Parse()
	if len(errs) > 0 {
		return "", errs[0]
	}

	var out bytes.Buffer
	it := interpreter.NewWithOrder(&out, strings.NewReader(stdin), order)
	if err := it.Run(program); err != nil {
		return out.String(), err
	}
	return out.String(), nil
}

func readOptionalFile(t *testing.T, filePath string) string {
	t.Helper()
	data, err := osReadFile(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ""
		}
		t.Fatalf("read %s: %v", filePath, err)
	}
	return string(data)
}

func joinExecutableLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func osReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
