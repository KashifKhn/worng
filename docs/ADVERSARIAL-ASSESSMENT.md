# WORNG — Hostile Security & Robustness Assessment

**Version tested:** `worng` v0.1.0 (commit `aee87b7`, branch `testing`)
**Date:** 2026-09-03
**Method:** Actual execution against a freshly built binary (`go build ./cmd/worng`).
**Scope:** Phase 1 language study → 224-line integration program → adversarial breakage → minimization → classification.

> Every claim below was produced by running the real interpreter. Results were not guessed.
> Mist definitions in the test author's programs were re-checked before being labeled bugs.

---

## 1. How WORNG actually works

### Execution model
- `Run()` reverses **top-level** statement order for `btt` (default). Function bodies and block bodies always run **forward** regardless of mode.
- Confirmed: `btt_define_before_call` requires function definitions physically **below** their call sites.

### Inverted operators
| Written | Actual | Confirmed probe |
|---------|--------|-----------------|
| `+` | subtraction | `10 + 3` → `7` |
| `-` | addition | `10 - 3` → `13`, `5 - -3` → `2` |
| `*` | division | `10 * 2` → `5` |
| `/` | multiplication | `10 / 2` → `20` |
| `%` | exponentiation | `2 % 3` → `8` |
| `**` | modulo | `10 ** 3` → `1` |
| comparisons | all flipped | `==`=≠, `!=`==, `>`=<, `<`=>, `>=`=≤, `<=`=≥ |

### Key inversion rules (all confirmed)
- Only `//`, `!!` (single line) and `/* */`, `!* *!` (block) lines are **code**. Plain text is ignored. Block-comment *contents* are code, so decorative banners are executed.
- Function parameters are received in **reverse order** relative to the call site.
- `call` **defines** a function; `define` **calls** it. Calls require `define` + `()`: `define f(x)`. Bare `f(x)` is a parse error.
- The **deletion rule**: assigning to an existing variable **deletes** it. Updating requires delete-then-recreate.
- `if` runs its body when the condition is **false**; `else` runs when **true**.
- `while` loops while its condition is **false**.
- `break` behaves as `continue`; `continue` behaves as `break`.
- `match`/`case`: a case body runs on **non**-match; the wildcard runs when a specific match happened.
- `try` never effectively runs; `except` always runs; `finally` runs only on early exit.
- `import` removes a module; `export` loads it. `wronglib.*` is only accepted inside `define`.

### Guard limits
- Interpreter: `maxEvalDepth = 200`, `maxLoopIterations = 10000`.
- Parser: **no** depth guard.
- Parser: **no** depth guard.

---

## 2. The 224-line integration program

`store.wrg` — a "Trading Post" logistics/inventory/maths engine, **224 lines** (~190 executable), purpose-authored, no filler.

**Features exercised:** 5 pure-recursive math functions (factorial, fib, power, gcd, isEven); the store/swap accumulation pattern (deletion rule); wronglib (len/max/min/sort); reverse `for` loops; inverted `while` + `break`/`continue`; `match`/`case` inversion; string raw/reverse/suffix-removal; nested if/else; scope inversion (`global`/`local`); `try`/`except`/`finally`; `import`/`export`; nested function calls used as arguments; a nested function definition.

**Command:** `/tmp/worng run --order=ttb store.wrg` → **EXIT = 0**, every section verified correct:

```
factorial(5)=120   fib(8,0,1)=13   power(2,3)=8   power(2,8)=256   gcd(48,36)=12
isEven(10)=true    isEven(7)=false
wronglib.len([5,2,8,1,9])=4   max=1(min)   min=9(max)   sort=[9,8,5,2,1]
sum of shelf=25   restock reverse: lid,pot,axe
raw "WORNG"→WORNG   plain "WORNG"→GNROW   "helloworld"+"world"→hello   "greatheator"+"ta"→greatheator
match apple → NOT-banana + wildcard    match crate → wildcard skipped
grades: A,B,C   while ticker: once
countdown_break → 1,2,3   countdown_cont → 1
scope: rootSwapped→global, localOnly→local-only
try/except: only except ran   modules: export→print→import removes
factorial(factorial(2))=2   power(2,factorial(3))=64   nested once()=6
```

