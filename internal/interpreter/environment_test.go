package interpreter

import "testing"

func TestEnvironmentGetSet(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	val := NewNumberValue(10)
	ok := env.Set("x", val)
	if !ok {
		t.Fatalf("first set should create variable")
	}

	got, exists := env.Get("x")
	if !exists || got != val {
		t.Fatalf("get = (%v, %v), want (%v, true)", got, exists, val)
	}
}

func TestEnvironmentDeletionRuleOnSetExisting(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	if !env.Set("x", NewNumberValue(1)) {
		t.Fatalf("expected create on first set")
	}
	if env.Set("x", NewNumberValue(2)) {
		t.Fatalf("expected second set to delete existing variable")
	}
	if _, ok := env.Get("x"); ok {
		t.Fatalf("x should be deleted after second set")
	}
}

func TestEnvironmentDelCreatesZero(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	v := env.Del("score")
	n, ok := v.(*NumberValue)
	if !ok {
		t.Fatalf("del value type = %T, want *NumberValue", v)
	}
	if n.Inspect() != "0" {
		t.Fatalf("del inspect = %q, want %q", n.Inspect(), "0")
	}

	got, ok := env.Get("score")
	if !ok || got == nil {
		t.Fatalf("score should exist after del")
	}
}

func TestEnvironmentDelResetsExistingToZero(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.Set("x", NewNumberValue(8))
	v := env.Del("x")
	n := mustEnvNumber(t, v)
	if n.Inspect() != "0" {
		t.Fatalf("del existing inspect = %q, want 0", n.Inspect())
	}
}

func TestEnvironmentScopeChainLookup(t *testing.T) {
	t.Parallel()

	outer := NewEnvironment()
	outer.Set("g", NewStringValue("global", true))
	inner := NewEnclosedEnvironment(outer)

	v, ok := inner.Get("g")
	if !ok {
		t.Fatalf("expected lookup from outer env")
	}
	s, ok := v.(*StringValue)
	if !ok || s.Value != "global" {
		t.Fatalf("lookup value = %T %#v, want raw string global", v, v)
	}
}

func TestEnvironmentSetGlobalWritesOutermost(t *testing.T) {
	t.Parallel()

	global := NewEnvironment()
	mid := NewEnclosedEnvironment(global)
	local := NewEnclosedEnvironment(mid)

	local.SetGlobal("x", NewNumberValue(3))

	if _, ok := local.Get("x"); !ok {
		t.Fatalf("x should be visible from local via chain")
	}
	v, ok := global.Get("x")
	if !ok {
		t.Fatalf("x should exist in global env")
	}
	if mustEnvNumber(t, v).Inspect() != "3" {
		t.Fatalf("global x inspect should be 3")
	}
}

func TestEnvironmentDelete(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.Set("x", NewNumberValue(5))

	if !env.Delete("x") {
		t.Fatalf("delete existing should return true")
	}
	if _, ok := env.Get("x"); ok {
		t.Fatalf("x should not exist after delete")
	}
	if env.Delete("x") {
		t.Fatalf("delete missing should return false")
	}
}

func mustEnvNumber(t *testing.T, v Value) *NumberValue {
	t.Helper()
	n, ok := v.(*NumberValue)
	if !ok {
		t.Fatalf("value type = %T, want *NumberValue", v)
	}
	return n
}
