package main

import "testing"

func TestRunCLIUsageAndUnknown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want int
	}{
		{name: "no args", args: nil, want: 2},
		{name: "unknown command", args: []string{"nope"}, want: 2},
		{name: "version", args: []string{"version"}, want: 0},
		{name: "run usage", args: []string{"run"}, want: 2},
		{name: "check usage", args: []string{"check"}, want: 2},
		{name: "fmt usage", args: []string{"fmt"}, want: 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := runCLI(tc.args); got != tc.want {
				t.Fatalf("runCLI(%v) = %d, want %d", tc.args, got, tc.want)
			}
		})
	}
}
