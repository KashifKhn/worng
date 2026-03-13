package parser

import (
	"reflect"
	"testing"

	"github.com/KashifKhn/worng/internal/ast"
	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/lexer"
)

func TestParseAssignStmt(t *testing.T) {
	t.Parallel()

	program, errs := parseProgramFromSource(t, "x = 42\n")
	assertNoParseErrors(t, errs)

	stmt := mustAssignStmt(t, program, 0)
	if stmt.Name != "x" {
		t.Fatalf("assign name = %q, want %q", stmt.Name, "x")
	}
	n := mustNumberLiteral(t, stmt.Value)
	if n.Value != 42 {
		t.Fatalf("number value = %v, want 42", n.Value)
	}
}

func TestParseLiteralAssignmentsTableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		check  func(t *testing.T, a *ast.AssignNode)
	}{
		{
			name:   "number",
			source: "x = 3.14\n",
			check: func(t *testing.T, a *ast.AssignNode) {
				t.Helper()
				n := mustNumberLiteral(t, a.Value)
				if n.Value != 3.14 {
					t.Fatalf("number value = %v, want 3.14", n.Value)
				}
			},
		},
		{
			name:   "string",
			source: "x = \"hello\"\n",
			check: func(t *testing.T, a *ast.AssignNode) {
				t.Helper()
				s := mustStringLiteral(t, a.Value)
				if s.Value != "hello" || s.Raw {
					t.Fatalf("string = {Value:%q Raw:%v}, want {Value:hello Raw:false}", s.Value, s.Raw)
				}
			},
		},
		{
			name:   "raw string",
			source: "x = ~\"hello\"\n",
			check: func(t *testing.T, a *ast.AssignNode) {
				t.Helper()
				s := mustStringLiteral(t, a.Value)
				if s.Value != "hello" || !s.Raw {
					t.Fatalf("string = {Value:%q Raw:%v}, want {Value:hello Raw:true}", s.Value, s.Raw)
				}
			},
		},
		{
			name:   "bool true",
			source: "x = true\n",
			check: func(t *testing.T, a *ast.AssignNode) {
				t.Helper()
				b := mustBoolLiteral(t, a.Value)
				if !b.Value {
					t.Fatalf("bool value = %v, want true", b.Value)
				}
			},
		},
		{
			name:   "bool false",
			source: "x = false\n",
			check: func(t *testing.T, a *ast.AssignNode) {
				t.Helper()
				b := mustBoolLiteral(t, a.Value)
				if b.Value {
					t.Fatalf("bool value = %v, want false", b.Value)
				}
			},
		},
		{
			name:   "null",
			source: "x = null\n",
			check: func(t *testing.T, a *ast.AssignNode) {
				t.Helper()
				if _, ok := a.Value.(*ast.NullLiteral); !ok {
					t.Fatalf("value type = %T, want *ast.NullLiteral", a.Value)
				}
			},
		},
		{
			name:   "array with elements",
			source: "x = [1, 2, 3]\n",
			check: func(t *testing.T, a *ast.AssignNode) {
				t.Helper()
				arr := mustArrayLiteral(t, a.Value)
				if len(arr.Elements) != 3 {
					t.Fatalf("array elements = %d, want 3", len(arr.Elements))
				}
			},
		},
		{
			name:   "empty array",
			source: "x = []\n",
			check: func(t *testing.T, a *ast.AssignNode) {
				t.Helper()
				arr := mustArrayLiteral(t, a.Value)
				if len(arr.Elements) != 0 {
					t.Fatalf("array elements = %d, want 0", len(arr.Elements))
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			program, errs := parseProgramFromSource(t, tc.source)
			assertNoParseErrors(t, errs)
			assign := mustAssignStmt(t, program, 0)
			tc.check(t, assign)
		})
	}
}

