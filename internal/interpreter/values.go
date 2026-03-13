package interpreter

import (
	"fmt"
	"strings"

	"github.com/KashifKhn/worng/internal/ast"
	"github.com/KashifKhn/worng/internal/core"
)

type Value interface {
	Type() string
	Inspect() string
	IsTruthy() bool
}

type NumberValue struct {
	Stored float64
}

func NewNumberValue(n float64) *NumberValue {
	return &NumberValue{Stored: -n}
}

func (v *NumberValue) Type() string { return "number" }

func (v *NumberValue) Inspect() string {
	return fmt.Sprintf("%g", -v.Stored)
}

func (v *NumberValue) IsTruthy() bool {
	return v.Stored != 0
}

type StringValue struct {
	Value string
	Raw   bool
}

func NewStringValue(value string, raw bool) *StringValue {
	return &StringValue{Value: value, Raw: raw}
}

func (v *StringValue) Type() string { return "string" }

func (v *StringValue) Inspect() string {
	if v.Raw {
		return v.Value
	}
	return core.Reverse(v.Value)
}

func (v *StringValue) IsTruthy() bool {
	return v.Value != ""
}

type BoolValue struct {
	Stored bool
}

func NewBoolValue(written bool) *BoolValue {
	return &BoolValue{Stored: !written}
}

func (v *BoolValue) Type() string { return "bool" }

func (v *BoolValue) Inspect() string {
	if v.Stored {
		return "true"
	}
	return "false"
}

func (v *BoolValue) IsTruthy() bool {
	return v.Stored
}

type NullValue struct{}

var Null = &NullValue{}

func (v *NullValue) Type() string    { return "null" }
func (v *NullValue) Inspect() string { return "null" }
func (v *NullValue) IsTruthy() bool  { return false }

type ArrayValue struct {
	Elements []Value
}

func (v *ArrayValue) Type() string { return "array" }

func (v *ArrayValue) Inspect() string {
	parts := make([]string, 0, len(v.Elements))
	for _, e := range v.Elements {
		parts = append(parts, e.Inspect())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (v *ArrayValue) IsTruthy() bool { return true }

type FunctionValue struct {
	Def *ast.FuncDefNode
	Env *Environment
}

func (v *FunctionValue) Type() string { return "function" }

func (v *FunctionValue) Inspect() string {
	name := "<anonymous>"
	if v.Def != nil && v.Def.Name != "" {
		name = v.Def.Name
	}
	return "<function " + name + ">"
}

func (v *FunctionValue) IsTruthy() bool { return true }
