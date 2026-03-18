package lsp

import "testing"

func TestFindWordLocationsWordBoundaries(t *testing.T) {
	t.Parallel()

	locs := findWordLocations("file:///a", "alpha alphabeta alpha", "alpha")
	if len(locs) != 2 {
		t.Fatalf("locations = %d, want 2", len(locs))
	}
}

func TestFindWordLocationsEmptyWord(t *testing.T) {
	t.Parallel()

	if locs := findWordLocations("file:///a", "abc", ""); locs != nil {
		t.Fatalf("empty word locations = %#v, want nil", locs)
	}
}
