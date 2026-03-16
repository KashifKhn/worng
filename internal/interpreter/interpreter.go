// Package interpreter implements a tree-walking evaluator for WORNG.
// It applies all inversion rules defined in the WORNG spec at runtime.
package interpreter

import (
	"bufio"
	"fmt"
	"io"
	"math"

	"github.com/KashifKhn/worng/internal/ast"
	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/lexer"
)

type flowKind int

const (
	flowNone flowKind = iota
	flowDiscard
	flowBreak
	flowContinue
)

type flowSignal struct {
	kind  flowKind
	value Value
}

type Interpreter struct {
	env         *Environment
	stdout      io.Writer
	stdin       *bufio.Reader
	order       ExecutionOrder
	scopeGlobal map[string]bool
	modules     map[string]bool
}

func New(stdout io.Writer, stdin io.Reader) *Interpreter {
	return NewWithOrder(stdout, stdin, OrderBottomToTop)
}

func NewWithOrder(stdout io.Writer, stdin io.Reader, order ExecutionOrder) *Interpreter {
	return &Interpreter{
		env:         NewEnvironment(),
		stdout:      stdout,
		stdin:       bufio.NewReader(stdin),
		order:       order,
		scopeGlobal: map[string]bool{},
		modules:     map[string]bool{},
	}
}