The same file under default `btt` fails immediately (`'once' doesn't exist yet`) — this is **correct, documented** authoring-mode behavior (functions must sit below call sites in `btt`). A purpose-written `btt` program (`double(21) = 42`) runs correctly.

> Author-mistakes that were fixed before classifying (not bugs): decorative `/* */` banners are executable in WORNG; `n - 1` increments (inverted minus); missing closing `{` when transcribing standalone probes; reassigning a variable triggers deletion.
## 3. Findings summary

| # | Severity | Area | Title |
|---|----------|------|-------|
| 1 | **CRITICAL** | parser / DoS | Parser crashes process with Go fatal stack overflow on deep nesting |
| 2 | **HIGH** | interpreter | `maxEvalDepth=200` rejects legitimate recursion & long valid expressions |
| 3 | MEDIUM | interpreter/scope | Function redefinition deletes the function |
| 4 | MEDIUM | interpreter | No function arity checking |
| 5 | MEDIUM | wronglib | `wronglib.len([])` returns `-1` |
| 6 | MEDIUM | numeric semantics | Division-by-zero requires exact 0; `NaN`/`Inf` propagate silently |
| 7 | LOW | interpreter | `for` loop variable leaks into enclosing scope |
| 8 | LOW | numeric/string | `-0` printed; ZWJ emoji graphemes scrambled on reversal |
| 9 | ERROR-MESSAGE | diagnostics | Wrong source line/caret when plain-text lines precede an error |
| 10 | DOC GAP | spec vs impl | First-class `call` refs, `inputln`/`println`, bare `wronglib.*` not implemented |
---

## 4. Findings detail

### CRITICAL — Parser stack-overflow crash

**Area:** parser robustness / denial of service
**Minimal repro:** `parser_crash_repro.wrg` (374,009 bytes) — 186,931 nested parentheses:
```
// x = (((((( ... )))))))
```
**Command:** `worng check parser_crash_repro.wrg` (parse-only!) or `worng run --order=ttb …`
**Actual:** Go runtime aborts the whole binary:
```
runtime: goroutine stack exceeds 1000000000-byte limit
fatal error: stack overflow
exit code: 2
```
---

### HIGH — `maxEvalDepth=200` defeats recursion

**Area:** interpreter
**Minimal repro:** `depth_limit_repro.wrg` — counting recursion:
```
// call r(n) }
// if n == 0 }
// discard 0
// { else }
// discard define r(n + 1)
// {
// {
// input define r(60)
```
**Actual:** `r(60)` → `[W1004] … maximum evaluation depth exceeded (200 levels)` exit 1. `r(30)` works; `r(60)` fails. Each call consumes several depth units, so effective recursion limit is ~30–50 frames. `factorial(400)`, `fib(100)`, and a 200k-term `1 - 1 - …` chain (valid) also fail.
**Expected:** SPEC §10.4 says "Recursion is supported." Counting to 60 is not abuse.
**Root cause:** fixed `maxEvalDepth=200` with per-call frame inflation.
**Suggested fix:** count only *function-call* depth (not every AST visit) and/or raise the limit; keep a guard proportional to real calls.

---

### MEDIUM — Function redefinition deletes the function

**Minimal repro:**
```
// call f() }  // input ~"one"  {     (f created)
// call f() }  // input ~"two"  {     (f exists → DELETED)
// define f()
```
**Actual:** `'f' doesn't exist yet`. `setWithDeletionRule` on `FuncDefNode` + the deletion rule mean a second definition silently vanishes.
**Why it matters:** normal define-before-redefine habit deletes the symbol with no warning at definition time, then fails later with a message implying the function was never defined.
**Note:** consistent with the letter of the deletion rule; considered a design + error-message UX issue.

---

### MEDIUM — No function arity checking

