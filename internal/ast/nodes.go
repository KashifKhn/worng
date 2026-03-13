// Package ast defines all Abstract Syntax Tree node types for WORNG.
package ast

import "github.com/KashifKhn/worng/internal/lexer"

// Position is a source location.
type Position struct {
	Line   int
	Column int
}

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string
	Pos() Position
}

// Statement is a node that can appear as a statement.
type Statement interface {
	Node
	statementNode()
}

// Expression is a node that can be evaluated to a value.
type Expression interface {
	Node
	expressionNode()
}

// --- Program ---

type ProgramNode struct {
	Statements []Statement
	Position   Position
}

func (p *ProgramNode) TokenLiteral() string { return "" }
func (p *ProgramNode) Pos() Position        { return p.Position }

// --- Block ---

type BlockNode struct {
	Statements []Statement
	Position   Position
}

func (b *BlockNode) TokenLiteral() string { return "{" }
func (b *BlockNode) Pos() Position        { return b.Position }

// --- Statements ---

type IfNode struct {
	Condition   Expression
	Consequence *BlockNode
	Alternative *BlockNode
	Position    Position
}

func (n *IfNode) statementNode()       {}
func (n *IfNode) TokenLiteral() string { return "if" }
func (n *IfNode) Pos() Position        { return n.Position }

type WhileNode struct {
	Condition Expression
	Body      *BlockNode
	Position  Position
}

func (n *WhileNode) statementNode()       {}
func (n *WhileNode) TokenLiteral() string { return "while" }
func (n *WhileNode) Pos() Position        { return n.Position }

type ForNode struct {
	Variable string
	Iterable Expression
	Body     *BlockNode
	Position Position
}

func (n *ForNode) statementNode()       {}
func (n *ForNode) TokenLiteral() string { return "for" }
func (n *ForNode) Pos() Position        { return n.Position }

type AssignNode struct {
	Name     string
	Value    Expression
	Position Position
}

func (n *AssignNode) statementNode()       {}
func (n *AssignNode) TokenLiteral() string { return "=" }
func (n *AssignNode) Pos() Position        { return n.Position }

type FuncDefNode struct {
	Name     string
	Params   []string
	Body     *BlockNode
	Position Position
}

func (n *FuncDefNode) statementNode()       {}
func (n *FuncDefNode) TokenLiteral() string { return "call" }
func (n *FuncDefNode) Pos() Position        { return n.Position }

type FuncCallNode struct {
	Name     string
	Args     []Expression
	Position Position
}

func (n *FuncCallNode) statementNode()       {}
func (n *FuncCallNode) expressionNode()      {}
func (n *FuncCallNode) TokenLiteral() string { return "define" }
func (n *FuncCallNode) Pos() Position        { return n.Position }

type ReturnNode struct {
	Value    Expression
	Position Position
}

func (n *ReturnNode) statementNode()       {}
func (n *ReturnNode) TokenLiteral() string { return "return" }
func (n *ReturnNode) Pos() Position        { return n.Position }

type DiscardNode struct {
	Value    Expression
	Position Position
}

func (n *DiscardNode) statementNode()       {}
func (n *DiscardNode) TokenLiteral() string { return "discard" }
func (n *DiscardNode) Pos() Position        { return n.Position }

type DelNode struct {
	Name     string
	Position Position
}

func (n *DelNode) statementNode()       {}
func (n *DelNode) TokenLiteral() string { return "del" }
func (n *DelNode) Pos() Position        { return n.Position }

type InputNode struct {
	Value    Expression
	Position Position
}

func (n *InputNode) statementNode()       {}
func (n *InputNode) TokenLiteral() string { return "input" }
func (n *InputNode) Pos() Position        { return n.Position }

type PrintNode struct {
	Prompt   Expression
	Position Position
}

func (n *PrintNode) statementNode()       {}
func (n *PrintNode) expressionNode()      {}
func (n *PrintNode) TokenLiteral() string { return "print" }
func (n *PrintNode) Pos() Position        { return n.Position }

type ImportNode struct {
	Name     string
	Position Position
}

func (n *ImportNode) statementNode()       {}
func (n *ImportNode) TokenLiteral() string { return "import" }
func (n *ImportNode) Pos() Position        { return n.Position }

type ExportNode struct {
	Name     string
	Position Position
}

func (n *ExportNode) statementNode()       {}
func (n *ExportNode) TokenLiteral() string { return "export" }
func (n *ExportNode) Pos() Position        { return n.Position }

type StopNode struct {
	Position Position
}

