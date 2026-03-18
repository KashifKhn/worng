package lsp

import (
	"encoding/json"
	"testing"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func TestResolvePositionEncodingDefaults(t *testing.T) {
	t.Parallel()

	s := NewServer()
	if got := s.resolvePositionEncoding(nil); got != "utf-16" {
		t.Fatalf("encoding = %q, want utf-16", got)
	}
}

func TestResolvePositionEncodingFromV2Capabilities(t *testing.T) {
	t.Parallel()

	s := NewServer()
	p := lsproto.InitializeParamsV2{
		Capabilities: lsproto.ClientCapabilities{
			General: lsproto.GeneralClientCapabilities{
				PositionEncodings: []lsproto.PositionEncodingKind{lsproto.PositionEncodingUTF8, lsproto.PositionEncodingUTF16},
			},
		},
	}
	b, _ := json.Marshal(p)
	if got := s.resolvePositionEncoding(b); got != "utf-8" {
		t.Fatalf("encoding = %q, want utf-8", got)
	}
}

func TestResolvePositionEncodingFallbackLegacy(t *testing.T) {
	t.Parallel()

	s := NewServer()
	legacy := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"general": map[string]interface{}{
				"positionEncodings": []interface{}{"utf-32", "utf-16"},
			},
		},
	}
	b, _ := json.Marshal(legacy)
	if got := s.resolvePositionEncoding(b); got != "utf-32" {
		t.Fatalf("encoding = %q, want utf-32", got)
	}
}

func TestResolvePositionEncodingInvalidJSON(t *testing.T) {
	t.Parallel()

	s := NewServer()
	if got := s.resolvePositionEncoding(json.RawMessage(`{not json}`)); got != "utf-16" {
		t.Fatalf("encoding = %q, want utf-16", got)
	}
}

func TestChoosePositionEncodingBranches(t *testing.T) {
	t.Parallel()

	if got := choosePositionEncoding(nil); got != "utf-16" {
		t.Fatalf("nil encodings = %q, want utf-16", got)
	}
	if got := choosePositionEncoding([]string{"utf-16"}); got != "utf-16" {
		t.Fatalf("utf-16 selection = %q, want utf-16", got)
	}
	if got := choosePositionEncoding([]string{"utf-32"}); got != "utf-32" {
		t.Fatalf("utf-32 selection = %q, want utf-32", got)
	}
	if got := choosePositionEncoding([]string{"weird"}); got != "utf-16" {
		t.Fatalf("unknown selection = %q, want utf-16", got)
	}
}

func TestExtractPositionEncodingsBranches(t *testing.T) {
	t.Parallel()

	if got := extractPositionEncodings(nil); got != nil {
		t.Fatalf("nil caps = %#v, want nil", got)
	}
	if got := extractPositionEncodings(map[string]interface{}{"general": 1}); got != nil {
		t.Fatalf("invalid general = %#v, want nil", got)
	}
	if got := extractPositionEncodings(map[string]interface{}{"general": map[string]interface{}{"positionEncodings": 1}}); got != nil {
		t.Fatalf("invalid positionEncodings = %#v, want nil", got)
	}

	g := extractPositionEncodings(map[string]interface{}{
		"general": map[string]interface{}{
			"positionEncodings": []interface{}{"utf-8", 42, "utf-16"},
		},
	})
	if len(g) != 2 || g[0] != "utf-8" || g[1] != "utf-16" {
		t.Fatalf("parsed encodings = %#v, want [utf-8 utf-16]", g)
	}
}

func TestResolvePositionEncodingV2NoEncodingsFallsBack(t *testing.T) {
	t.Parallel()

	s := NewServer()
	p := lsproto.InitializeParamsV2{Capabilities: lsproto.ClientCapabilities{}}
	b, _ := json.Marshal(p)
	if got := s.resolvePositionEncoding(b); got != "utf-16" {
		t.Fatalf("encoding = %q, want utf-16", got)
	}
}