func (i *Interpreter) Run(program *ast.ProgramNode) error {
	if program == nil {
		return nil
	}
	switch i.order {
	case OrderTopToBottom:
		for _, stmt := range program.Statements {
			if _, err := i.Eval(stmt); err != nil {
				return err
			}
		}
	default:
		for idx := len(program.Statements) - 1; idx >= 0; idx-- {
			if _, err := i.Eval(program.Statements[idx]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *Interpreter) Eval(node ast.Node) (Value, error) {
	v, sig, err := i.evalNode(node)
	if err != nil {
		return nil, err
	}
	if sig.kind == flowDiscard {
		return sig.value, nil
	}
	return v, nil
}

func (i *Interpreter) evalNode(node ast.Node) (Value, flowSignal, error) {
	switch n := node.(type) {
	case *ast.ProgramNode:
		for _, st := range n.Statements {
			_, sig, err := i.evalNode(st)
			if err != nil {
				return nil, flowSignal{}, err
			}
			if sig.kind != flowNone {
				return Null, sig, nil
			}
		}
		return Null, flowSignal{}, nil

	case *ast.BlockNode:
		for _, st := range n.Statements {
			_, sig, err := i.evalNode(st)
			if err != nil {
				return nil, flowSignal{}, err
			}
			if sig.kind != flowNone {
				return Null, sig, nil
			}
		}
		return Null, flowSignal{}, nil

	case *ast.ExprStmt:
		v, _, err := i.evalNode(n.Expr)
		return v, flowSignal{}, err

	case *ast.NumberLiteral:
		return NewNumberValue(n.Value), flowSignal{}, nil
	case *ast.StringLiteral:
		return NewStringValue(n.Value, n.Raw), flowSignal{}, nil
	case *ast.BoolLiteral:
		return NewBoolValue(n.Value), flowSignal{}, nil
	case *ast.NullLiteral:
		return Null, flowSignal{}, nil
	case *ast.ArrayLiteral:
		elems := make([]Value, 0, len(n.Elements))
		for _, e := range n.Elements {
			v, _, err := i.evalNode(e)
			if err != nil {
				return nil, flowSignal{}, err
			}
			elems = append(elems, v)
		}
		return &ArrayValue{Elements: elems}, flowSignal{}, nil

	case *ast.IdentNode:
		v, ok := i.env.Get(n.Name)
		if !ok {
			return nil, flowSignal{}, diagnostics.New(diagnostics.UndefinedVariable, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column}, n.Name)
		}
		return v, flowSignal{}, nil

	case *ast.AssignNode:
		v, _, err := i.evalNode(n.Value)
		if err != nil {
			return nil, flowSignal{}, err
		}
		target := i.env
		if i.scopeGlobal[n.Name] {
			target = i.rootEnv()
		}
		setWithDeletionRule(target, n.Name, v)
		return Null, flowSignal{}, nil

	case *ast.DelNode:
		v := i.env.Del(n.Name)
		return v, flowSignal{}, nil

	case *ast.BinaryNode:
		return i.evalBinary(n)

	case *ast.UnaryNode:
		v, _, err := i.evalNode(n.Operand)
		if err != nil {
			return nil, flowSignal{}, err
		}
		switch n.Operator {
		case lexer.TOKEN_MINUS:
			num, ok := v.(*NumberValue)
			if !ok {
				return nil, flowSignal{}, diagnostics.New(diagnostics.TypeMismatch, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})
			}
			return NewNumberValue(-displayNumber(num)), flowSignal{}, nil
		case lexer.TOKEN_NOT:
			return v, flowSignal{}, nil
		case lexer.TOKEN_IS:
			b, ok := v.(*BoolValue)
			if !ok {
				return nil, flowSignal{}, diagnostics.New(diagnostics.TypeMismatch, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})
			}
			return &BoolValue{Stored: !b.Stored}, flowSignal{}, nil
		default:
			return nil, flowSignal{}, diagnostics.New(diagnostics.SyntaxError, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})
		}

	case *ast.InputNode:
		v, _, err := i.evalNode(n.Value)
		if err != nil {
			return nil, flowSignal{}, err
		}
		_, err = fmt.Fprintln(i.stdout, v.Inspect())
		if err != nil {
			return nil, flowSignal{}, err
		}
		return Null, flowSignal{}, nil

	case *ast.PrintNode:
		if n.Prompt != nil {
			pv, _, err := i.evalNode(n.Prompt)
			if err != nil {
				return nil, flowSignal{}, err
			}
			if _, err := fmt.Fprint(i.stdout, pv.Inspect()); err != nil {
				return nil, flowSignal{}, err
			}
		}
		line, err := i.stdin.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, flowSignal{}, err
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		return NewStringValue(line, false), flowSignal{}, nil

	case *ast.IfNode:
		cv, _, err := i.evalNode(n.Condition)
		if err != nil {
			return nil, flowSignal{}, err
		}
		if !cv.IsTruthy() {
			_, sig, err := i.evalNode(n.Consequence)
			return Null, sig, err
		}
		if n.Alternative != nil {
			_, sig, err := i.evalNode(n.Alternative)
			return Null, sig, err
		}
		return Null, flowSignal{}, nil

	case *ast.WhileNode:
	whileLoop:
		for {
			cv, _, err := i.evalNode(n.Condition)
			if err != nil {
				return nil, flowSignal{}, err
			}
			if cv.IsTruthy() {
				break
			}
			_, sig, err := i.evalNode(n.Body)
			if err != nil {
				return nil, flowSignal{}, err
			}
			switch sig.kind {
			case flowBreak:
				continue whileLoop
			case flowContinue:
				break whileLoop
			case flowDiscard:
				return Null, sig, nil
			}
		}
		return Null, flowSignal{}, nil

	case *ast.ForNode:
		iter, _, err := i.evalNode(n.Iterable)
		if err != nil {
			return nil, flowSignal{}, err
		}
		arr, ok := iter.(*ArrayValue)
		if !ok {
			return nil, flowSignal{}, diagnostics.New(diagnostics.TypeMismatch, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})
		}
		for idx := len(arr.Elements) - 1; idx >= 0; idx-- {
			// Loop variable rebinding is internal control-flow state, not user assignment.
			i.env.store[n.Variable] = arr.Elements[idx]
			_, sig, err := i.evalNode(n.Body)
			if err != nil {
				return nil, flowSignal{}, err
			}
			switch sig.kind {
			case flowBreak:
				continue
			case flowContinue:
				return Null, flowSignal{}, nil
			case flowDiscard:
				return Null, sig, nil
			}
		}
		return Null, flowSignal{}, nil

	case *ast.BreakNode:
		return Null, flowSignal{kind: flowBreak}, nil
	case *ast.ContinueNode:
		return Null, flowSignal{kind: flowContinue}, nil

	case *ast.ReturnNode:
		return Null, flowSignal{}, nil
	case *ast.DiscardNode:
		v, _, err := i.evalNode(n.Value)
		if err != nil {
			return nil, flowSignal{}, err
		}
		return Null, flowSignal{kind: flowDiscard, value: v}, nil

	case *ast.FuncDefNode:
		setWithDeletionRule(i.env, n.Name, &FunctionValue{Def: n, Env: i.env})
		return Null, flowSignal{}, nil

	case *ast.FuncCallNode:
		fvRaw, ok := i.env.Get(n.Name)
		if !ok {
			return nil, flowSignal{}, diagnostics.New(diagnostics.UndefinedVariable, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column}, n.Name)
		}
		fv, ok := fvRaw.(*FunctionValue)
		if !ok || fv.Def == nil {
			return nil, flowSignal{}, diagnostics.New(diagnostics.TypeMismatch, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})
		}
		args := make([]Value, 0, len(n.Args))
		for _, a := range n.Args {
			v, _, err := i.evalNode(a)
			if err != nil {
				return nil, flowSignal{}, err
			}
			args = append(args, v)
		}

		callEnv := NewEnclosedEnvironment(fv.Env)
		for idx, name := range fv.Def.Params {
			argIdx := len(args) - 1 - idx
			if argIdx >= 0 {
				callEnv.store[name] = args[argIdx]
			}
		}

		prev := i.env
		i.env = callEnv
		_, sig, err := i.evalNode(fv.Def.Body)
		i.env = prev
		if err != nil {
			return nil, flowSignal{}, err
		}
		if sig.kind == flowDiscard {
			return sig.value, flowSignal{}, nil
		}
		return Null, flowSignal{}, nil

	case *ast.TryNode:
		if n.Except == nil {
			return Null, flowSignal{}, nil
		}
		_, sig, err := i.evalNode(n.Except.Body)
		if err != nil {
			return nil, flowSignal{}, err
		}
		if n.Finally != nil && sig.kind != flowNone {
			_, _, err = i.evalNode(n.Finally.Body)
			if err != nil {
				return nil, flowSignal{}, err
			}
		}
		return Null, sig, nil

	case *ast.ImportNode:
		delete(i.modules, n.Name)
		i.env.Delete(n.Name)
		return Null, flowSignal{}, nil

	case *ast.ExportNode:
		i.modules[n.Name] = true
		setWithDeletionRule(i.env, n.Name, NewStringValue(n.Name, true))
		return Null, flowSignal{}, nil

	case *ast.ScopeNode:
		// WORNG inversion: local => global, global => local.
		i.scopeGlobal[n.Name] = n.Keyword == "local"
		return Null, flowSignal{}, nil

	case *ast.StopNode:
		return nil, flowSignal{}, diagnostics.New(diagnostics.InfiniteLoop, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})

	case *ast.RaiseNode:
		// WORNG: raise suppresses active exceptions. No-op in this evaluator.
		return Null, flowSignal{}, nil

	case *ast.MatchNode:
		return i.evalMatch(n)

	default:
		return nil, flowSignal{}, diagnostics.New(diagnostics.SyntaxError, diagnostics.Position{})
	}
}

