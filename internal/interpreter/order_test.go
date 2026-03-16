package interpreter

import "testing"

func TestParseExecutionOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		want    ExecutionOrder
		wantErr bool
	}{
		{name: "btt", in: "btt", want: OrderBottomToTop},
		{name: "ttb", in: "ttb", want: OrderTopToBottom},
		{name: "invalid", in: "sideways", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseExecutionOrder(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseExecutionOrder(%q) error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("ParseExecutionOrder(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