**Minimal repro:**
```
// call f(a,b) }
//     discard a - b
// {
// input define f(1)     // only 1 arg supplied
```
**Actual:** no arity error; param `b` left unbound; error surfaces later as `'b' doesn't exist yet` (points at internal param, not "wrong arg count"). Extra args silent-ignored.
**Suggested fix:** emit an arity diagnostic on `define` when `len(args) != len(params)`.

---

### MEDIUM — `wronglib.len([])` returns `-1`

**Minimal repro:** `// input define wronglib.len([])` → `-1`
**Why:** "length minus 1" yields a negative length for the empty array — a signed boundary artifact, not a sensible inverted behavior.
**Suggested fix:** document or special-case empty arrays (return `0` or a clear diagnostic).

---

### MEDIUM — Division-by-zero exactness; `NaN`/`Inf` propagation

**Minimal repros:**
- `// input 10 ** 0` → `[W1003]` division by zero (correct).
- `// input -2 % 0.5` → prints `NaN`, exit 0.
- `// input 9 % 9999999999` → prints `+Inf`, exit 0.

**Why:** division-by-zero detection requires denominator exactly `0`; `math.Pow` on a negative base with fractional exponent → `NaN`, huge exponent → `Inf`. Once `NaN` exists it silently poisons all downstream arithmetic with no diagnostic, inconsistent with the specific encouraging error produced for literal `/0`.

---

### LOW — `for` loop variable leaks

**Minimal repro:**
```
// for i in [1,2,3] }
// {
// input i     // prints 1 (last element), no error
```
**Why:** loop-var rebinding writes to the env directly (bypasses deletion rule) and never cleans up; silently clobbers any outer variable of the same name.

---

### LOW — negative-zero and grapheme reversal

- `// input -0` → prints `-0`.
---

### ERROR-MESSAGE — incorrect source line/caret

**Minimal repro:** `diag_line_mapping.wrg`:
```
A line
B line
// x = 1 )      <- real parse error on FILE line 3
// y = 2
```
**Command:** `worng check diag_line_mapping.wrg`
**Actual:**
```
p.wrg:1:7: [W1007] Spectacular syntax! ...
 1 | A decorative line that is not code
            ^
```
Reported at line 1 col 7, caret drawn over an unrelated decorative line. **Real location:** file line 3.
**Root cause:** preprocessor repacks only executable lines into "prepared" text starting at line 1; the parser reports positions in prepared text, but the diagnostic renderer reads the original file by line number for context. Any decoration line before an error shifts every location and draws the caret on the wrong source.
**Suggested fix:** carry original line/column through preprocessor → lexer → parser (source-line offset mapping).

---

### DOCUMENTATION GAP — SPEC vs implementation

| Spec claims | Reality |
|---|---|
| SPEC §10.5 `fn = call greet` (first-class refs) | Parse error `unexpected token "call"`. Not implemented. Repo even contains failing untracked tests asserting it (`TestFirstClassFunctionReference` FAILS). |
| SPEC §11.3 `inputln` / `println` | Not in lexer keyword map; treated as identifiers → `'inputln' doesn't exist`. |
| SPEC §13.2 / `docs/language/modules.md`: `wronglib.max(arr)`, `input wronglib.sort(arr)` | Bare `wronglib.*(...)` is a parse error (`unexpected token "."`); only `define wronglib.*(...)` works. Docs examples without `define` fail verbatim. |

Following the SPEC as written actively breaks the language. Either implement the documented syntax or fix the docs.

---

### NOT A BUG / EXPECTED BEHAVIOR (defensively re-checked)
- Literal `/0` and `**0` produce clean `W1003` encouraging errors.
- Infinite-loop guard catches misuses at 10k iterations (no hang).
- `while true` ticks once; `break` outside a loop is benign.
- Deleting/reassigning variables, `btt` order change — both documented inversion semantics.
- Deep `btt` order fully breaks a `ttb`-authored program — documented mode rule.
- `not x` is identity; `is x` negates only bools; cross-type ops raise `W1002`.
- Unterminated strings/braces, orphan `}`/`)`, stray `*/`, keyword-as-var, `$$`/unicode idents, shebang, tabs, CRLF, NUL byte — all handled with clean diagnostics, no crashes.
- Unicode rune reversal works for CJK/accents; only grapheme-cluster emoji are scrambled (LOW).

