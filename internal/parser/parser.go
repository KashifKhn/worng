// Package parser implements a hand-written recursive descent parser for WORNG.
// It consumes a []lexer.Token and produces a *ast.ProgramNode.
//
// The parser is tolerant: it never panics and always returns a (partial) AST
// even when syntax errors are encountered.
package parser

import (
	"strconv"

	"github.com/KashifKhn/worng/internal/ast"
	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/lexer"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
	errors []error
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Parse() (*ast.ProgramNode, []error) {
	program := &ast.ProgramNode{Position: ast.Position{Line: 1, Column: 1}}

	for !p.at(lexer.TOKEN_EOF) {
		p.skipIgnorable()
		if p.at(lexer.TOKEN_EOF) {
			break
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
	case lexer.TOKEN_PRINT:
		n := p.parsePrintExpr()
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
		p.addSyntaxError(tok)
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
		p.addSyntaxError(tok)
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
		return &ast.NumberLiteral{Value: v, Position: toASTPos(tok)}
	case lexer.TOKEN_STRING:
		p.next()
		return &ast.StringLiteral{Value: tok.Literal, Raw: false, Position: toASTPos(tok)}
	case lexer.TOKEN_RAW_STRING:
		p.next()
		return &ast.StringLiteral{Value: tok.Literal, Raw: true, Position: toASTPos(tok)}
	case lexer.TOKEN_TRUE:
		p.next()
		return &ast.BoolLiteral{Value: true, Position: toASTPos(tok)}
	case lexer.TOKEN_FALSE:
		p.next()
		return &ast.BoolLiteral{Value: false, Position: toASTPos(tok)}
	case lexer.TOKEN_NULL:
		p.next()
		return &ast.NullLiteral{Position: toASTPos(tok)}
	case lexer.TOKEN_IDENT:
		p.next()
		return &ast.IdentNode{Name: tok.Literal, Position: toASTPos(tok)}
	case lexer.TOKEN_LPAREN:
		p.next()
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}
		if !p.expect(lexer.TOKEN_RPAREN) {
			return nil
		}
		return expr
	case lexer.TOKEN_LBRACKET:
		return p.parseArrayLiteral()
	case lexer.TOKEN_DEFINE:
		return p.parseDefineCallExpr()
	case lexer.TOKEN_PRINT:
		return p.parsePrintExpr()
	default:
		p.addSyntaxError(tok)
		return nil
	}
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
		p.addSyntaxError(tok)
		return nil
	}
	return &ast.PrintNode{Prompt: prompt, Position: toASTPos(tok)}
}

func (p *Parser) parseBlockBody() *ast.BlockNode {
	openTok := p.prev()
	body := &ast.BlockNode{Position: toASTPos(openTok)}

	for !p.at(lexer.TOKEN_EOF) {
		p.skipIgnorable()
		if p.at(lexer.TOKEN_RBRACE) {
			p.next()
			return body
		}

		stmt := p.parseStatement()
		if stmt != nil {
			body.Statements = append(body.Statements, stmt)
		}
		if p.at(lexer.TOKEN_NEWLINE) {
			p.next()
		}
	}

	p.addSyntaxError(openTok)
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

func (p *Parser) addSyntaxError(tok lexer.Token) {
	err := diagnostics.New(diagnostics.SyntaxError, diagnostics.Position{Line: tok.Line, Column: tok.Column})
	p.errors = append(p.errors, err)
}

func (p *Parser) expect(t lexer.TokenType) bool {
	if p.at(t) {
		p.next()
		return true
	}
	p.addSyntaxError(p.cur())
	return false
}

func (p *Parser) expectIdent() (lexer.Token, bool) {
	if p.at(lexer.TOKEN_IDENT) {
		return p.next(), true
	}
	p.addSyntaxError(p.cur())
	return lexer.Token{}, false
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
