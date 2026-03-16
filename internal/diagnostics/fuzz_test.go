package diagnostics

import "testing"

func FuzzWorngErrorFormatting(f *testing.F) {
	f.Add("x", "file.wrg", 1, 1)
	f.Add("", "", 0, 0)

	f.Fuzz(func(t *testing.T, arg, file string, line, col int) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("diagnostics formatting panicked: %v", r)
			}
		}()

		err := New(UndefinedVariable, Position{File: file, Line: line, Column: col}, arg)
		msg := err.Error()
		if msg == "" {
			t.Fatal("formatted message should not be empty")
		}
	})
}