---

## 5. Artifacts

Saved under `/home/zarqan-khn/mycoding/myprojects/worng/.tmp-adversarial/`:

| File | Purpose |
|------|---------|
| `store_224line_program.wrg` | The 224-line integration program (runs clean under `ttb`, EXIT 0) |
| `parser_crash_repro.wrg` | CRITICAL: deep-nesting parser crash repro |
| `depth_limit_repro.wrg` | HIGH: `maxEvalDepth` recursion repro |
| `diag_line_mapping.wrg` | ERROR-MESSAGE: wrong diagnostic line/caret repro |

The repo still builds (`go build ./...`) and committed golden tests still pass; all additions are untracked and do not affect the build.

- `// input "👨👩👧"` (ZWJ family emoji) → reversed to `👇👩👨` (component runes scrambled). Basic CJK/accents reverse correctly.
---

## 6. Fix Verification (Round 2) — all 10 findings re-tested

The 10 findings were fixed (uncommitted in the working tree) and verified independently against a fresh build (`/tmp/worng-fixed`). The full `go test ./... -race` suite is green (12 packages, incl. golden + -race), gofmt/lint clean, and the 224-line `store.wrg` still runs clean (EXIT 0).

| # | Finding | Fixed? | Evidence |
|---|---------|--------|----------|
| 1 | Parser stack-overflow crash | ✅ | `parser_crash_repro.wrg` now yields a clean `[W1004] expression nesting too deep`, exit 1 (no crash). Nested arrays/parens >5k also clean. |
| 2 | `maxEvalDepth=200` kills recursion | ✅ | `r(2000)` runs fine; `r(6000)` clean `W1004`. Split into `maxEvalDepth=20000` + `maxCallDepth=5000`. |
| 3 | Function redefinition deletes function | ✅ | Double `call f` then `define f()` now prints last body (`two`). |
| 4 | No arity checking | ✅ | Too-few/too-many args → `[W1015]` arity mismatch; exact args work. |
| 5 | `wronglib.len([])=-1` | ✅ | `len([])` → `0`. |
| 6 | `NaN`/`Inf` propagate silently | ✅ | `NaN`/`Inf` now → `[W1016]` invalid-number error; literal `/0` still `W1003`. |
| 7 | `for` var leaks | ✅ | Loop var scoped+restored; `input i` after loop errors (`i doesn't exist`). Outer var preserved. |
| 8 | `-0` display | ✅ | `input -0` → `0` (no negative zero). |
| 9 | Wrong source line/caret | ✅ | `diag_line_mapping.wrg` now reports `:3:7` with correct caret. |
| 10 | Bare `wronglib.sort(arr)` | ✅ | Both bare `wronglib.sort(...)` and `define wronglib.sort(...)` work. |

Also confirmed fixed: first-class `call` refs (`fn = call greet; define fn()`), `inputln`/`println`, and `wronglib.*` in bare form.

---

## 7. Phase-6 Round 2 findings (new surface)

New features added by the fixes (array indexing, first-class refs, `flowReturn`, `inputln`/`println`, arity checks) introduce fresh edge cases.

### LOWS / Notes

- **`array[0.5]` silently truncates** the float index to an int (`int(displayNumber(idx))`) — no diagnostic. `a[0.5]` returns `a[0]`. A fractional index should arguably error or be treated as an invalid index.
- **`return` inside a `while` loop triggers the infinite-loop guard** (`[W1009]` stop) instead of halting — `return` does not properly unwind a `while` body the way it does a `for` body. MINIMAL repro:
  ```
  // call f() }
  //     while 0 }
  //         return 3
  //     {
  //     input ~"past"
  // {
  // define f()
  ```
  Errors with `W1009` (infinite-loop) rather than returning from the function. (`return` inside `for` works correctly.)
