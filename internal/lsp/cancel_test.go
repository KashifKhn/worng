package lsp

import (
	"encoding/json"
	"testing"
)

func TestIsCanceledPath(t *testing.T) {
	t.Parallel()

	s := NewServer()
	id := json.RawMessage("1")
	if isCanceled(s, id) {
		t.Fatal("unexpected canceled=true")
	}
	s.canceled["1"] = true
	if !isCanceled(s, id) {
		t.Fatal("expected canceled=true")
	}
	if isCanceled(s, id) {
		t.Fatal("expected one-shot cancel consumption")
	}
}