func TestParseExpressionPrecedence(t *testing.T) {
	t.Parallel()

	program, errs := parseProgramFromSource(t, "x = 1 + 2 * 3\n")
	assertNoParseErrors(t, errs)

	assign := mustAssignStmt(t, program, 0)
	root := mustBinaryNode(t, assign.Value)
	if root.Operator != lexer.TOKEN_PLUS {
		t.Fatalf("root operator = %v, want %v", root.Operator, lexer.TOKEN_PLUS)
	}
	left := mustNumberLiteral(t, root.Left)
	if left.Value != 1 {
		t.Fatalf("left number = %v, want 1", left.Value)
	}
	right := mustBinaryNode(t, root.Right)
	if right.Operator != lexer.TOKEN_STAR {
		t.Fatalf("right operator = %v, want %v", right.Operator, lexer.TOKEN_STAR)
	}
	rightLeft := mustNumberLiteral(t, right.Left)
	rightRight := mustNumberLiteral(t, right.Right)
	if rightLeft.Value != 2 || rightRight.Value != 3 {
		t.Fatalf("right operands = (%v, %v), want (2, 3)", rightLeft.Value, rightRight.Value)
	}
}

func TestParseAllComparisonOperators(t *testing.T) {
	t.Parallel()

	tests := []struct {
		op     string
		token  lexer.TokenType
		source string
	}{
		{op: "==", token: lexer.TOKEN_EQ, source: "x = 1 == 2\n"},
		{op: "!=", token: lexer.TOKEN_NEQ, source: "x = 1 != 2\n"},
		{op: "<", token: lexer.TOKEN_LT, source: "x = 1 < 2\n"},
		{op: ">", token: lexer.TOKEN_GT, source: "x = 1 > 2\n"},
		{op: "<=", token: lexer.TOKEN_LTE, source: "x = 1 <= 2\n"},
		{op: ">=", token: lexer.TOKEN_GTE, source: "x = 1 >= 2\n"},
	}

	for _, tc := range tests {
		t.Run(tc.op, func(t *testing.T) {
			t.Parallel()
			program, errs := parseProgramFromSource(t, tc.source)
			assertNoParseErrors(t, errs)
			assign := mustAssignStmt(t, program, 0)
			b := mustBinaryNode(t, assign.Value)
			if b.Operator != tc.token {
				t.Fatalf("operator = %v, want %v", b.Operator, tc.token)
			}
		})
	}
}

func TestParseLogicalOperatorsAndPrecedence(t *testing.T) {
	t.Parallel()

	program, errs := parseProgramFromSource(t, "x = not a and b or is c\n")
	assertNoParseErrors(t, errs)

	assign := mustAssignStmt(t, program, 0)
	orNode := mustBinaryNode(t, assign.Value)
	if orNode.Operator != lexer.TOKEN_OR {
		t.Fatalf("root operator = %v, want %v", orNode.Operator, lexer.TOKEN_OR)
	}

	andNode := mustBinaryNode(t, orNode.Left)
	if andNode.Operator != lexer.TOKEN_AND {
		t.Fatalf("left operator = %v, want %v", andNode.Operator, lexer.TOKEN_AND)
	}

	notNode := mustUnaryNode(t, andNode.Left)
	if notNode.Operator != lexer.TOKEN_NOT {
		t.Fatalf("not operator = %v, want %v", notNode.Operator, lexer.TOKEN_NOT)
	}
	mustIdentNodeWithName(t, notNode.Operand, "a")
	mustIdentNodeWithName(t, andNode.Right, "b")

	isNode := mustUnaryNode(t, orNode.Right)
	if isNode.Operator != lexer.TOKEN_IS {
		t.Fatalf("is operator = %v, want %v", isNode.Operator, lexer.TOKEN_IS)
	}
	mustIdentNodeWithName(t, isNode.Operand, "c")
}

