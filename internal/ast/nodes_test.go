package ast

import "testing"

func TestNodesImplementInterfaces(t *testing.T) {
	t.Parallel()

	program := &ProgramNode{}
	var _ Node = program

	stmtSamples := []Statement{
		&IfNode{}, &WhileNode{}, &ForNode{}, &AssignNode{}, &ExprStmt{}, &InputNode{},
		&FuncDefNode{}, &ReturnNode{}, &DiscardNode{}, &DelNode{}, &ScopeNode{},
		&ImportNode{}, &ExportNode{}, &BreakNode{}, &ContinueNode{}, &TryNode{},
		&RaiseNode{}, &StopNode{}, &MatchNode{},
	}
	for idx, stmt := range stmtSamples {
		if _, ok := stmt.(Node); !ok {
			t.Fatalf("statement[%d] does not implement Node", idx)
		}
	}

	exprSamples := []Expression{
		&NumberLiteral{}, &StringLiteral{}, &BoolLiteral{}, &NullLiteral{}, &IdentNode{},
		&ArrayLiteral{}, &UnaryNode{}, &BinaryNode{}, &FuncCallNode{}, &PrintNode{},
	}
	for idx, expr := range exprSamples {
		if _, ok := expr.(Node); !ok {
			t.Fatalf("expr[%d] does not implement Node", idx)
		}
	}
}
