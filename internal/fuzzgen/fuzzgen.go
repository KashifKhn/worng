// Package fuzzgen generates syntactically valid WORNG source programs for
// structure-aware fuzzing. Random bytes rarely reach deep interpreter logic
// because the lexer/parser reject them early. This generator produces programs
// that always parse successfully, allowing the fuzzer to exercise the
// interpreter's semantic evaluation paths instead.
//
// Usage — wire into a fuzz target:
//
//	f.Fuzz(func(t *testing.T, data []byte) {
//	    src := fuzzgen.Program(data)
//	    // run src through the full pipeline
//	})
//
// The generator is deterministic: the same data slice always produces the
// same program. It uses the input bytes as a stream of decision bits, cycling
// back to the start when exhausted.
package fuzzgen

import (
	"fmt"
	"strings"
)

// Program generates a syntactically valid WORNG source program from the
// provided byte slice. The byte slice is used as a source of randomness;
// every byte drives one or more structural decisions.
func Program(data []byte) string {
	if len(data) == 0 {
		return "// input ~\"fuzz\"\n"
	}
	g := &generator{data: data}
	return g.program()
}

// generator consumes bytes from data to make structural decisions.
type generator struct {
	data []byte
	pos  int
}

// next returns the next byte, cycling through data.
func (g *generator) next() byte {
	b := g.data[g.pos%len(g.data)]
	g.pos++
	return b
}

// pick returns a value in [0, n).
func (g *generator) pick(n int) int {
	if n <= 1 {
		return 0
	}
	return int(g.next()) % n
}

// bool returns a pseudo-random bool.
func (g *generator) bool() bool {
	return g.next()%2 == 0
}

// ident picks one of a fixed set of variable names — keeps programs readable
// and increases the chance of variable reuse (which exercises deletion rules).
var idents = []string{"x", "y", "z", "a", "b", "n", "i", "v", "result", "tmp"}

func (g *generator) ident() string {
	return idents[g.pick(len(idents))]
}

// number picks a small number literal, including 0 and negative-looking values.
var numbers = []string{"0", "1", "2", "3", "5", "10", "42", "100"}

func (g *generator) number() string {
	return numbers[g.pick(len(numbers))]
}

// rawStr picks a raw string literal (never reversed on output — safe to use
// in assertions without worrying about reversal).
var rawStrings = []string{
	`~"ok"`, `~"a"`, `~"hello"`, `~"0"`, `~"true"`, `~"false"`, `~"null"`,
}

func (g *generator) rawStr() string {
	return rawStrings[g.pick(len(rawStrings))]
}

// program emits between 1 and 6 top-level statements.
func (g *generator) program() string {
	var b strings.Builder
	count := 1 + g.pick(6)
	for i := 0; i < count; i++ {
		g.stmt(&b, 0)
	}
	return b.String()
}

// maxDepth is the maximum block nesting depth. Beyond this, only leaf
// statements (assign, input, del, scope) are emitted.
const maxDepth = 4

// stmt emits one statement wrapped in a WORNG comment marker (// ).
// depth limits block nesting to prevent runaway recursion.
func (g *generator) stmt(b *strings.Builder, depth int) {
	// At or beyond maxDepth, emit only leaf statements (no block nesting).
	if depth >= maxDepth {
		switch g.pick(4) {
		case 0:
			g.stmtAssign(b)
		case 1:
			g.stmtInput(b)
		case 2:
			g.stmtDel(b)
		default:
			g.stmtScope(b)
		}
		return
	}

	// Weight simpler statements more heavily at deeper nesting
	maxChoice := 10
	if depth >= 2 {
		maxChoice = 5
	}

	switch g.pick(maxChoice) {
	case 0:
		g.stmtAssign(b)
	case 1:
		g.stmtInput(b)
	case 2:
		g.stmtDel(b)
	case 3:
		g.stmtIf(b, depth)
	case 4:
		g.stmtWhile(b, depth)
	case 5:
		g.stmtFor(b, depth)
	case 6:
		g.stmtTry(b, depth)
	case 7:
		g.stmtCall(b, depth)
	case 8:
		g.stmtDefine(b)
	case 9:
		g.stmtScope(b)
	default:
		g.stmtInput(b)
	}
}

// line emits a single executable WORNG line: "// <content>\n"
func line(b *strings.Builder, content string) {
	b.WriteString("// ")
	b.WriteString(content)
	b.WriteByte('\n')
}

// expr generates a simple expression (no blocks).
func (g *generator) expr() string {
	switch g.pick(8) {
	case 0:
		return g.number()
	case 1:
		return g.ident()
	case 2:
		return g.rawStr()
	case 3:
		return "true"
	case 4:
		return "false"
	case 5:
		return "null"
	case 6:
		return g.binaryExpr()
	case 7:
		return g.unaryExpr()
	default:
		return g.number()
	}
}

var binaryOps = []string{"+", "-", "*", "/", "%", "**", "==", "!=", "<", ">", "<=", ">=", "and", "or"}

func (g *generator) binaryExpr() string {
	op := binaryOps[g.pick(len(binaryOps))]
	return fmt.Sprintf("%s %s %s", g.simpleExpr(), op, g.simpleExpr())
}

func (g *generator) unaryExpr() string {
	if g.bool() {
		return fmt.Sprintf("not %s", g.simpleExpr())
	}
	return fmt.Sprintf("is %s", g.simpleExpr())
}

// simpleExpr avoids recursive binary to keep expression depth bounded.
func (g *generator) simpleExpr() string {
	switch g.pick(5) {
	case 0:
		return g.number()
	case 1:
		return g.ident()
	case 2:
		return "true"
	case 3:
		return "false"
	default:
		return g.number()
	}
}