func TestParseComparisonBindsTighterThanAnd(t *testing.T) {
	t.Parallel()

	program, errs := parseProgramFromSource(t, "x = a == b and c\n")
	assertNoParseErrors(t, errs)

	assign := mustAssignStmt(t, program, 0)
	andNode := mustBinaryNode(t, assign.Value)
	if andNode.Operator != lexer.TOKEN_AND {
		t.Fatalf("root operator = %v, want %v", andNode.Operator, lexer.TOKEN_AND)
	}
	cmp := mustBinaryNode(t, andNode.Left)
	if cmp.Operator != lexer.TOKEN_EQ {
		t.Fatalf("comparison operator = %v, want %v", cmp.Operator, lexer.TOKEN_EQ)
	}
	mustIdentNodeWithName(t, cmp.Left, "a")
	mustIdentNodeWithName(t, cmp.Right, "b")
	mustIdentNodeWithName(t, andNode.Right, "c")
}

func TestParseStarStarOperator(t *testing.T) {
	t.Parallel()

	program, errs := parseProgramFromSource(t, "x = 10 ** 3\n")
	assertNoParseErrors(t, errs)

	assign := mustAssignStmt(t, program, 0)
	b := mustBinaryNode(t, assign.Value)
	if b.Operator != lexer.TOKEN_STARSTAR {
		t.Fatalf("operator = %v, want %v", b.Operator, lexer.TOKEN_STARSTAR)
	}
	if mustNumberLiteral(t, b.Left).Value != 10 || mustNumberLiteral(t, b.Right).Value != 3 {
		t.Fatalf("operands should be 10 and 3")
	}
}

func TestParseIfElseStmt(t *testing.T) {
	t.Parallel()

	source := "if x == 1 }\ninput x\n{ else }\ninput y\n{\n"
	program, errs := parseProgramFromSource(t, source)
	assertNoParseErrors(t, errs)

	if len(program.Statements) != 1 {
		t.Fatalf("statement count = %d, want 1", len(program.Statements))
	}

	n, ok := program.Statements[0].(*ast.IfNode)
	if !ok {
		t.Fatalf("statement type = %T, want *ast.IfNode", program.Statements[0])
	}

	cond := mustBinaryNode(t, n.Condition)
	if cond.Operator != lexer.TOKEN_EQ {
		t.Fatalf("if condition operator = %v, want %v", cond.Operator, lexer.TOKEN_EQ)
	}
	mustIdentNodeWithName(t, cond.Left, "x")
	right := mustNumberLiteral(t, cond.Right)
	if right.Value != 1 {
		t.Fatalf("if condition right number = %v, want 1", right.Value)
	}

	if n.Consequence == nil || n.Alternative == nil {
		t.Fatalf("if consequence/alternative must be non-nil")
	}
	if len(n.Consequence.Statements) != 1 || len(n.Alternative.Statements) != 1 {
		t.Fatalf("if body sizes = (%d, %d), want (1,1)", len(n.Consequence.Statements), len(n.Alternative.Statements))
	}
}

func TestParseIfWithoutElse(t *testing.T) {
	t.Parallel()

	source := "if x }\ninput x\n{\n"
	program, errs := parseProgramFromSource(t, source)
	assertNoParseErrors(t, errs)

	n, ok := program.Statements[0].(*ast.IfNode)
	if !ok {
		t.Fatalf("statement type = %T, want *ast.IfNode", program.Statements[0])
	}
	if n.Alternative != nil {
		t.Fatalf("if alternative = %#v, want nil", n.Alternative)
	}
	if n.Consequence == nil || len(n.Consequence.Statements) != 1 {
		t.Fatalf("if consequence should contain one statement")
	}
}

func TestParseNestedBlocks(t *testing.T) {
	t.Parallel()

	source := "if x }\nwhile y }\n{\n{\n"
	program, errs := parseProgramFromSource(t, source)
	assertNoParseErrors(t, errs)

	ifNode, ok := program.Statements[0].(*ast.IfNode)
	if !ok {
		t.Fatalf("statement[0] type = %T, want *ast.IfNode", program.Statements[0])
	}
	if ifNode.Consequence == nil || len(ifNode.Consequence.Statements) != 1 {
		t.Fatalf("if consequence should contain one statement")
	}
	if _, ok := ifNode.Consequence.Statements[0].(*ast.WhileNode); !ok {
		t.Fatalf("nested statement type = %T, want *ast.WhileNode", ifNode.Consequence.Statements[0])
	}
}