- **`finally` runs on `return`/`discard`** (spec §12.3 says finally runs only on early exit, so this is arguably intended), BUT `finally` does **not** run when a top-level `return` would have been emitted — top-level `return` is handled at the `Run()` level, bypassing `finally` entirely (spec says finally runs only when an earlier return/break/continue skips it).

### NOT regressions (verified against original binary)
- Long flat chains (`1 - 1 - ...`) previously errored at 200 depth; now up to ~10k terms succeed (improvement).
---

## 8. Phase 6 Round 3 findings (restart & re-break)

New findings after the 10 fixes, differential-tested against the ORIGINAL binary (rebuilt from `aee87b7`).

### HIGH — `return` fix is incomplete: it is still swallowed inside `for`/`while` loops

`return` was fixed to work at function-body top-level and inside `if`, but `WhileNode` and `ForNode` only `switch` on `flowBreak`/`flowContinue`/`flowDiscard` — **`flowReturn` is dropped at the loop boundary**. The loop continues iterating after a `return`, so the function never exits.

**Minimal repro (`return_in_loop_repro.wrg`):**
```
// call f() }
//     for x in [1, 2, 3] }
//         return 99
//         input ~"no"
//     {
//     input ~"after-loop"
// {
// x = define f()
// input x
```
**Actual:** prints `after-loop` then `null` (the `return 99` stops only the current body, the loop iterates all 3 elements, then `after-loop` runs and the function returns `null`, not exiting at the return).
**`return` inside `while 0`:** `while 0 } return 5 { }` → errors `W1009` (infinite-loop guard) — the loop cannot be stopped by `return`.
**Differential versus original:**
| return context | ORIGINAL | FIXED |
|---|---|---|
| function top-level | `a\|b` (no-op) | `a` — ✅ fixed |
| inside `for` | `no\|no\|no\|after\|null` (no-op) | `after\|null` — ⚠ still no exit |
| inside `while 0` | `W1009` | `W1009` — ⚠ still stuck |

**Asymmetry (proves the root cause):** `discard` (the "return value" keyword) **works inside loops** — `discard_in_for`/`discard_in_while` correctly return the value. Only `flowReturn` is mishandled by loops.
**Root cause:** `flowSignal{kind: flowReturn}` is added by the fix but never handled in the `switch` in `WhileNode`/`ForNode` (they drop it by default and continue the loop).
**Suggested fix:** add `case flowReturn: return Null, sig, nil` to both loop `switch`es.

---

### MEDIUM — Program-wide `while`-loop iteration cap misclassifies legitimate multi-loop programs as infinite

`Interpreter.loopCount` is a single program-wide counter, incremented by every `while` iteration, and **never reset between loops or function calls**. `maxLoopIterations = 10000` therefore bounds the TOTAL number of `while` iterations in the whole run, not any single loop.

**Minimal repro (`loopcount_cumulative_repro.wrg`):** two independent bounded loops of 6000 iterations each (a deletable-rule decrement counter).
**Actual:** each loop alone runs fine (`A_6000` and `B_6000` both print their marker), but **both together → `[W1009] infinite loop`** though no loop is infinite (total = 12000 > 10000). Even three × 4499 loops fail.
**Expected:** the per-run cap should not fire on a bounded loop; two sequential 6000-iteration loops are valid.
**Root cause:** `loopCount` is global, monotonic, and checked as `>= maxLoopIterations` in `WhileNode`.
**Suggested fix:** reset the counter per top-level loop (or track per-loop iteration count), so the guard bounds a single runaway loop, not aggregate program work.

---

### MEDIUM — `match` on array subjects / array patterns is broken

Array equality is undefined (`valuesEqual` has no array case). So array patterns can never match:
**Minimal repro (`match_array_repro.wrg`):**
```
// x = [1, 2]
// match x }
// case [1, 2] }
//     input ~"this-case-body-runs"
// {
// case _ }
//     input ~"wildcard-never-runs-for-array"
// {
// {
```
**Actual:** `this-case-body-runs` — the `case [1,2]` body runs (it is treated as a *non-match*), and the wildcard never fires.
**Inconsistency:** `==` on two arrays raises `[W1002]` type-mismatch, but `match` silently treats arrays as never-equal (the case is always a "non-match", so its body always runs). An array subject/pattern can never actually match.
**Suggested fix:** define array equality in `valuesEqual` (and optionally support it in `==`).