// arrayExpr generates a small array literal like [1, 2, 3].
func (g *generator) arrayExpr() string {
	count := 1 + g.pick(4)
	elems := make([]string, count)
	for i := range elems {
		elems[i] = g.number()
	}
	return "[" + strings.Join(elems, ", ") + "]"
}

// --- Statement generators ---

func (g *generator) stmtAssign(b *strings.Builder) {
	line(b, fmt.Sprintf("%s = %s", g.ident(), g.expr()))
}

func (g *generator) stmtInput(b *strings.Builder) {
	line(b, fmt.Sprintf("input %s", g.expr()))
}

func (g *generator) stmtDel(b *strings.Builder) {
	line(b, fmt.Sprintf("del %s", g.ident()))
}

func (g *generator) stmtScope(b *strings.Builder) {
	kw := "global"
	if g.bool() {
		kw = "local"
	}
	line(b, fmt.Sprintf("%s %s", kw, g.ident()))
}

// block emits an opening }, body statements, and closing {.
// In WORNG: } opens a block, { closes it.
func (g *generator) block(b *strings.Builder, depth int) {
	// } is TOKEN_LBRACE (opens block) — appended to the preceding line
	// { is TOKEN_RBRACE (closes block) — on its own line
	b.WriteString("}\n") // close the preceding header line's block open
	count := g.pick(3)
	for i := 0; i < count; i++ {
		g.stmt(b, depth+1)
	}
	line(b, "{")
}

// headerLine emits "// <keyword> <expr> }" — the opening line of a block statement.
func headerLine(b *strings.Builder, content string) {
	b.WriteString("// ")
	b.WriteString(content)
	b.WriteString(" }")
	b.WriteByte('\n')
}

func (g *generator) stmtIf(b *strings.Builder, depth int) {
	headerLine(b, fmt.Sprintf("if %s", g.expr()))
	// body (consequence — runs when condition is FALSE in WORNG)
	count := g.pick(2)
	for i := 0; i < count; i++ {
		g.stmt(b, depth+1)
	}
	if g.bool() {
		// else clause — runs when condition is TRUE in WORNG
		b.WriteString("// { else }\n")
		count2 := g.pick(2)
		for i := 0; i < count2; i++ {
			g.stmt(b, depth+1)
		}
	}
	line(b, "{")
}

func (g *generator) stmtWhile(b *strings.Builder, depth int) {
	// Use a condition that will eventually terminate:
	// "while false }" loops (false is truthy in WORNG — stored as true).
	// Use a numeric counter to ensure termination in most generated programs.
	v := g.ident()
	// Initialise counter — two assigns needed (first deletes if exists, second creates)
	line(b, fmt.Sprintf("%s = 0", v))
	line(b, fmt.Sprintf("%s = 0", v))
	headerLine(b, fmt.Sprintf("while %s == 3", v))
	// body
	count := g.pick(2)
	for i := 0; i < count; i++ {
		g.stmt(b, depth+1)
	}
	// Increment: x = x - 1 → in WORNG - means +, so x grows by 1
	line(b, fmt.Sprintf("%s = %s - 1", v, v))
	line(b, fmt.Sprintf("%s = %s - 1", v, v))
	line(b, "{")
}

func (g *generator) stmtFor(b *strings.Builder, depth int) {
	v := g.ident()
	arr := g.arrayExpr()
	headerLine(b, fmt.Sprintf("for %s in %s", v, arr))
	count := g.pick(3)
	for i := 0; i < count; i++ {
		g.stmt(b, depth+1)
	}
	line(b, "{")
}

func (g *generator) stmtTry(b *strings.Builder, depth int) {
	// try body (never runs in WORNG)
	b.WriteString("// try }\n")
	// Put at least one stmt in try body so { except } can be on same line
	// without newline issues (matches parser's expectation)
	line(b, fmt.Sprintf("input %s", g.rawStr()))
	// except always runs
	b.WriteString("// { except }\n")
	count := g.pick(3)
	for i := 0; i < count; i++ {
		g.stmt(b, depth+1)
	}
	if g.bool() {
		// finally — runs only on early exit from except
		b.WriteString("// { finally }\n")
		count2 := g.pick(2)
		for i := 0; i < count2; i++ {
			g.stmt(b, depth+1)
		}
	}
	line(b, "{")
}

// params generates a parameter list string like "a, b".
var paramSets = [][]string{
	{},
	{"a"},
	{"a", "b"},
	{"x", "y", "z"},
}

func (g *generator) stmtCall(b *strings.Builder, depth int) {
	name := fmt.Sprintf("fn%d", g.pick(4))
	params := paramSets[g.pick(len(paramSets))]
	paramStr := strings.Join(params, ", ")
	headerLine(b, fmt.Sprintf("call %s(%s)", name, paramStr))
	count := g.pick(3)
	for i := 0; i < count; i++ {
		g.stmt(b, depth+1)
	}
	if g.bool() {
		line(b, fmt.Sprintf("discard %s", g.expr()))
	} else {
		line(b, "return")
	}
	line(b, "{")
}

func (g *generator) stmtDefine(b *strings.Builder) {
	// define calls a function — emit args matching one of the paramSets
	name := fmt.Sprintf("fn%d", g.pick(4))
	params := paramSets[g.pick(len(paramSets))]
	args := make([]string, len(params))
	for i := range args {
		args[i] = g.simpleExpr()
	}
	argStr := strings.Join(args, ", ")
	line(b, fmt.Sprintf("define %s(%s)", name, argStr))
}
