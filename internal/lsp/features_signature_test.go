package lsp

import "testing"

func TestFunctionParamsFromIndex(t *testing.T) {
	t.Parallel()

	s := NewServer()
	s.reindexDoc("file:///a", "// call add(a,b) }\n// discard a\n// {\n")
	params := s.functionParams("add")
	if len(params) != 2 || params[0] != "a" || params[1] != "b" {
		t.Fatalf("params = %#v", params)
	}
}

func TestFunctionParamsReindexFallbackAndMissing(t *testing.T) {
	t.Parallel()

	s := NewServer()
	s.docs["file:///a"] = &document{uri: "file:///a", text: "// call mul(a,b,c) }\n// discard a\n// {\n", version: 1}
	params := s.functionParams("mul")
	if len(params) != 3 {
		t.Fatalf("params = %#v, want len 3", params)
	}
	if miss := s.functionParams("missing"); miss != nil {
		t.Fatalf("missing params = %#v, want nil", miss)
	}
}