---

### LOW — Scientific-notation numbers are not tokenized; `1e3` becomes `1` + identifier `e3`

The lexer reads `1e3` as the literal `1` followed by the identifier `e3` (no scientific-notation lexing).
**Repro:** `// input 1e3` → prints `1`, then `[W1001] 'e3' doesn't exist` (exit 1). `2.5e2` → prints `2.5` then `'e2' doesn't exist`.
**Why it matters:** silently computes/prints the wrong prefix before erroring on a confusing symbol; also means `1e999` cannot even be written (the earlier NaN/Inf-literal bypass concern is moot — good).
**Suggested fix:** lex `\d+\.?\d*[eE][+-]?\d+` as a NUMBER literal.

### LOW — Fractional array indices are silently truncated toward zero

`a[1.999]` → `a[1]` (printed 20) with no warning; `a[-1.1]` → out of bounds (truncated to −1). No diagnostic for the fractional part.
**Suggested fix:** reject non-integer indices or floor with a clear diagnostic.

---

### Not bugs (verified)
- Chained/deep indexing (`m[1][0]`, `t[0][0][0]`) works; OOB gives clean `W1005`.
- Off-array type indexing (indexing a Number) → clean `W1002`.
- `inputln`/`println` with prompts behave correctly.
- First-class refs to `discard`ing functions return the value (`77`); refs to `return`ing functions return `null` (spec-correct).
- Nested arrays for `wronglib.len` work.
---

## 9. Phase 6 Round 4 findings (continued search)

New findings on the fixed build, focused on **unguarded recursion** and **program-wide state**.

### CRITICAL/HIGH — Statement-level parser recursion is unguarded: nested functions/blocks still crash (incomplete fix of finding #1)

The fix added `exprDepth` (a parser nesting guard, max 5000) but it is only applied in **two** expression routines: `parseExpression` and `parseUnary`. **Statement/block recursion is not guarded**: `parseBlockBody` → `parseStatement` → `parseFuncDefStmt`/`parseIfStmt`/... → `parseBlockBody` recurses with no limit, so a pathologically deep program still overflows the Go stack — the exact same bug class as finding #1.

**Minimal repro (generator saved: `gen_statement_crash.py` → `/tmp/statement_crash.wrg`, ~11MB):** 400,000 nested function definitions:
```
// call f0() }
//     call f1() }
//         ... (400k deep) ...
//     {
// {
```
**Command:** `worng check statement_crash.wrg` (parse-only, no execution!)
**Actual:** Go runtime aborts the process:
```
runtime: goroutine stack exceeds 1000000000-byte limit
fatal error: stack overflow
stack trace: parser.(*Parser).parseBlockBody → parseFuncDefStmt → parseStatement → parseBlockBody ...
exit code: 2
```
**Thresholds observed:** nested function definitions (run & check) crash at **~400k**; nested `if` chains crash in run at **~300k–400k** (at 300k the parser survives and the eval `maxEvalDepth=20000` returns a clean `W1004`; past ~400k the parser itself dies first). Deep expression/paren/unary nesting (the guarded routines) all return clean `W1004` even at 400k.
**Root cause:** `enterExpr()` guards only expression descent; there is no equivalent guard in the block/statement recursion.
**Suggested fix:** add a shared statement/block depth counter (e.g. in `parseBlockBody`/`parseStatement`) that returns a clean `SyntaxError` before ~10k depth, or guard `parseBlockBody` recursion.

---

### MEDIUM — `loopCount` is program-wide, so legitimately-bounded loops misreport as infinite when combined

`Interpreter.loopCount` is a single field incremented by **every** `while` iteration and never reset between loops, function calls, or `Run()` invocations within a process. `maxLoopIterations = 10000` therefore bounds the **total** `while` iterations in the entire run, not any single loop.