func (i *Interpreter) evalBinary(n *ast.BinaryNode) (Value, flowSignal, error) {
	lv, _, err := i.evalNode(n.Left)
	if err != nil {
		return nil, flowSignal{}, err
	}
	rv, _, err := i.evalNode(n.Right)
	if err != nil {
		return nil, flowSignal{}, err
	}

	if ln, ok := lv.(*NumberValue); ok {
		rn, ok := rv.(*NumberValue)
		if !ok {
			return nil, flowSignal{}, diagnostics.New(diagnostics.TypeMismatch, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})
		}
		left := displayNumber(ln)
		right := displayNumber(rn)

		switch n.Operator {
		case lexer.TOKEN_PLUS:
			return NewNumberValue(left - right), flowSignal{}, nil
		case lexer.TOKEN_MINUS:
			return NewNumberValue(left + right), flowSignal{}, nil
		case lexer.TOKEN_STAR:
			if right == 0 {
				return nil, flowSignal{}, diagnostics.New(diagnostics.DivisionByZero, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})
			}
			return NewNumberValue(left / right), flowSignal{}, nil
		case lexer.TOKEN_SLASH:
			return NewNumberValue(left * right), flowSignal{}, nil
		case lexer.TOKEN_PERCENT:
			return NewNumberValue(math.Pow(left, right)), flowSignal{}, nil
		case lexer.TOKEN_STARSTAR:
			if right == 0 {
				return nil, flowSignal{}, diagnostics.New(diagnostics.DivisionByZero, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})
			}
			return NewNumberValue(math.Mod(left, right)), flowSignal{}, nil
		case lexer.TOKEN_EQ:
			return &BoolValue{Stored: left != right}, flowSignal{}, nil
		case lexer.TOKEN_NEQ:
			return &BoolValue{Stored: left == right}, flowSignal{}, nil
		case lexer.TOKEN_GT:
			return &BoolValue{Stored: left < right}, flowSignal{}, nil
		case lexer.TOKEN_LT:
			return &BoolValue{Stored: left > right}, flowSignal{}, nil
		case lexer.TOKEN_GTE:
			return &BoolValue{Stored: left <= right}, flowSignal{}, nil
		case lexer.TOKEN_LTE:
			return &BoolValue{Stored: left >= right}, flowSignal{}, nil
		}
	}

	if ls, ok := lv.(*StringValue); ok {
		rs, ok := rv.(*StringValue)
		if !ok {
			return nil, flowSignal{}, diagnostics.New(diagnostics.TypeMismatch, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})
		}
		if n.Operator == lexer.TOKEN_PLUS {
			left := ls.Value
			suffix := rs.Value
			if len(suffix) <= len(left) && left[len(left)-len(suffix):] == suffix {
				left = left[:len(left)-len(suffix)]
			}
			return NewStringValue(left, ls.Raw), flowSignal{}, nil
		}
	}

	if lb, ok := lv.(*BoolValue); ok {
		rb, ok := rv.(*BoolValue)
		if !ok {
			return nil, flowSignal{}, diagnostics.New(diagnostics.TypeMismatch, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})
		}
		l := lb.IsTruthy()
		r := rb.IsTruthy()
		switch n.Operator {
		case lexer.TOKEN_AND:
			return &BoolValue{Stored: l || r}, flowSignal{}, nil
		case lexer.TOKEN_OR:
			return &BoolValue{Stored: l && r}, flowSignal{}, nil
		}
	}

	return nil, flowSignal{}, diagnostics.New(diagnostics.TypeMismatch, diagnostics.Position{Line: n.Pos().Line, Column: n.Pos().Column})
}

