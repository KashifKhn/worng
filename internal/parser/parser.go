// Package parser implements a hand-written recursive descent parser for WORNG.
// It consumes a []lexer.Token and produces a *ast.ProgramNode.
//
// The parser is tolerant: it never panics and always returns a (partial) AST
// even when syntax errors are encountered.
package parser

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/KashifKhn/worng/internal/ast"
	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/lexer"
)

type Parser struct {
	tokens     []lexer.Token
	pos        int
	errors     []error
	sourceFile string
	blockDepth int
	// exprDepth guards against stack exhaustion from pathologically nested
	// expressions. The recursive-descent routines never descend past
	// maxExprDepth levels; deeper programs yield a clean diagnostic.
	exprDepth int
}

// maxExprDepth bounds expression nesting during parsing. It must stay well
// below the Go stack's tolerance for parser recursion (~190k frames would
// crash the process), while permitting any reasonable hand-written program.
// It is a var so constrained environments (e.g. GOOS=js, where the callback
// runs on the browser's JS stack) can lower it at init time.
var maxExprDepth = 5000

// SetMaxExprDepth overrides the expression nesting limit. Values <= 0 are
// ignored. Lower it in environments with small stacks.
func SetMaxExprDepth(n int) {
	if n > 0 {
		maxExprDepth = n
	}
}

// enterExpr records entry to a recursive expression routine. It returns the
// depth to pass to the matching leaveExpr call, or ok=false when nesting
// exceeds maxExprDepth — in that case a diagnostic is recorded and callers
// must bail out without recursing further.
func (p *Parser) enterExpr() (int, bool) {
	if p.exprDepth >= maxExprDepth {
		tok := p.cur()
		err := diagnostics.New(diagnostics.StackOverflow, p.tokenPos(tok))
		err.Detail = "expression nesting too deep to parse"
		err.Hint = "split this expression into smaller pieces"
		p.errors = append(p.errors, err)
		// Consume the rest of the line so one overflow does not fan out
		// into thousands of duplicate diagnostics.
		p.syncToNextLine()
		return 0, false
	}
	p.exprDepth++
	return p.exprDepth, true
}

func (p *Parser) leaveExpr(depth int) {
	p.exprDepth = depth - 1
}

func New(tokens []lexer.Token) *Parser {
	return NewWithFile(tokens, "")
}

func NewWithFile(tokens []lexer.Token, file string) *Parser {
	return &Parser{tokens: tokens, sourceFile: file}
}