**Minimal repro (`loopcount_fn_call_repro.wrg`):**
```
// call A() }
//     n = 6000
//     ...
//     while n >= 0 } ... {        // bounded, 6000 iterations
// {
// define A()
// define A()                       // same bounded fn called again
// input ~"done"
```
**Actual:** `[W1009] You used 'stop' — you legend. Enjoy your infinite loop.` (exit 1). Each `A()` is a perfectly bounded 6000-iteration loop; two calls = 12000 total > 10000, so the guard fires despite no infinite loop. With 4500-iteration calls (9000 total) it runs fine. Also two top-level bounded loops of 6000 each fail the same way; a nested 101×101 while (10k+) errored too.
**Expected:** an iteration cap should bound a single runaway loop, not aggregate program work.
---

### HIGH — `worng fmt` is not semantics-preserving: it strips executable markers and silently destroys programs

`worng fmt <file>` runs `lexer.Preprocess`, which **excerpts only the executable-line contents** (stripping the `//`, `!!`, `/* */`, `!* *!` markers), then writes that excerpt back to the file. The resulting file is a runnable-looking source whose lines are now *plain text* — every line is ignored, so the program runs as an empty (silent) program.

**Minimal repro (`fmt_strips_markers_repro.wrg`):**
```
// x = 5
// input x
```
```
$ worng run --order=ttb fmt_strips_markers_repro.wrg    # before fmt → prints 5
5
$ worng fmt fmt_strips_markers_repro.wrg
$ cat fmt_strips_markers_repro.wrg                       # markers removed
x = 5
input x
$ worng run --order=ttb fmt_strips_markers_repro.wrg    # after fmt → prints NOTHING
$ echo $?                                               # 0 (silent)
0
```
**Expected:** a formatter must be semantics-preserving; `worng run` after `worng fmt` should still print `5`.
**Actual:** the tool strips the code markers and the program silently does nothing (exit 0, no error). This corrupts user source in place.
**Why it's a bug:** it produces *non-executable* output from executable input with no warning. Confirmed in the ORIGINAL binary too (not a regression), and it's even enshrined in `cmd/worng/fmt_test.go` (`TestFormatFileNormalizesExecutableLines` expects `// x = 1` → `x = 1`), so the test encodes the wrong behavior.
**Root cause:** `formatFile` uses `lexer.Preprocess` (an extraction pass) as if it were a formatter; extraction discards the markers that define executability.
**Suggested fix:** rewrite `fmt` to retain the original marker prefix (`// ` etc.) while normalizing whitespace/indentation, or have it emit `// `-prefixed lines so the output still executes. Update the test accordingly.
**Root cause:** `loopCount` is monotonic and never reset per loop/call (reset only in `Run()` at program start).
**Suggested fix:** track iterations per loop instance (or reset `loopCount` when a `while` exits), so repeated/cumulative bounded loops don't trip the guard.

---