func TestParseFuncDefAndCall(t *testing.T) {
	t.Parallel()

	source := "call add(a, b) }\ndiscard a - b\n{\ndefine add(3, 7)\n"
	program, errs := parseProgramFromSource(t, source)
	assertNoParseErrors(t, errs)

	if len(program.Statements) != 2 {
		t.Fatalf("statement count = %d, want 2", len(program.Statements))
	}

	def, ok := program.Statements[0].(*ast.FuncDefNode)
	if !ok {
		t.Fatalf("statement[0] type = %T, want *ast.FuncDefNode", program.Statements[0])
	}
	if def.Name != "add" {
		t.Fatalf("func name = %q, want add", def.Name)
	}
	if len(def.Params) != 2 || def.Params[0] != "a" || def.Params[1] != "b" {
		t.Fatalf("params = %#v, want [a b]", def.Params)
	}

	callStmt, ok := program.Statements[1].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("statement[1] type = %T, want *ast.ExprStmt", program.Statements[1])
	}
	call, ok := callStmt.Expr.(*ast.FuncCallNode)
	if !ok {
		t.Fatalf("expr type = %T, want *ast.FuncCallNode", callStmt.Expr)
	}
	if call.Name != "add" || len(call.Args) != 2 {
		t.Fatalf("call = {name:%q args:%d}, want {name:add args:2}", call.Name, len(call.Args))
	}
	arg0 := mustNumberLiteral(t, call.Args[0])
	arg1 := mustNumberLiteral(t, call.Args[1])
	if arg0.Value != 3 || arg1.Value != 7 {
		t.Fatalf("args = (%v, %v), want (3, 7)", arg0.Value, arg1.Value)
	}
}

func TestParseDefineAsExpression(t *testing.T) {
	t.Parallel()

	program, errs := parseProgramFromSource(t, "x = define add(1, 2)\n")
	assertNoParseErrors(t, errs)

	assign := mustAssignStmt(t, program, 0)
	call, ok := assign.Value.(*ast.FuncCallNode)
	if !ok {
		t.Fatalf("assign value type = %T, want *ast.FuncCallNode", assign.Value)
	}
	if call.Name != "add" || len(call.Args) != 2 {
		t.Fatalf("call = {name:%q args:%d}, want {name:add args:2}", call.Name, len(call.Args))
	}
}

func TestParseGroupedAndUnaryExpressions(t *testing.T) {
	t.Parallel()

	program, errs := parseProgramFromSource(t, "x = -(1 + 2)\n")
	assertNoParseErrors(t, errs)

	assign := mustAssignStmt(t, program, 0)
	u := mustUnaryNode(t, assign.Value)
	if u.Operator != lexer.TOKEN_MINUS {
		t.Fatalf("unary operator = %v, want %v", u.Operator, lexer.TOKEN_MINUS)
	}
	inner := mustBinaryNode(t, u.Operand)
	if inner.Operator != lexer.TOKEN_PLUS {
		t.Fatalf("inner operator = %v, want %v", inner.Operator, lexer.TOKEN_PLUS)
	}
	if mustNumberLiteral(t, inner.Left).Value != 1 || mustNumberLiteral(t, inner.Right).Value != 2 {
		t.Fatalf("inner operands should be 1 and 2")
	}
}

func TestParseLeftAssociativeAddSub(t *testing.T) {
	t.Parallel()

	program, errs := parseProgramFromSource(t, "x = 1 + 2 - 3\n")
	assertNoParseErrors(t, errs)

	assign := mustAssignStmt(t, program, 0)
	root := mustBinaryNode(t, assign.Value)
	if root.Operator != lexer.TOKEN_MINUS {
		t.Fatalf("root operator = %v, want %v", root.Operator, lexer.TOKEN_MINUS)
	}
	left := mustBinaryNode(t, root.Left)
	if left.Operator != lexer.TOKEN_PLUS {
		t.Fatalf("left operator = %v, want %v", left.Operator, lexer.TOKEN_PLUS)
	}
	if mustNumberLiteral(t, root.Right).Value != 3 {
		t.Fatalf("root right operand should be 3")
	}
}

