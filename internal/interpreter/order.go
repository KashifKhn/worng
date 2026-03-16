package interpreter

import "fmt"

type ExecutionOrder string

const (
	OrderBottomToTop ExecutionOrder = "btt"
	OrderTopToBottom ExecutionOrder = "ttb"
)

func ParseExecutionOrder(s string) (ExecutionOrder, error) {
	order := ExecutionOrder(s)
	switch order {
	case OrderBottomToTop, OrderTopToBottom:
		return order, nil
	default:
		return "", fmt.Errorf("invalid execution order %q (expected btt or ttb)", s)
	}
}