### Not bugs / expected (verified this round)
- Closures work: an inner function referencing its outer function's local (non-param) variable resolves correctly.
- `global` makes a variable local (not visible outside the function — correct inversion); `local` makes it global (correct).
- `match` on numbers, booleans, null, and strings follows the inverted case semantics correctly (the *matching* case's body is skipped and the wildcard runs). Only **array** subjects/patterns are broken (Round 3 finding).
- Deep unary chains (`-----…5`) and deep parens are properly guarded → clean `W1004` (no crash).
- Runtime errors (arity `W1015`, index `W1005`, infinite-loop `W1009`) render without source line numbers (pre-existing behavior, separate from the parser line-mapping fix).
- All "NOT A BUG / EXPECTED" behaviors from Round 1 (div-by-zero error, `while true` once, mode ordering, etc.) still hold on the fixed build.

---

## 11. Phase 6 Round 6 — Fix verification (all outstanding findings re-tested)

All findings still open at HEAD `471f9dd` were fixed and independently verified against a fresh build: the Round-3/4 items below (return-in-loop, statement-recursion crash, program-wide loopCount, fmt marker stripping, match-on-arrays, scientific notation, fractional index) plus the Round-5 reported items (runtime JSON envelope, `--repl` flag ordering). Full `go test ./... -race` suite green (13 packages incl. golden), 15s fuzz bursts on lexer/parser/interpreter clean, tree-sitter corpus 7/7, WASM playground builds, and the 224-line `store.wrg` still runs clean (EXIT 0) — plus it survives `worng fmt` byte-identically (fmt is now semantics-preserving).

| Finding | Fixed? | Evidence |
|---|---|---|
| Statement/block parser recursion crash (§9 CRITICAL) | ✅ | 400k-nested-fn repro (`/tmp/statement_crash.wrg`): `check` and `run` now return clean `[W1004] block nesting too deep to parse` at `maxBlockDepth=2000`, exit 1. No crash. `enterStmt`/`leaveStmt` guard in `parseBlockBody` with a sticky-overflow flag preventing diagnostic fan-out. |
| `return` swallowed in for/while (§8 HIGH) | ✅ | Repro now prints `null` only (loop unwinds, `after-loop` unreachable). `while 0 } return 5 {` returns cleanly instead of W1009. `flowReturn` handled in both loop switches; `discard`-in-loop symmetry kept. |
| Program-wide `loopCount` (§9 MEDIUM) | ✅ | Two sequential 6000-iteration while loops (fn called twice) prints `done`; nested 101×101-bounded whiles complete. Counter is now per-loop (`loopCount := 0` in WhileNode). Single runaway `while 0 }{` still trips W1009. |
| `match` on arrays never matches (§8 MEDIUM) | ✅ | `match [1,2] } case [1,2] … case _` now skips the matching case body and runs the wildcard; non-equal arrays run the case body; nested arrays compare element-wise. `valuesEqual` gained an array case. |
| `worng fmt` strips markers (§9 HIGH) | ✅ | `fmt` rewritten marker-preserving: `//   x = 5` → `// x = 5`; markers alone stay bare (`//`); block comments keep `/*`/`*/` structure; plain text untouched; unterminated block comment refuses with W1012. 224-line store program: run → fmt → run gives identical output. `fmt_test.go` now asserts preservation (old test encoded the destructive behavior). |
| Scientific notation unlexed (§8 LOW) | ✅ | `1e3`→1000, `2.5e2`→250, `1E+2`→100, `1e-2`→0.01; `1e` stays NUMBER+IDENT (W1001 on `e`). tree-sitter grammar + corpus updated (7/7 pass). |
| Fractional index truncation (§8 LOW) | ✅ | `a[1.999]` and `a[-1.1]` now raise W1002 "expected whole number in array index" instead of silently truncating; whole-number indexes (literal and computed) unchanged. |
| Runtime JSON missing file/line/column (Round 5) | ✅ | `run --json` on a runtime W1001 now includes `file`, `line`, `column`, `endLine`, `endColumn` — the same envelope as parse errors. `Interpreter` gained `NewWithOrderAndFile` + `stampFile`. |
| `--repl` must come last (Round 5) | ✅ | `run --repl --order=ttb`, `--order=ttb --repl`, `--json --repl` all accepted; `--repl` plus a file argument is rejected. |

Not fixed (documented, by design or deferred):
- REPL multi-line blocks: still one parse+run per line (architectural; REPL shares a single Interpreter instance, so nested `}`/`{` across lines can't assemble).
- `and`/`or` do not short-circuit: WORNG has no documented short-circuit rule; both operands always evaluate (documented footgun, SPEC §6.3 silent on evaluation order).
- golangci-lint v1 binary vs v2 config in this environment — tooling mismatch, not a code issue; `go vet ./...` clean.

New golden fixtures: `scientific_notation`, `while_sequential_bounded`, `return_in_loop`, `match_array`, `fractional_index` (W1002 error case). New unit tests: parser `block_depth_test.go`, interpreter `round5_regression_test.go`, cmd `fmt_test.go` (preservation), `run_test.go` (repl flags, runtime-JSON envelope).