func TestParseReturnWithoutValue(t *testing.T) {
	t.Parallel()

	program, errs := parseProgramFromSource(t, "return\n")
	assertNoParseErrors(t, errs)

	ret, ok := program.Statements[0].(*ast.ReturnNode)
	if !ok {
		t.Fatalf("statement type = %T, want *ast.ReturnNode", program.Statements[0])
	}
	if ret.Value != nil {
		t.Fatalf("return value = %#v, want nil", ret.Value)
	}
}

func TestParsePrintStatementAndExpression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		check  func(t *testing.T, program *ast.ProgramNode)
	}{
		{
			name:   "standalone print statement",
			source: "print\n",
			check: func(t *testing.T, program *ast.ProgramNode) {
				t.Helper()
				if len(program.Statements) != 1 {
					t.Fatalf("statement count = %d, want 1", len(program.Statements))
				}
				mustExtractPrintNodeFromStmt(t, program.Statements[0])
			},
		},
		{
			name:   "print with prompt statement",
			source: "print \"Enter name: \"\n",
			check: func(t *testing.T, program *ast.ProgramNode) {
				t.Helper()
				p := mustExtractPrintNodeFromStmt(t, program.Statements[0])
				s := mustStringLiteral(t, p.Prompt)
				if s.Value != "Enter name: " {
					t.Fatalf("prompt value = %q, want %q", s.Value, "Enter name: ")
				}
			},
		},
		{
			name:   "print as expression",
			source: "x = print\n",
			check: func(t *testing.T, program *ast.ProgramNode) {
				t.Helper()
				a := mustAssignStmt(t, program, 0)
				if _, ok := a.Value.(*ast.PrintNode); !ok {
					t.Fatalf("assign value type = %T, want *ast.PrintNode", a.Value)
				}
			},
		},
		{
			name:   "print with prompt expression",
			source: "x = print ~\"Enter name: \"\n",
			check: func(t *testing.T, program *ast.ProgramNode) {
				t.Helper()
				a := mustAssignStmt(t, program, 0)
				p, ok := a.Value.(*ast.PrintNode)
				if !ok {
					t.Fatalf("assign value type = %T, want *ast.PrintNode", a.Value)
				}
				s := mustStringLiteral(t, p.Prompt)
				if s.Value != "Enter name: " || !s.Raw {
					t.Fatalf("prompt string = {Value:%q Raw:%v}, want {Value:Enter name:  Raw:true}", s.Value, s.Raw)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			program, errs := parseProgramFromSource(t, tc.source)
			assertNoParseErrors(t, errs)
			tc.check(t, program)
		})
	}
}

func TestParseStatementFormsTableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		source    string
		wantType  interface{}
		wantCount int
		verify    func(t *testing.T, stmt ast.Statement)
	}{
		{
			name:      "while",
			source:    "while x }\n{\n",
			wantType:  &ast.WhileNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.WhileNode)
				mustIdentNodeWithName(t, n.Condition, "x")
				if n.Body == nil {
					t.Fatalf("while body is nil")
				}
			},
		},
		{
			name:      "for",
			source:    "for i in arr }\n{\n",
			wantType:  &ast.ForNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.ForNode)
				if n.Variable != "i" {
					t.Fatalf("for variable = %q, want i", n.Variable)
				}
				mustIdentNodeWithName(t, n.Iterable, "arr")
				if n.Body == nil {
					t.Fatalf("for body is nil")
				}
			},
		},
		{
			name:      "del",
			source:    "del x\n",
			wantType:  &ast.DelNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.DelNode)
				if n.Name != "x" {
					t.Fatalf("del name = %q, want x", n.Name)
				}
			},
		},
		{
			name:      "scope global",
			source:    "global x\n",
			wantType:  &ast.ScopeNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.ScopeNode)
				if n.Keyword != "global" || n.Name != "x" {
					t.Fatalf("scope = {Keyword:%q Name:%q}, want {global x}", n.Keyword, n.Name)
				}
			},
		},
		{
			name:      "scope local",
			source:    "local x\n",
			wantType:  &ast.ScopeNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.ScopeNode)
				if n.Keyword != "local" || n.Name != "x" {
					t.Fatalf("scope = {Keyword:%q Name:%q}, want {local x}", n.Keyword, n.Name)
				}
			},
		},
		{
			name:      "return",
			source:    "return 1\n",
			wantType:  &ast.ReturnNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.ReturnNode)
				if mustNumberLiteral(t, n.Value).Value != 1 {
					t.Fatalf("return value should be 1")
				}
			},
		},
		{
			name:      "discard",
			source:    "discard 1\n",
			wantType:  &ast.DiscardNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.DiscardNode)
				if mustNumberLiteral(t, n.Value).Value != 1 {
					t.Fatalf("discard value should be 1")
				}
			},
		},
		{
			name:      "input",
			source:    "input 1\n",
			wantType:  &ast.InputNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.InputNode)
				if mustNumberLiteral(t, n.Value).Value != 1 {
					t.Fatalf("input value should be 1")
				}
			},
		},
		{
			name:      "import",
			source:    "import math\n",
			wantType:  &ast.ImportNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.ImportNode)
				if n.Name != "math" {
					t.Fatalf("import name = %q, want math", n.Name)
				}
			},
		},
		{
			name:      "export",
			source:    "export math\n",
			wantType:  &ast.ExportNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.ExportNode)
				if n.Name != "math" {
					t.Fatalf("export name = %q, want math", n.Name)
				}
			},
		},
		{
			name:      "stop",
			source:    "stop\n",
			wantType:  &ast.StopNode{},
			wantCount: 1,
			verify:    func(t *testing.T, stmt ast.Statement) {},
		},
		{
			name:      "raise with message",
			source:    "raise MyErr(\"hi\")\n",
			wantType:  &ast.RaiseNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.RaiseNode)
				if n.ErrorName != "MyErr" {
					t.Fatalf("raise error name = %q, want MyErr", n.ErrorName)
				}
				s := mustStringLiteral(t, n.Message)
				if s.Value != "hi" {
					t.Fatalf("raise message = %q, want hi", s.Value)
				}
			},
		},
		{
			name:      "raise without message",
			source:    "raise MyErr\n",
			wantType:  &ast.RaiseNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.RaiseNode)
				if n.ErrorName != "MyErr" {
					t.Fatalf("raise error name = %q, want MyErr", n.ErrorName)
				}
				if n.Message != nil {
					t.Fatalf("raise message = %#v, want nil", n.Message)
				}
			},
		},
		{
			name:      "break",
			source:    "break\n",
			wantType:  &ast.BreakNode{},
			wantCount: 1,
			verify:    func(t *testing.T, stmt ast.Statement) {},
		},
		{
			name:      "continue",
			source:    "continue\n",
			wantType:  &ast.ContinueNode{},
			wantCount: 1,
			verify:    func(t *testing.T, stmt ast.Statement) {},
		},
		{
			name:      "try except finally",
			source:    "try }\n{ except }\n{ finally }\n{\n",
			wantType:  &ast.TryNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.TryNode)
				if n.Except == nil {
					t.Fatalf("try except clause is nil")
				}
				if n.Finally == nil {
					t.Fatalf("try finally clause is nil")
				}
			},
		},
		{
			name:      "try only",
			source:    "try }\n{\n",
			wantType:  &ast.TryNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.TryNode)
				if n.Except != nil || n.Finally != nil {
					t.Fatalf("try-only should have nil except/finally")
				}
			},
		},
		{
			name:      "try except only",
			source:    "try }\n{ except }\n{\n",
			wantType:  &ast.TryNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.TryNode)
				if n.Except == nil {
					t.Fatalf("except should be non-nil")
				}
				if n.Finally != nil {
					t.Fatalf("finally should be nil")
				}
			},
		},
		{
			name:      "try except named var",
			source:    "try }\n{ except(e) }\n{\n",
			wantType:  &ast.TryNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.TryNode)
				if n.Except == nil {
					t.Fatalf("except should be non-nil")
				}
				if n.Except.ErrVar != "e" {
					t.Fatalf("except err var = %q, want e", n.Except.ErrVar)
				}
			},
		},
		{
			name:      "match case wildcard",
			source:    "match x }\ncase 1 }\n{\ncase _ }\n{\n{\n",
			wantType:  &ast.MatchNode{},
			wantCount: 1,
			verify: func(t *testing.T, stmt ast.Statement) {
				t.Helper()
				n := stmt.(*ast.MatchNode)
				if len(n.Cases) != 2 {
					t.Fatalf("match case count = %d, want 2", len(n.Cases))
				}
				if n.Cases[1].Pattern != nil {
					t.Fatalf("wildcard case pattern should be nil")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			program, errs := parseProgramFromSource(t, tc.source)
			assertNoParseErrors(t, errs)
			if len(program.Statements) != tc.wantCount {
				t.Fatalf("statement count = %d, want %d", len(program.Statements), tc.wantCount)
			}
			if reflect.TypeOf(program.Statements[0]) != reflect.TypeOf(tc.wantType) {
				t.Fatalf("statement type = %T, want %T", program.Statements[0], tc.wantType)
			}
			tc.verify(t, program.Statements[0])
		})
	}
}

