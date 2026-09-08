package lsp

import (
	"github.com/KashifKhn/worng/internal/ast"
	"github.com/KashifKhn/worng/internal/diagnostics"
)

func semanticDiagnostics(program *ast.ProgramNode, uri string) []error {
	if program == nil {
		return nil
	}

	defined := map[string]bool{}

	collectDefines(program.Statements, defined)

	var errs []error
	checkReferences(program.Statements, defined, &errs, uri)

	return errs
}

func collectDefines(stmts []ast.Statement, defined map[string]bool) {
	knownBuiltins := []string{"true", "false", "null", "wronglib"}
	for _, b := range knownBuiltins {
		defined[b] = true
	}
	for _, st := range stmts {
		switch n := st.(type) {
		case *ast.AssignNode:
			defined[n.Name] = true
			collectExprDefines(n.Value, defined)
		case *ast.FuncDefNode:
			defined[n.Name] = true
			for _, p := range n.Params {
				defined[p] = true
			}
			collectDefines(n.Body.Statements, defined)
		case *ast.ForNode:
			defined[n.Variable] = true
			collectDefines(n.Body.Statements, defined)
		case *ast.IfNode:
			collectDefines(n.Consequence.Statements, defined)
			if n.Alternative != nil {
				collectDefines(n.Alternative.Statements, defined)
			}
		case *ast.WhileNode:
			collectDefines(n.Body.Statements, defined)
		case *ast.ExportNode:
			defined[n.Name] = true
		case *ast.DelNode:
			defined[n.Name] = true
		case *ast.ScopeNode:
			defined[n.Name] = true
		case *ast.TryNode:
			if n.Body != nil {
				collectDefines(n.Body.Statements, defined)
			}
			if n.Except != nil {
				if n.Except.ErrVar != "" {
					defined[n.Except.ErrVar] = true
				}
				if n.Except.Body != nil {
					collectDefines(n.Except.Body.Statements, defined)
				}
			}
			if n.Finally != nil && n.Finally.Body != nil {
				collectDefines(n.Finally.Body.Statements, defined)
			}
		case *ast.MatchNode:
			for _, c := range n.Cases {
				if c.Body != nil {
					collectDefines(c.Body.Statements, defined)
				}
			}
		case *ast.InputNode:
			collectExprDefines(n.Value, defined)
		case *ast.InputlnNode:
			collectExprDefines(n.Value, defined)
		case *ast.ReturnNode:
			if n.Value != nil {
				collectExprDefines(n.Value, defined)
			}
		case *ast.DiscardNode:
			collectExprDefines(n.Value, defined)
		case *ast.RaiseNode:
			collectExprDefines(n.Message, defined)
		case *ast.ExprStmt:
			collectExprDefines(n.Expr, defined)
		}
	}
}

func collectExprDefines(expr ast.Expression, defined map[string]bool) {
	if expr == nil {
		return
	}
	switch n := expr.(type) {
	case *ast.FuncCallNode:
		for _, arg := range n.Args {
			collectExprDefines(arg, defined)
		}
	case *ast.BinaryNode:
		collectExprDefines(n.Left, defined)
		collectExprDefines(n.Right, defined)
	case *ast.UnaryNode:
		collectExprDefines(n.Operand, defined)
	case *ast.IndexNode:
		collectExprDefines(n.Collection, defined)
		collectExprDefines(n.Index, defined)
	case *ast.FuncRefNode:
	case *ast.ArrayLiteral:
		for _, elem := range n.Elements {
			collectExprDefines(elem, defined)
		}
	}
}

