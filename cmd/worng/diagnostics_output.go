package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/vfs"
)

type jsonDiagnostic struct {
	Code     int      `json:"code"`
	Key      string   `json:"key"`
	Severity string   `json:"severity"`
	File     string   `json:"file,omitempty"`
	Line     int      `json:"line,omitempty"`
	Column   int      `json:"column,omitempty"`
	EndLine  int      `json:"endLine,omitempty"`
	EndCol   int      `json:"endColumn,omitempty"`
	Message  string   `json:"message"`
	Detail   string   `json:"detail,omitempty"`
	Hint     string   `json:"hint,omitempty"`
	Expected []string `json:"expected,omitempty"`
	Found    string   `json:"found,omitempty"`
}

func printDiagnostics(w io.Writer, err error, fs vfs.FS, sourcePath string, jsonOutput bool) {
	if jsonOutput {
		printDiagnosticsJSON(w, err)
		return
	}
	printDiagnosticsPretty(w, err, fs, sourcePath)
}

func printDiagnosticsPretty(w io.Writer, err error, fs vfs.FS, sourcePath string) {
	items := flattenErrors(err)
	if len(items) == 0 {
		_, _ = fmt.Fprintln(w, err)
		return
	}

	lines := map[string][]string{}
	if sourcePath != "" {
		if data, readErr := fs.ReadFile(sourcePath); readErr == nil {
			lines[sourcePath] = splitLines(string(data))
		}
	}

	for idx, item := range items {
		if idx > 0 {
			_, _ = fmt.Fprintln(w)
		}
		if we, ok := item.(*diagnostics.WorngError); ok {
			printWorngErrorPretty(w, we, lines)
			continue
		}
		_, _ = fmt.Fprintln(w, item.Error())
	}
}

func printWorngErrorPretty(w io.Writer, we *diagnostics.WorngError, fileLines map[string][]string) {
	if we.Pos.File != "" {
		_, _ = fmt.Fprintf(w, "%s:%d:%d: [W%04d] %s\n", we.Pos.File, we.Pos.Line, we.Pos.Column, we.Diag.Code, we.Message())
	} else {
		_, _ = fmt.Fprintf(w, "[W%04d] %s\n", we.Diag.Code, we.Message())
	}

	if strings.TrimSpace(we.Detail) != "" {
		_, _ = fmt.Fprintf(w, "detail: %s\n", we.Detail)
	}
	if strings.TrimSpace(we.Hint) != "" {
		_, _ = fmt.Fprintf(w, "hint: %s\n", we.Hint)
	}

	if we.Pos.File == "" || we.Pos.Line <= 0 {
		return
	}
	source := fileLines[we.Pos.File]
	if len(source) == 0 || we.Pos.Line > len(source) {
		return
	}
	lineText := source[we.Pos.Line-1]
	_, _ = fmt.Fprintf(w, " %d | %s\n", we.Pos.Line, lineText)
	caretCol := we.Pos.Column
	if caretCol <= 0 {
		caretCol = 1
	}
	endCol := we.Pos.EndColumn
	if endCol < caretCol {
		endCol = caretCol
	}
	pad := strings.Repeat(" ", len(fmt.Sprintf(" %d | ", we.Pos.Line))+caretCol-1)
	marks := strings.Repeat("^", endCol-caretCol+1)
	_, _ = fmt.Fprintf(w, "%s%s\n", pad, marks)
}

func printDiagnosticsJSON(w io.Writer, err error) {
	items := flattenErrors(err)
	out := make([]jsonDiagnostic, 0, len(items))
	for _, item := range items {
		we, ok := item.(*diagnostics.WorngError)
		if !ok {
			out = append(out, jsonDiagnostic{Message: item.Error(), Severity: "error"})
			continue
		}
		out = append(out, jsonDiagnostic{
			Code:     we.Diag.Code,
			Key:      we.Diag.Key,
			Severity: "error",
			File:     we.Pos.File,
			Line:     we.Pos.Line,
			Column:   we.Pos.Column,
			EndLine:  we.Pos.EndLine,
			EndCol:   we.Pos.EndColumn,
			Message:  we.Message(),
			Detail:   we.Detail,
			Hint:     we.Hint,
			Expected: we.Expected,
			Found:    we.Found,
		})
	}
	b, marshalErr := json.MarshalIndent(out, "", "  ")
	if marshalErr != nil {
		_, _ = fmt.Fprintln(w, err)
		return
	}
	_, _ = fmt.Fprintln(w, string(b))
}

func flattenErrors(err error) []error {
	if err == nil {
		return nil
	}
	if list, ok := err.(*diagnostics.ErrorList); ok {
		return list.Items
	}
	return []error{err}
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	norm := strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(norm, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
