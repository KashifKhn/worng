package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/KashifKhn/worng/internal/lexer"
)

func main() {
	data, _ := os.ReadFile(os.Args[1])
	lines := lexer.Preprocess(string(data))
	fmt.Printf("lines (%d):\n", len(lines))
	for i, l := range lines {
		fmt.Printf("[%d] %q\n", i, l)
	}
	fmt.Println("\njoined:")
	fmt.Println(strings.Join(lines, "\n"))
}