func TestParseDotAccessExpression(t *testing.T) {
	t.Parallel()
	t.Skip("dot-access AST/parser support not implemented yet; add positive test when feature lands")
}

func TestParsePositionsOnNodes(t *testing.T) {
	t.Parallel()

	source := "x = 1\nif y }\ninput y\n{\n"
	program, errs := parseProgramFromSource(t, source)
	assertNoParseErrors(t, errs)

	assign := mustAssignStmt(t, program, 0)
	if p := assign.Pos(); p.Line != 1 || p.Column != 1 {
		t.Fatalf("assign pos = (%d,%d), want (1,1)", p.Line, p.Column)
	}
	if num := mustNumberLiteral(t, assign.Value); num.Pos().Line != 1 {
		t.Fatalf("number pos line = %d, want 1", num.Pos().Line)
	}

	ifNode, ok := program.Statements[1].(*ast.IfNode)
	if !ok {
		t.Fatalf("statement[1] type = %T, want *ast.IfNode", program.Statements[1])
	}
	if p := ifNode.Pos(); p.Line != 2 || p.Column != 1 {
		t.Fatalf("if pos = (%d,%d), want (2,1)", p.Line, p.Column)
	}
}

func TestParseEmptySource(t *testing.T) {
	t.Parallel()

	program, errs := parseProgramFromSource(t, "")
	assertNoParseErrors(t, errs)
	if program == nil {
		t.Fatalf("program should not be nil")
	}
	if len(program.Statements) != 0 {
		t.Fatalf("statement count = %d, want 0", len(program.Statements))
	}
}

func TestParseCollectsErrorsAndReturnsPartialAST(t *testing.T) {
	t.Parallel()

	source := "x =\ny = 2\n"
	program, errs := parseProgramFromSource(t, source)

	if len(errs) == 0 {
		t.Fatalf("expected parse errors, got none")
	}
	assertAllParseErrorsAreWorngSyntax(t, errs)
	if program == nil {
		t.Fatalf("program should not be nil on parse error")
	}
	if len(program.Statements) != 1 {
		t.Fatalf("expected one recovered statement, got %d", len(program.Statements))
	}
	recovered, ok := program.Statements[0].(*ast.AssignNode)
	if !ok || recovered.Name != "y" {
		t.Fatalf("recovered statement = %T %#v, want assign to y", program.Statements[0], program.Statements[0])
	}
}

func TestParseMultipleErrorsCollected(t *testing.T) {
	t.Parallel()

	// We expect one syntax error per malformed input line with panic-mode recovery
	// continuing to subsequent statements.
	source := "x =\nif }\nreturn (\n"
	_, errs := parseProgramFromSource(t, source)

	if len(errs) != 3 {
		t.Fatalf("expected exactly 3 errors, got %d", len(errs))
	}
	assertAllParseErrorsAreWorngSyntax(t, errs)
}