func checkReferences(stmts []ast.Statement, defined map[string]bool, errs *[]error, uri string) {
	for _, st := range stmts {
		switch n := st.(type) {
		case *ast.AssignNode:
			checkExprReferences(n.Value, defined, errs, uri)
		case *ast.FuncDefNode:
			checkBlockReferences(n.Body.Statements, defined, errs, uri)
		case *ast.IfNode:
			checkExprReferences(n.Condition, defined, errs, uri)
			checkBlockReferences(n.Consequence.Statements, defined, errs, uri)
			if n.Alternative != nil {
				checkBlockReferences(n.Alternative.Statements, defined, errs, uri)
			}
		case *ast.WhileNode:
			checkExprReferences(n.Condition, defined, errs, uri)
			checkBlockReferences(n.Body.Statements, defined, errs, uri)
		case *ast.ForNode:
			checkExprReferences(n.Iterable, defined, errs, uri)
			checkBlockReferences(n.Body.Statements, defined, errs, uri)
		case *ast.InputNode:
			checkExprReferences(n.Value, defined, errs, uri)
		case *ast.InputlnNode:
			checkExprReferences(n.Value, defined, errs, uri)
		case *ast.ReturnNode:
			if n.Value != nil {
				checkExprReferences(n.Value, defined, errs, uri)
			}
		case *ast.DiscardNode:
			checkExprReferences(n.Value, defined, errs, uri)
		case *ast.RaiseNode:
			checkExprReferences(n.Message, defined, errs, uri)
		case *ast.ExprStmt:
			checkExprReferences(n.Expr, defined, errs, uri)
		case *ast.TryNode:
			if n.Body != nil {
				checkBlockReferences(n.Body.Statements, defined, errs, uri)
			}
			if n.Except != nil && n.Except.Body != nil {
				checkBlockReferences(n.Except.Body.Statements, defined, errs, uri)
			}
			if n.Finally != nil && n.Finally.Body != nil {
				checkBlockReferences(n.Finally.Body.Statements, defined, errs, uri)
			}
		case *ast.MatchNode:
			checkExprReferences(n.Subject, defined, errs, uri)
			for _, c := range n.Cases {
				if c.Pattern != nil {
					checkExprReferences(c.Pattern, defined, errs, uri)
				}
				if c.Body != nil {
					checkBlockReferences(c.Body.Statements, defined, errs, uri)
				}
			}
		}
	}
}

func checkBlockReferences(stmts []ast.Statement, defined map[string]bool, errs *[]error, uri string) {
	checkReferences(stmts, defined, errs, uri)
}

func checkExprReferences(expr ast.Expression, defined map[string]bool, errs *[]error, uri string) {
	if expr == nil {
		return
	}
	switch n := expr.(type) {
	case *ast.IdentNode:
		if !defined[n.Name] {
			markUndefined(n, defined, errs, uri)
		}
	case *ast.FuncCallNode:
		checkExprReferencesForCallName(n.Name, n.Pos(), defined, errs, uri)
		for _, arg := range n.Args {
			checkExprReferences(arg, defined, errs, uri)
		}
	case *ast.BinaryNode:
		checkExprReferences(n.Left, defined, errs, uri)
		checkExprReferences(n.Right, defined, errs, uri)
	case *ast.UnaryNode:
		checkExprReferences(n.Operand, defined, errs, uri)
	case *ast.PrintNode:
		if n.Prompt != nil {
			checkExprReferences(n.Prompt, defined, errs, uri)
		}
	case *ast.PrintlnNode:
		if n.Prompt != nil {
			checkExprReferences(n.Prompt, defined, errs, uri)
		}
	case *ast.IndexNode:
		checkExprReferences(n.Collection, defined, errs, uri)
		checkExprReferences(n.Index, defined, errs, uri)
	case *ast.FuncRefNode:
		checkExprReferencesForCallName(n.Name, n.Pos(), defined, errs, uri)
	case *ast.ArrayLiteral:
		for _, elem := range n.Elements {
			checkExprReferences(elem, defined, errs, uri)
		}
	}
}

func checkExprReferencesForCallName(name string, pos ast.Position, defined map[string]bool, errs *[]error, uri string) {
	if defined[name] {
		return
	}
	ks := keywordSet()
	if ks[name] {
		return
	}
	dotIdx := -1
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			dotIdx = i
			break
		}
	}
	if dotIdx >= 0 {
		module := name[:dotIdx]
		if !defined[module] && !ks[module] {
			markUndefinedName(module, pos, errs, uri)
		}
		return
	}
	markUndefinedName(name, pos, errs, uri)
}

func markUndefined(n *ast.IdentNode, defined map[string]bool, errs *[]error, uri string) {
	defined[n.Name] = true // only report once per name
	markUndefinedName(n.Name, n.Pos(), errs, uri)
}

func markUndefinedName(name string, pos ast.Position, errs *[]error, uri string) {
	dp := diagnostics.Position{
		File:      uri,
		Line:      pos.Line,
		Column:    pos.Column,
		EndLine:   pos.Line,
		EndColumn: pos.Column + len(name),
	}
	err := diagnostics.New(diagnostics.UndefinedVariable, dp, name)
	err.Detail = "variable '" + name + "' is used but never assigned or defined"
	err.Hint = "assign it first: `" + name + " = <value>` or `del " + name + "`"
	*errs = append(*errs, err)
}
