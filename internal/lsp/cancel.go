package lsp

import "encoding/json"

func isCanceled(s *Server, id json.RawMessage) bool {
	if len(s.canceled) == 0 {
		return false
	}
	key := string(id)
	if s.canceled[key] {
		delete(s.canceled, key)
		return true
	}
	return false
}
