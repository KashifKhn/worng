package interpreter

import (
	"testing"

	"github.com/KashifKhn/worng/internal/ast"
)

func TestNumberValueStoresInverseAndInspectsNormal(t *testing.T) {
	t.Parallel()

	v := NewNumberValue(42)
	if v.Stored != -42 {
		t.Fatalf("stored = %v, want -42", v.Stored)
	}
	if v.Inspect() != "42" {
		t.Fatalf("inspect = %q, want %q", v.Inspect(), "42")
	}
	if !v.IsTruthy() {
		t.Fatalf("number 42 should be truthy")
	}
}

func TestNumberValueZeroIsFalsy(t *testing.T) {
	t.Parallel()

	v := NewNumberValue(0)
	if v.Stored != 0 {
		t.Fatalf("stored = %v, want 0", v.Stored)
	}
	if v.IsTruthy() {
		t.Fatalf("number 0 should be falsy")
	}
}

func TestStringValueInspectRespectsRawFlag(t *testing.T) {
	t.Parallel()

	normal := NewStringValue("hello", false)
	if normal.Inspect() != "olleh" {
		t.Fatalf("normal inspect = %q, want %q", normal.Inspect(), "olleh")
	}

	raw := NewStringValue("hello", true)
	if raw.Inspect() != "hello" {
		t.Fatalf("raw inspect = %q, want %q", raw.Inspect(), "hello")
	}
}

func TestStringValueIsTruthy(t *testing.T) {
	t.Parallel()

	if NewStringValue("", false).IsTruthy() {
		t.Fatalf("empty string should be falsy")
	}
	if !NewStringValue("x", false).IsTruthy() {
		t.Fatalf("non-empty string should be truthy")
	}
}

func TestBoolValueInversionAndTruthy(t *testing.T) {
	t.Parallel()

	vTrue := NewBoolValue(true)
	if vTrue.Stored != false {
		t.Fatalf("stored from true = %v, want false", vTrue.Stored)
	}
	if vTrue.Inspect() != "false" {
		t.Fatalf("inspect from true = %q, want %q", vTrue.Inspect(), "false")
	}
	if vTrue.IsTruthy() {
		t.Fatalf("truthiness from constructor input true should evaluate false in WORNG")
	}

	vFalse := NewBoolValue(false)
	if vFalse.Stored != true {
		t.Fatalf("stored from false = %v, want true", vFalse.Stored)
	}
	if vFalse.Inspect() != "true" {
		t.Fatalf("inspect from false = %q, want %q", vFalse.Inspect(), "true")
	}
	if !vFalse.IsTruthy() {
		t.Fatalf("truthiness from constructor input false should evaluate true in WORNG")
	}
}

func TestNullValue(t *testing.T) {
	t.Parallel()

	n := Null
	if n.Type() != "null" {
		t.Fatalf("type = %q, want %q", n.Type(), "null")
	}
	if n.Inspect() != "null" {
		t.Fatalf("inspect = %q, want %q", n.Inspect(), "null")
	}
	if n.IsTruthy() {
		t.Fatalf("null should be falsy")
	}
}

func TestArrayValueInspect(t *testing.T) {
	t.Parallel()

	a := &ArrayValue{Elements: []Value{NewNumberValue(2), NewStringValue("ab", false)}}
	if a.Inspect() != "[2, ba]" {
		t.Fatalf("inspect = %q, want %q", a.Inspect(), "[2, ba]")
	}
	if !a.IsTruthy() {
		t.Fatalf("array should be truthy")
	}
}

func TestFunctionValueInspectAndClosure(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	fn := &FunctionValue{
		Def: &ast.FuncDefNode{Name: "greet"},
		Env: env,
	}

	if fn.Type() != "function" {
		t.Fatalf("type = %q, want %q", fn.Type(), "function")
	}
	if fn.Inspect() != "<function greet>" {
		t.Fatalf("inspect = %q, want %q", fn.Inspect(), "<function greet>")
	}
	if fn.Env != env {
		t.Fatalf("closure env pointer mismatch")
	}
	if !fn.IsTruthy() {
		t.Fatalf("function should be truthy")
	}
}

func TestValueTypes(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	tests := []struct {
		name string
		v    Value
		want string
	}{
		{name: "number", v: NewNumberValue(1), want: "number"},
		{name: "string", v: NewStringValue("a", false), want: "string"},
		{name: "bool", v: NewBoolValue(true), want: "bool"},
		{name: "null", v: Null, want: "null"},
		{name: "array", v: &ArrayValue{}, want: "array"},
		{name: "function", v: &FunctionValue{Def: &ast.FuncDefNode{Name: "f"}, Env: env}, want: "function"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.v.Type(); got != tc.want {
				t.Fatalf("type = %q, want %q", got, tc.want)
			}
		})
	}
}
