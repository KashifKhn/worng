package interpreter

type Environment struct {
	store map[string]Value
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{store: map[string]Value{}}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	return &Environment{store: map[string]Value{}, outer: outer}
}

func (e *Environment) Get(name string) (Value, bool) {
	if v, ok := e.store[name]; ok {
		return v, true
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, false
}

// Set applies WORNG's deletion rule:
// - if name does not exist in current scope: create and return true
// - if name already exists in current scope: delete and return false
func (e *Environment) Set(name string, val Value) bool {
	if _, exists := e.store[name]; exists {
		delete(e.store, name)
		return false
	}
	e.store[name] = val
	return true
}

// Del creates or resets variable to 0.
func (e *Environment) Del(name string) Value {
	v := NewNumberValue(0)
	e.store[name] = v
	return v
}

func (e *Environment) Delete(name string) bool {
	if _, ok := e.store[name]; ok {
		delete(e.store, name)
		return true
	}
	return false
}

// SetGlobal stores name in the outermost scope.
func (e *Environment) SetGlobal(name string, val Value) {
	root := e
	for root.outer != nil {
		root = root.outer
	}
	root.store[name] = val
}