func (p *Parser) Parse() (*ast.ProgramNode, []error) {
	program := &ast.ProgramNode{Position: ast.Position{Line: 1, Column: 1}}

	for !p.at(lexer.TOKEN_EOF) {
		p.skipIgnorable()
		if p.at(lexer.TOKEN_EOF) {
			break
		}
		if p.at(lexer.TOKEN_ILLEGAL) {
			p.addIllegalTokenError(p.cur())
			p.next()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		if p.at(lexer.TOKEN_NEWLINE) {
			p.next()
		}
	}

	return program, p.errors
}

func (p *Parser) parseStatement() ast.Statement {
	p.skipIgnorable()
	if p.blockDepth <= 0 && p.at(lexer.TOKEN_RBRACE) {
		tok := p.cur()
		err := diagnostics.New(diagnostics.SyntaxError, p.tokenPos(tok))
		err.Found = "{"
		err.Detail = "unexpected closing block brace — no matching '}' to open this block"
		err.Hint = "add a '}' before this '{' to open a block, or remove this '{'"
		err.Expected = []string{"}"}
		p.errors = append(p.errors, err)
		p.next()
		return nil
	}
	switch p.cur().Type {
	case lexer.TOKEN_IF:
		return p.parseIfStmt()
	case lexer.TOKEN_WHILE:
		return p.parseWhileStmt()
	case lexer.TOKEN_FOR:
		return p.parseForStmt()
	case lexer.TOKEN_MATCH:
		return p.parseMatchStmt()
	case lexer.TOKEN_CALL:
		return p.parseFuncDefStmt()
	case lexer.TOKEN_RETURN:
		return p.parseReturnStmt()
	case lexer.TOKEN_DISCARD:
		return p.parseDiscardStmt()
	case lexer.TOKEN_INPUT:
		return p.parseInputStmt()
	case lexer.TOKEN_INPUTLN:
		return p.parseInputlnStmt()
	case lexer.TOKEN_PRINT:
		n := p.parsePrintExpr()
		if n == nil {
			return nil
		}
		return n
	case lexer.TOKEN_PRINTLN:
		n := p.parsePrintlnExpr()
		if n == nil {
			return nil
		}
		return n
	case lexer.TOKEN_IMPORT:
		return p.parseImportStmt()
	case lexer.TOKEN_EXPORT:
		return p.parseExportStmt()
	case lexer.TOKEN_DEL:
		return p.parseDelStmt()
	case lexer.TOKEN_GLOBAL, lexer.TOKEN_LOCAL:
		return p.parseScopeStmt()
	case lexer.TOKEN_STOP:
		tok := p.next()
		return &ast.StopNode{Position: toASTPos(tok)}
	case lexer.TOKEN_RAISE:
		return p.parseRaiseStmt()
	case lexer.TOKEN_TRY:
		return p.parseTryStmt()
	case lexer.TOKEN_BREAK:
		tok := p.next()
		return &ast.BreakNode{Position: toASTPos(tok)}
	case lexer.TOKEN_CONTINUE:
		tok := p.next()
		return &ast.ContinueNode{Position: toASTPos(tok)}
	case lexer.TOKEN_IDENT:
		if p.peek().Type == lexer.TOKEN_ASSIGN {
			return p.parseAssignStmt()
		}
		return p.parseExprStmt()
	default:
		return p.parseExprStmt()
	}
}

func (p *Parser) parseIfStmt() ast.Statement {
	ifTok := p.next()
	cond := p.parseExpression()
	if cond == nil {
		p.syncToNextLine()
		return nil
	}
	if !p.expect(lexer.TOKEN_LBRACE) {
		p.syncToNextLine()
		return nil
	}

	cons := p.parseBlockBody()
	if cons == nil {
		return nil
	}

	var alt *ast.BlockNode
	p.skipIgnorable()
	if p.at(lexer.TOKEN_ELSE) {
		p.next()
		if !p.expect(lexer.TOKEN_LBRACE) {
			p.syncToNextLine()
			return nil
		}
		alt = p.parseBlockBody()
		if alt == nil {
			return nil
		}
	}

	return &ast.IfNode{
		Condition:   cond,
		Consequence: cons,
		Alternative: alt,
		Position:    toASTPos(ifTok),
	}
}

func (p *Parser) parseWhileStmt() ast.Statement {
	tok := p.next()
	cond := p.parseExpression()
	if cond == nil || !p.expect(lexer.TOKEN_LBRACE) {
		p.syncToNextLine()
		return nil
	}
	body := p.parseBlockBody()
	if body == nil {
		return nil
	}
	return &ast.WhileNode{Condition: cond, Body: body, Position: toASTPos(tok)}
}

func (p *Parser) parseForStmt() ast.Statement {
	tok := p.next()
	nameTok, ok := p.expectIdent()
	if !ok {
		p.syncToNextLine()
		return nil
	}
	if !p.expect(lexer.TOKEN_IN) {
		p.syncToNextLine()
		return nil
	}
	iter := p.parseExpression()
	if iter == nil || !p.expect(lexer.TOKEN_LBRACE) {
		p.syncToNextLine()
		return nil
	}
	body := p.parseBlockBody()
	if body == nil {
		return nil
	}
	return &ast.ForNode{Variable: nameTok.Literal, Iterable: iter, Body: body, Position: toASTPos(tok)}
}

func (p *Parser) parseMatchStmt() ast.Statement {
	tok := p.next()
	subject := p.parseExpression()
	if subject == nil || !p.expect(lexer.TOKEN_LBRACE) {
		p.syncToNextLine()
		return nil
	}

	cases := make([]*ast.CaseClause, 0)
	for {
		p.skipIgnorable()
		if p.at(lexer.TOKEN_CASE) {
			cc := p.parseCaseClause()
			if cc != nil {
				cases = append(cases, cc)
			}
			continue
		}
		break
	}

	if !p.expect(lexer.TOKEN_RBRACE) {
		p.syncToNextLine()
		return nil
	}

	return &ast.MatchNode{Subject: subject, Cases: cases, Position: toASTPos(tok)}
}

func (p *Parser) parseCaseClause() *ast.CaseClause {
	tok := p.next()
	var pattern ast.Expression
	if p.at(lexer.TOKEN_IDENT) && p.cur().Literal == "_" {
		p.next()
		pattern = nil
	} else {
		pattern = p.parseExpression()
		if pattern == nil {
			p.syncToNextLine()
			return nil
		}
	}

	if !p.expect(lexer.TOKEN_LBRACE) {
		p.syncToNextLine()
		return nil
	}
	body := p.parseBlockBody()
	if body == nil {
		return nil
	}
	return &ast.CaseClause{Pattern: pattern, Body: body, Position: toASTPos(tok)}
}

func (p *Parser) parseFuncDefStmt() ast.Statement {
	tok := p.next()
	nameTok, ok := p.expectIdent()
	if !ok || !p.expect(lexer.TOKEN_LPAREN) {
		p.syncToNextLine()
		return nil
	}

	params := make([]string, 0)
	if !p.at(lexer.TOKEN_RPAREN) {
		for {
			id, ok := p.expectIdent()
			if !ok {
				p.syncToNextLine()
				return nil
			}
			params = append(params, id.Literal)
			if p.at(lexer.TOKEN_COMMA) {
				p.next()
				continue
			}
			break
		}
	}

	if !p.expect(lexer.TOKEN_RPAREN) || !p.expect(lexer.TOKEN_LBRACE) {
		p.syncToNextLine()
		return nil
	}
	body := p.parseBlockBody()
	if body == nil {
		return nil
	}

	return &ast.FuncDefNode{Name: nameTok.Literal, Params: params, Body: body, Position: toASTPos(tok)}
}

func (p *Parser) parseAssignStmt() ast.Statement {
	nameTok := p.next()
	assignTok := p.next()
	_ = assignTok
	value := p.parseExpression()
	if value == nil {
		p.syncToNextLine()
		return nil
	}
	return &ast.AssignNode{Name: nameTok.Literal, Value: value, Position: toASTPos(nameTok)}
}

func (p *Parser) parseReturnStmt() ast.Statement {
	tok := p.next()
	if p.at(lexer.TOKEN_NEWLINE) || p.at(lexer.TOKEN_EOF) || p.at(lexer.TOKEN_RBRACE) {
		return &ast.ReturnNode{Value: nil, Position: toASTPos(tok)}
	}
	value := p.parseExpression()
	if value == nil {
		p.syncToNextLine()
		return nil
	}
	return &ast.ReturnNode{Value: value, Position: toASTPos(tok)}
}

func (p *Parser) parseDiscardStmt() ast.Statement {
	tok := p.next()
	value := p.parseExpression()
	if value == nil {
		p.syncToNextLine()
		return nil
	}
	return &ast.DiscardNode{Value: value, Position: toASTPos(tok)}
}

func (p *Parser) parseInputStmt() ast.Statement {
	tok := p.next()
	value := p.parseExpression()
	if value == nil {
		p.syncToNextLine()
		return nil
	}
	return &ast.InputNode{Value: value, Position: toASTPos(tok)}
}

func (p *Parser) parseInputlnStmt() ast.Statement {
	tok := p.next()
	value := p.parseExpression()
	if value == nil {
		p.syncToNextLine()
		return nil
	}
	return &ast.InputlnNode{Value: value, Position: toASTPos(tok)}
}

func (p *Parser) parseImportStmt() ast.Statement {
	tok := p.next()
	id, ok := p.expectIdent()
	if !ok {
		p.syncToNextLine()
		return nil
	}
	return &ast.ImportNode{Name: id.Literal, Position: toASTPos(tok)}
}

func (p *Parser) parseExportStmt() ast.Statement {
	tok := p.next()
	id, ok := p.expectIdent()
	if !ok {
		p.syncToNextLine()
		return nil
	}
	return &ast.ExportNode{Name: id.Literal, Position: toASTPos(tok)}
}

func (p *Parser) parseDelStmt() ast.Statement {
	tok := p.next()
	id, ok := p.expectIdent()
	if !ok {
		p.syncToNextLine()
		return nil
	}
	return &ast.DelNode{Name: id.Literal, Position: toASTPos(tok)}
}

func (p *Parser) parseScopeStmt() ast.Statement {
	kw := p.next()
	id, ok := p.expectIdent()
	if !ok {
		p.syncToNextLine()
		return nil
	}
	return &ast.ScopeNode{Keyword: kw.Literal, Name: id.Literal, Position: toASTPos(kw)}
}

func (p *Parser) parseRaiseStmt() ast.Statement {
	tok := p.next()
	id, ok := p.expectIdent()
	if !ok {
		p.syncToNextLine()
		return nil
	}

	var msg ast.Expression
	if p.at(lexer.TOKEN_LPAREN) {
		p.next()
		if !p.at(lexer.TOKEN_RPAREN) {
			msg = p.parseExpression()
			if msg == nil {
				p.syncToNextLine()
				return nil
			}
		}
		if !p.expect(lexer.TOKEN_RPAREN) {
			p.syncToNextLine()
			return nil
		}
	}

	return &ast.RaiseNode{ErrorName: id.Literal, Message: msg, Position: toASTPos(tok)}
}

func (p *Parser) parseTryStmt() ast.Statement {
	tok := p.next()
	if !p.expect(lexer.TOKEN_LBRACE) {
		p.syncToNextLine()
		return nil
	}
	body := p.parseBlockBody()
	if body == nil {
		return nil
	}

	var exc *ast.ExceptClause
	p.skipIgnorable()
	if p.at(lexer.TOKEN_EXCEPT) {
		exc = p.parseExceptClause()
		if exc == nil {
			return nil
		}
	}

	var fin *ast.FinallyClause
	p.skipIgnorable()
	if p.at(lexer.TOKEN_FINALLY) {
		fin = p.parseFinallyClause()
		if fin == nil {
			return nil
		}
	}

	return &ast.TryNode{Body: body, Except: exc, Finally: fin, Position: toASTPos(tok)}
}

func (p *Parser) parseExceptClause() *ast.ExceptClause {
	tok := p.next()
	errVar := ""
	if p.at(lexer.TOKEN_LPAREN) {
		p.next()
		id, ok := p.expectIdent()
		if !ok {
			p.syncToNextLine()
			return nil
		}
		errVar = id.Literal
		if !p.expect(lexer.TOKEN_RPAREN) {
			p.syncToNextLine()
			return nil
		}
	}

	if !p.expect(lexer.TOKEN_LBRACE) {
		p.syncToNextLine()
		return nil
	}
	body := p.parseBlockBody()
	if body == nil {
		return nil
	}

	return &ast.ExceptClause{ErrVar: errVar, Body: body, Position: toASTPos(tok)}
}

func (p *Parser) parseFinallyClause() *ast.FinallyClause {
	tok := p.next()
	if !p.expect(lexer.TOKEN_LBRACE) {
		p.syncToNextLine()
		return nil
	}
	body := p.parseBlockBody()
	if body == nil {
		return nil
	}
	return &ast.FinallyClause{Body: body, Position: toASTPos(tok)}
}

func (p *Parser) parseExprStmt() ast.Statement {
	expr := p.parseExpression()
	if expr == nil {
		p.syncToNextLine()
		return nil
	}
	return &ast.ExprStmt{Expr: expr, Position: expr.Pos()}
}

func (p *Parser) parseExpression() ast.Expression {
	depth, ok := p.enterExpr()
	if !ok {
		return nil
	}
	defer p.leaveExpr(depth)
	return p.parseOr()
}

func (p *Parser) parseOr() ast.Expression {
	left := p.parseAnd()
	for p.at(lexer.TOKEN_OR) {
		op := p.next()
		right := p.parseAnd()
		if right == nil {
			return left
		}
		left = &ast.BinaryNode{Left: left, Operator: op.Type, Right: right, Position: toASTPos(op)}
	}
	return left
}

func (p *Parser) parseAnd() ast.Expression {
	left := p.parseNot()
	for p.at(lexer.TOKEN_AND) {
		op := p.next()
		right := p.parseNot()
		if right == nil {
			return left
		}
		left = &ast.BinaryNode{Left: left, Operator: op.Type, Right: right, Position: toASTPos(op)}
	}
	return left
}

func (p *Parser) parseNot() ast.Expression {
	if p.at(lexer.TOKEN_NOT) {
		tok := p.next()
		op := p.parseNot()
		if op == nil {
			return nil
		}
		return &ast.UnaryNode{Operator: tok.Type, Operand: op, Position: toASTPos(tok)}
	}
	return p.parseIs()
}

func (p *Parser) parseIs() ast.Expression {
	if p.at(lexer.TOKEN_IS) {
		tok := p.next()
		op := p.parseIs()
		if op == nil {
			return nil
		}
		return &ast.UnaryNode{Operator: tok.Type, Operand: op, Position: toASTPos(tok)}
	}
	return p.parseComparison()
}

func (p *Parser) parseComparison() ast.Expression {
	left := p.parseTerm()
	for p.at(lexer.TOKEN_EQ) || p.at(lexer.TOKEN_NEQ) || p.at(lexer.TOKEN_LT) || p.at(lexer.TOKEN_GT) || p.at(lexer.TOKEN_LTE) || p.at(lexer.TOKEN_GTE) {
		op := p.next()
		right := p.parseTerm()
		if right == nil {
			return left
		}
		left = &ast.BinaryNode{Left: left, Operator: op.Type, Right: right, Position: toASTPos(op)}
	}
	return left
}

func (p *Parser) parseTerm() ast.Expression {
	left := p.parseFactor()
	for p.at(lexer.TOKEN_PLUS) || p.at(lexer.TOKEN_MINUS) {
		op := p.next()
		right := p.parseFactor()
		if right == nil {
			return left
		}
		left = &ast.BinaryNode{Left: left, Operator: op.Type, Right: right, Position: toASTPos(op)}
	}
	return left
}

func (p *Parser) parseFactor() ast.Expression {
	left := p.parseUnary()
	for p.at(lexer.TOKEN_STAR) || p.at(lexer.TOKEN_SLASH) || p.at(lexer.TOKEN_PERCENT) || p.at(lexer.TOKEN_STARSTAR) {
		op := p.next()
		right := p.parseUnary()
		if right == nil {
			return left
		}
		left = &ast.BinaryNode{Left: left, Operator: op.Type, Right: right, Position: toASTPos(op)}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expression {
	if p.at(lexer.TOKEN_MINUS) {
		depth, ok := p.enterExpr()
		if !ok {
			return nil
		}
		defer p.leaveExpr(depth)
		tok := p.next()
		op := p.parseUnary()
		if op == nil {
			return nil
		}
		return &ast.UnaryNode{Operator: tok.Type, Operand: op, Position: toASTPos(tok)}
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() ast.Expression {
	tok := p.cur()
	switch tok.Type {
	case lexer.TOKEN_NUMBER:
		p.next()
		v, _ := strconv.ParseFloat(tok.Literal, 64)
		return p.parsePostfix(&ast.NumberLiteral{Value: v, Position: toASTPos(tok)})
	case lexer.TOKEN_STRING:
		p.next()
		return p.parsePostfix(&ast.StringLiteral{Value: tok.Literal, Raw: false, Position: toASTPos(tok)})
	case lexer.TOKEN_RAW_STRING:
		p.next()
		return p.parsePostfix(&ast.StringLiteral{Value: tok.Literal, Raw: true, Position: toASTPos(tok)})
	case lexer.TOKEN_TRUE:
		p.next()
		return p.parsePostfix(&ast.BoolLiteral{Value: true, Position: toASTPos(tok)})
	case lexer.TOKEN_FALSE:
		p.next()
		return p.parsePostfix(&ast.BoolLiteral{Value: false, Position: toASTPos(tok)})
	case lexer.TOKEN_NULL:
		p.next()
		return p.parsePostfix(&ast.NullLiteral{Position: toASTPos(tok)})
	case lexer.TOKEN_IDENT:
		p.next()
		if p.at(lexer.TOKEN_DOT) && isQualifiedCallAhead(p) {
			return p.parsePostfix(p.parseQualifiedCallFromIdent(tok.Literal, toASTPos(tok)))
		}
		return p.parsePostfix(&ast.IdentNode{Name: tok.Literal, Position: toASTPos(tok)})
	case lexer.TOKEN_LPAREN:
		p.next()
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}
		if !p.expect(lexer.TOKEN_RPAREN) {
			return nil
		}
		return p.parsePostfix(expr)
	case lexer.TOKEN_LBRACKET:
		return p.parsePostfix(p.parseArrayLiteral())
	case lexer.TOKEN_DEFINE:
		return p.parsePostfix(p.parseDefineCallExpr())
	case lexer.TOKEN_PRINT:
		return p.parsePostfix(p.parsePrintExpr())
	case lexer.TOKEN_PRINTLN:
		return p.parsePostfix(p.parsePrintlnExpr())
	case lexer.TOKEN_CALL:
		return p.parsePostfix(p.parseFuncRefExpr())
	default:
		if tok.Type == lexer.TOKEN_ILLEGAL {
			p.addIllegalTokenError(tok)
			p.next()
			return nil
		}
		p.addUnexpectedToken(tok)
		return nil
	}
}

// parsePostfix parses indexing suffixes ([...]) applied to a primary
// expression: arr[0], m[i][j], define f()[0], ...
func (p *Parser) parsePostfix(expr ast.Expression) ast.Expression {
	if expr == nil {
		return nil
	}
	for p.at(lexer.TOKEN_LBRACKET) {
		openTok := p.next()
		idx := p.parseExpression()
		if idx == nil {
			return nil
		}
		if !p.expect(lexer.TOKEN_RBRACKET) {
			return nil
		}
		expr = &ast.IndexNode{Collection: expr, Index: idx, Position: toASTPos(openTok)}
	}
	return expr
}

// parseFuncRefExpr parses `call name` or `call mod.name` used as an
// expression: a first-class function reference. Parentheses after the name
// are a definition, not part of the reference.
func (p *Parser) parseFuncRefExpr() ast.Expression {
	tok := p.next()
	name, ok := p.parseQualifiedIdent()
	if !ok {
		return nil
	}
	return &ast.FuncRefNode{Name: name, Position: toASTPos(tok)}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	tok := p.next()
	elements := make([]ast.Expression, 0)
	if !p.at(lexer.TOKEN_RBRACKET) {
		for {
			expr := p.parseExpression()
			if expr == nil {
				return nil
			}
			elements = append(elements, expr)
			if p.at(lexer.TOKEN_COMMA) {
				p.next()
				continue
			}
			break
		}
	}
	if !p.expect(lexer.TOKEN_RBRACKET) {
		return nil
	}
	return &ast.ArrayLiteral{Elements: elements, Position: toASTPos(tok)}
}

func (p *Parser) parseDefineCallExpr() ast.Expression {
	tok := p.next()
	name, ok := p.parseQualifiedIdent()
	if !ok || !p.expect(lexer.TOKEN_LPAREN) {
		return nil
	}

	args := make([]ast.Expression, 0)
	if !p.at(lexer.TOKEN_RPAREN) {
		for {
			a := p.parseExpression()
			if a == nil {
				return nil
			}
			args = append(args, a)
			if p.at(lexer.TOKEN_COMMA) {
				p.next()
				continue
			}
			break
		}
	}

	if !p.expect(lexer.TOKEN_RPAREN) {
		return nil
	}

	return &ast.FuncCallNode{Name: name, Args: args, Position: toASTPos(tok)}
}

func (p *Parser) parseQualifiedIdent() (string, bool) {
	first, ok := p.expectIdent()
	if !ok {
		return "", false
	}
	name := first.Literal
	for p.at(lexer.TOKEN_DOT) {
		p.next()
		next, ok := p.expectIdent()
		if !ok {
			return "", false
		}
		name += "." + next.Literal
	}
	return name, true
}

func (p *Parser) parsePrintExpr() *ast.PrintNode {
	tok := p.next()
	if p.at(lexer.TOKEN_NEWLINE) || p.at(lexer.TOKEN_EOF) || p.at(lexer.TOKEN_RBRACE) {
		return &ast.PrintNode{Prompt: nil, Position: toASTPos(tok)}
	}
	prompt := p.parseExpression()
	if prompt == nil {
		return nil
	}
	return &ast.PrintNode{Prompt: prompt, Position: toASTPos(tok)}
}

func (p *Parser) parsePrintlnExpr() *ast.PrintlnNode {
	tok := p.next()
	if p.at(lexer.TOKEN_NEWLINE) || p.at(lexer.TOKEN_EOF) || p.at(lexer.TOKEN_RBRACE) {
		return &ast.PrintlnNode{Prompt: nil, Position: toASTPos(tok)}
	}
	prompt := p.parseExpression()
	if prompt == nil {
		return nil
	}
	return &ast.PrintlnNode{Prompt: prompt, Position: toASTPos(tok)}
}

// isQualifiedCallAhead reports whether the token stream after an identifier
// and a '.' forms a module-qualified call (e.g. wronglib.sort(...)).
func isQualifiedCallAhead(p *Parser) bool {
	saved := p.pos
	defer func() { p.pos = saved }()

	p.next() // consume '.'
	if !p.at(lexer.TOKEN_IDENT) {
		return false
	}
	p.next()
	return p.at(lexer.TOKEN_LPAREN)
}

// parseQualifiedCallFromIdent parses the remainder of a module-qualified
// call (".name(args)") after the leading identifier has been consumed.
func (p *Parser) parseQualifiedCallFromIdent(first string, pos ast.Position) ast.Expression {
	name := first
	for p.at(lexer.TOKEN_DOT) {
		p.next()
		next, ok := p.expectIdent()
		if !ok {
			return nil
		}
		name += "." + next.Literal
	}
	if !p.expect(lexer.TOKEN_LPAREN) {
		return nil
	}
	args := make([]ast.Expression, 0)
	if !p.at(lexer.TOKEN_RPAREN) {
		for {
			a := p.parseExpression()
			if a == nil {
				return nil
			}
			args = append(args, a)
			if p.at(lexer.TOKEN_COMMA) {
				p.next()
				continue
			}
			break
		}
	}
	if !p.expect(lexer.TOKEN_RPAREN) {
		return nil
	}
	return &ast.FuncCallNode{Name: name, Args: args, Position: pos}
}

func (p *Parser) parseBlockBody() *ast.BlockNode {
	openTok := p.prev()
	body := &ast.BlockNode{Position: toASTPos(openTok)}
	p.blockDepth++

	for !p.at(lexer.TOKEN_EOF) {
		p.skipIgnorable()
		if p.at(lexer.TOKEN_RBRACE) {
			p.next()
			p.blockDepth--
			return body
		}
		if p.at(lexer.TOKEN_ILLEGAL) {
			p.addIllegalTokenError(p.cur())
			p.next()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			body.Statements = append(body.Statements, stmt)
		}
		if p.at(lexer.TOKEN_NEWLINE) {
			p.next()
		}
	}

	p.blockDepth--
	err := diagnostics.New(diagnostics.SyntaxError, p.tokenPos(openTok))
	err.Found = "<eof>"
	err.Detail = "unclosed block — missing closing '{'"
	err.Hint = "add a '{' at the end of this block to close it"
	err.Expected = []string{"{"}
	p.errors = append(p.errors, err)
	return nil
}

func (p *Parser) skipIgnorable() {
	for {
		t := p.cur().Type
		if t == lexer.TOKEN_NEWLINE || t == lexer.TOKEN_COMMENT_CONTENT || t == lexer.TOKEN_SLASHSLASH || t == lexer.TOKEN_BANGBANG || t == lexer.TOKEN_BLOCK_COMMENT_OPEN || t == lexer.TOKEN_BLOCK_COMMENT_CLOSE || t == lexer.TOKEN_WBLOCK_COMMENT_OPEN || t == lexer.TOKEN_WBLOCK_COMMENT_CLOSE {
			p.next()
			continue
		}
		break
	}
}

func (p *Parser) syncToNextLine() {
	for !p.at(lexer.TOKEN_EOF) && !p.at(lexer.TOKEN_NEWLINE) {
		p.next()
	}
	if p.at(lexer.TOKEN_NEWLINE) {
		p.next()
	}
}

func (p *Parser) addUnexpectedToken(tok lexer.Token) {
	err := diagnostics.NewUnexpectedToken(p.tokenPos(tok), p.foundToken(tok))
	p.errors = append(p.errors, err)
}

func (p *Parser) addExpectedToken(tok lexer.Token, expected ...lexer.TokenType) {
	ex := make([]string, 0, len(expected))
	for _, t := range expected {
		ex = append(ex, tokenLabel(t))
	}
	err := diagnostics.NewExpectedToken(p.tokenPos(tok), ex, p.foundToken(tok))
	p.errors = append(p.errors, err)
}

func (p *Parser) addExpectedIdentifier(tok lexer.Token) {
	err := diagnostics.NewExpectedToken(p.tokenPos(tok), []string{"identifier"}, p.foundToken(tok))
	p.errors = append(p.errors, err)
}

func (p *Parser) addIllegalTokenError(tok lexer.Token) {
	pos := p.tokenPos(tok)
	var err *diagnostics.WorngError
	switch tok.Literal {
	case `"`, `'`, "~":
		err = diagnostics.NewUnterminatedString(pos)
	case "/*", "!*":
		err = diagnostics.NewUnterminatedBlockComment(pos, tok.Literal)
	default:
		err = diagnostics.NewIllegalToken(pos, p.foundToken(tok))
	}
	p.errors = append(p.errors, err)
}

func (p *Parser) expect(t lexer.TokenType) bool {
	if p.at(t) {
		p.next()
		return true
	}
	p.addExpectedToken(p.cur(), t)
	return false
}

func (p *Parser) expectIdent() (lexer.Token, bool) {
	if p.at(lexer.TOKEN_IDENT) {
		return p.next(), true
	}
	p.addExpectedIdentifier(p.cur())
	return lexer.Token{}, false
}

func (p *Parser) tokenPos(tok lexer.Token) diagnostics.Position {
	width := utf8.RuneCountInString(tok.Literal)
	if width <= 0 {
		width = 1
	}
	endCol := tok.Column + width - 1
	if endCol < tok.Column {
		endCol = tok.Column
	}
	return diagnostics.Position{
		File:      p.sourceFile,
		Line:      tok.Line,
		Column:    tok.Column,
		EndLine:   tok.Line,
		EndColumn: endCol,
	}
}

func (p *Parser) foundToken(tok lexer.Token) string {
	if tok.Type == lexer.TOKEN_EOF {
		return "<eof>"
	}
	if tok.Literal != "" {
		return tok.Literal
	}
	return "token"
}

func quoteDiagToken(tok string) string {
	if strings.TrimSpace(tok) == "" {
		return "<eof>"
	}
	return strconv.Quote(tok)
}

func tokenLabel(t lexer.TokenType) string {
	switch t {
	case lexer.TOKEN_IDENT:
		return "identifier"
	case lexer.TOKEN_NUMBER:
		return "number"
	case lexer.TOKEN_STRING:
		return "string"
	case lexer.TOKEN_RAW_STRING:
		return "raw string"
	case lexer.TOKEN_ASSIGN:
		return "="
	case lexer.TOKEN_LPAREN:
		return "("
	case lexer.TOKEN_RPAREN:
		return ")"
	case lexer.TOKEN_LBRACE:
		return "}"
	case lexer.TOKEN_RBRACE:
		return "{"
	case lexer.TOKEN_LBRACKET:
		return "["
	case lexer.TOKEN_RBRACKET:
		return "]"
	case lexer.TOKEN_COMMA:
		return ","
	case lexer.TOKEN_DOT:
		return "."
	case lexer.TOKEN_IF:
		return "if"
	case lexer.TOKEN_ELSE:
		return "else"
	case lexer.TOKEN_WHILE:
		return "while"
	case lexer.TOKEN_FOR:
		return "for"
	case lexer.TOKEN_IN:
		return "in"
	case lexer.TOKEN_CALL:
		return "call"
	case lexer.TOKEN_DEFINE:
		return "define"
	case lexer.TOKEN_RETURN:
		return "return"
	case lexer.TOKEN_DISCARD:
		return "discard"
	case lexer.TOKEN_INPUT:
		return "input"
	case lexer.TOKEN_INPUTLN:
		return "inputln"
	case lexer.TOKEN_PRINT:
		return "print"
	case lexer.TOKEN_PRINTLN:
		return "println"
	case lexer.TOKEN_EOF:
		return "<eof>"
	default:
		return "token"
	}
}

func (p *Parser) at(t lexer.TokenType) bool {
	return p.cur().Type == t
}

func (p *Parser) cur() lexer.Token {
	if len(p.tokens) == 0 {
		return lexer.Token{Type: lexer.TOKEN_EOF, Line: 1, Column: 1}
	}
	if p.pos < 0 {
		return p.tokens[0]
	}
	if p.pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.pos]
}

func (p *Parser) peek() lexer.Token {
	if len(p.tokens) == 0 {
		return lexer.Token{Type: lexer.TOKEN_EOF, Line: 1, Column: 1}
	}
	if p.pos+1 >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) prev() lexer.Token {
	if len(p.tokens) == 0 {
		return lexer.Token{Type: lexer.TOKEN_EOF, Line: 1, Column: 1}
	}
	if p.pos-1 < 0 {
		return p.tokens[0]
	}
	return p.tokens[p.pos-1]
}

func (p *Parser) next() lexer.Token {
	tok := p.cur()
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
	return tok
}

func toASTPos(tok lexer.Token) ast.Position {
	return ast.Position{Line: tok.Line, Column: tok.Column}
}