func FuzzParse(f *testing.F) {
	f.Add("x = 1\n")
	f.Add("if x }\ninput x\n{\n")
	f.Add("call add(a,b) }\ndiscard a-b\n{\n")

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("parser panicked: %v", r)
			}
		}()

		tokens := lexer.New(input).Tokenize()
		p := New(tokens)
		_, _ = p.Parse()
	})
}

func parseProgramFromSource(t *testing.T, source string) (*ast.ProgramNode, []error) {
	t.Helper()
	tokens := lexer.New(source).Tokenize()
	p := New(tokens)
	return p.Parse()
}

func assertNoParseErrors(t *testing.T, errs []error) {
	t.Helper()
	if len(errs) != 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
}

func assertAllParseErrorsAreWorngSyntax(t *testing.T, errs []error) {
	t.Helper()
	for i, err := range errs {
		we, ok := err.(*diagnostics.WorngError)
		if !ok {
			t.Fatalf("err[%d] type = %T, want *diagnostics.WorngError", i, err)
		}
		if we.Diag.Code != diagnostics.SyntaxError.Code {
			t.Fatalf("err[%d] code = %d, want %d", i, we.Diag.Code, diagnostics.SyntaxError.Code)
		}
	}
}

func mustAssignStmt(t *testing.T, program *ast.ProgramNode, idx int) *ast.AssignNode {
	t.Helper()
	if len(program.Statements) <= idx {
		t.Fatalf("statement count = %d, need index %d", len(program.Statements), idx)
	}
	stmt, ok := program.Statements[idx].(*ast.AssignNode)
	if !ok {
		t.Fatalf("statement[%d] type = %T, want *ast.AssignNode", idx, program.Statements[idx])
	}
	return stmt
}

func mustBinaryNode(t *testing.T, expr ast.Expression) *ast.BinaryNode {
	t.Helper()
	n, ok := expr.(*ast.BinaryNode)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.BinaryNode", expr)
	}
	return n
}

func mustUnaryNode(t *testing.T, expr ast.Expression) *ast.UnaryNode {
	t.Helper()
	n, ok := expr.(*ast.UnaryNode)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.UnaryNode", expr)
	}
	return n
}

func mustNumberLiteral(t *testing.T, expr ast.Expression) *ast.NumberLiteral {
	t.Helper()
	n, ok := expr.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.NumberLiteral", expr)
	}
	return n
}

func mustStringLiteral(t *testing.T, expr ast.Expression) *ast.StringLiteral {
	t.Helper()
	s, ok := expr.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.StringLiteral", expr)
	}
	return s
}

func mustBoolLiteral(t *testing.T, expr ast.Expression) *ast.BoolLiteral {
	t.Helper()
	b, ok := expr.(*ast.BoolLiteral)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.BoolLiteral", expr)
	}
	return b
}

func mustArrayLiteral(t *testing.T, expr ast.Expression) *ast.ArrayLiteral {
	t.Helper()
	a, ok := expr.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.ArrayLiteral", expr)
	}
	return a
}

func mustIdentNodeWithName(t *testing.T, expr ast.Expression, want string) *ast.IdentNode {
	t.Helper()
	id, ok := expr.(*ast.IdentNode)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.IdentNode", expr)
	}
	if id.Name != want {
		t.Fatalf("identifier = %q, want %q", id.Name, want)
	}
	return id
}

func mustExtractPrintNodeFromStmt(t *testing.T, stmt ast.Statement) *ast.PrintNode {
	t.Helper()
	if p, ok := stmt.(*ast.PrintNode); ok {
		return p
	}
	es, ok := stmt.(*ast.ExprStmt)
	if !ok {
		t.Fatalf("statement type = %T, want *ast.PrintNode or *ast.ExprStmt", stmt)
	}
	p, ok := es.Expr.(*ast.PrintNode)
	if !ok {
		t.Fatalf("expr type = %T, want *ast.PrintNode", es.Expr)
	}
	return p
}