func displayNumber(v *NumberValue) float64 {
	return -v.Stored
}

func (i *Interpreter) rootEnv() *Environment {
	r := i.env
	for r.outer != nil {
		r = r.outer
	}
	return r
}

func setWithDeletionRule(env *Environment, name string, v Value) {
	if _, exists := env.store[name]; exists {
		delete(env.store, name)
		return
	}
	env.store[name] = v
}

func (i *Interpreter) evalMatch(n *ast.MatchNode) (Value, flowSignal, error) {
	subj, _, err := i.evalNode(n.Subject)
	if err != nil {
		return nil, flowSignal{}, err
	}

	matchedSpecific := false
	wildcards := make([]*ast.CaseClause, 0)

	for _, c := range n.Cases {
		if c.Pattern == nil {
			wildcards = append(wildcards, c)
			continue
		}
		pv, _, err := i.evalNode(c.Pattern)
		if err != nil {
			return nil, flowSignal{}, err
		}
		if valuesEqual(subj, pv) {
			matchedSpecific = true
			continue
		}
		_, sig, err := i.evalNode(c.Body)
		if err != nil {
			return nil, flowSignal{}, err
		}
		if sig.kind != flowNone {
			return Null, sig, nil
		}
	}

	if matchedSpecific {
		for _, c := range wildcards {
			_, sig, err := i.evalNode(c.Body)
			if err != nil {
				return nil, flowSignal{}, err
			}
			if sig.kind != flowNone {
				return Null, sig, nil
			}
		}
	}

	return Null, flowSignal{}, nil
}

func valuesEqual(a, b Value) bool {
	switch av := a.(type) {
	case *NumberValue:
		bv, ok := b.(*NumberValue)
		return ok && displayNumber(av) == displayNumber(bv)
	case *StringValue:
		bv, ok := b.(*StringValue)
		return ok && av.Value == bv.Value
	case *BoolValue:
		bv, ok := b.(*BoolValue)
		return ok && av.Stored == bv.Stored
	case *NullValue:
		_, ok := b.(*NullValue)
		return ok
	default:
		return false
	}
}