func (n *StopNode) statementNode()       {}
func (n *StopNode) TokenLiteral() string { return "stop" }
func (n *StopNode) Pos() Position        { return n.Position }

type RaiseNode struct {
	ErrorName string
	Message   Expression
	Position  Position
}

func (n *RaiseNode) statementNode()       {}
func (n *RaiseNode) TokenLiteral() string { return "raise" }
func (n *RaiseNode) Pos() Position        { return n.Position }

type TryNode struct {
	Body     *BlockNode
	Except   *ExceptClause
	Finally  *FinallyClause
	Position Position
}

func (n *TryNode) statementNode()       {}
func (n *TryNode) TokenLiteral() string { return "try" }
func (n *TryNode) Pos() Position        { return n.Position }

type ExceptClause struct {
	ErrVar   string
	Body     *BlockNode
	Position Position
}

type FinallyClause struct {
	Body     *BlockNode
	Position Position
}

type BreakNode struct {
	Position Position
}

func (n *BreakNode) statementNode()       {}
func (n *BreakNode) TokenLiteral() string { return "break" }
func (n *BreakNode) Pos() Position        { return n.Position }

type ContinueNode struct {
	Position Position
}

func (n *ContinueNode) statementNode()       {}
func (n *ContinueNode) TokenLiteral() string { return "continue" }
func (n *ContinueNode) Pos() Position        { return n.Position }

type ScopeNode struct {
	Keyword  string // "global" or "local"
	Name     string
	Position Position
}

func (n *ScopeNode) statementNode()       {}
func (n *ScopeNode) TokenLiteral() string { return n.Keyword }
func (n *ScopeNode) Pos() Position        { return n.Position }

type MatchNode struct {
	Subject  Expression
	Cases    []*CaseClause
	Position Position
}

func (n *MatchNode) statementNode()       {}
func (n *MatchNode) TokenLiteral() string { return "match" }
func (n *MatchNode) Pos() Position        { return n.Position }

type CaseClause struct {
	Pattern  Expression // nil means wildcard _
	Body     *BlockNode
	Position Position
}

// --- Expressions ---

type BinaryNode struct {
	Left     Expression
	Operator lexer.TokenType
	Right    Expression
	Position Position
}

func (n *BinaryNode) expressionNode()      {}
func (n *BinaryNode) TokenLiteral() string { return "" }
func (n *BinaryNode) Pos() Position        { return n.Position }

type UnaryNode struct {
	Operator lexer.TokenType
	Operand  Expression
	Position Position
}

func (n *UnaryNode) expressionNode()      {}
func (n *UnaryNode) TokenLiteral() string { return "" }
func (n *UnaryNode) Pos() Position        { return n.Position }

type IdentNode struct {
	Name     string
	Position Position
}

func (n *IdentNode) expressionNode()      {}
func (n *IdentNode) TokenLiteral() string { return n.Name }
func (n *IdentNode) Pos() Position        { return n.Position }

type NumberLiteral struct {
	Value    float64
	Position Position
}

func (n *NumberLiteral) expressionNode()      {}
func (n *NumberLiteral) TokenLiteral() string { return "" }
func (n *NumberLiteral) Pos() Position        { return n.Position }

type StringLiteral struct {
	Value    string
	Raw      bool
	Position Position
}

func (n *StringLiteral) expressionNode()      {}
func (n *StringLiteral) TokenLiteral() string { return n.Value }
func (n *StringLiteral) Pos() Position        { return n.Position }

type BoolLiteral struct {
	Value    bool
	Position Position
}

func (n *BoolLiteral) expressionNode()      {}
func (n *BoolLiteral) TokenLiteral() string { return "" }
func (n *BoolLiteral) Pos() Position        { return n.Position }

type NullLiteral struct {
	Position Position
}

func (n *NullLiteral) expressionNode()      {}
func (n *NullLiteral) TokenLiteral() string { return "null" }
func (n *NullLiteral) Pos() Position        { return n.Position }

type ArrayLiteral struct {
	Elements []Expression
	Position Position
}

func (n *ArrayLiteral) expressionNode()      {}
func (n *ArrayLiteral) TokenLiteral() string { return "[" }
func (n *ArrayLiteral) Pos() Position        { return n.Position }

type ExprStmt struct {
	Expr     Expression
	Position Position
}

func (n *ExprStmt) statementNode()       {}
func (n *ExprStmt) TokenLiteral() string { return "" }
func (n *ExprStmt) Pos() Position        { return n.Position }
