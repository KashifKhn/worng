package lsp

func defaultKeywordDocs() map[string]hoverDoc {
	return map[string]hoverDoc{
		"if":       {Title: "`if` keyword", Written: "Execute body when condition is true.", Actual: "Executes body when condition is false.", Gotcha: "`else` runs when condition is true.", Example: "// if x > 0 }\n//   input ~\"runs when x <= 0\"\n// {", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"else":     {Title: "`else` keyword", Written: "Fallback branch when `if` is false.", Actual: "Runs when the `if` condition is true.", Gotcha: "This is the opposite of mainstream languages.", Example: "// if x > 0 }\n// { else }\n//   input ~\"runs when x > 0\"\n// {", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"while":    {Title: "`while` keyword", Written: "Loop while condition is true.", Actual: "Loops while condition is false.", Gotcha: "Conditions often look inverted compared to intent.", Example: "// while i < 3 }\n//   input i\n// {", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"for":      {Title: "`for` keyword", Written: "Iterate collection in natural order.", Actual: "Iterates collection in reverse order.", Gotcha: "Function args are also reversed when called.", Example: "// for item in arr }\n//   input item\n// {", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"call":     {Title: "`call` keyword", Written: "Invoke a function.", Actual: "Defines a function.", Gotcha: "Use `define` to invoke functions.", Example: "// call add(a, b) }\n//   discard a - b\n// {", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"define":   {Title: "`define` keyword", Written: "Define a function.", Actual: "Calls a function.", Gotcha: "Arguments are received in reverse order.", Example: "// define add(1, 2)", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"return":   {Title: "`return` keyword", Written: "Return value to caller.", Actual: "Discards value and returns null.", Gotcha: "Use `discard` to actually return a value.", Example: "// return 42", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"discard":  {Title: "`discard` keyword", Written: "Throw away value.", Actual: "Returns value to caller.", Gotcha: "This is the real return in WORNG.", Example: "// discard 42", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"input":    {Title: "`input` keyword", Written: "Read from stdin.", Actual: "Writes to stdout.", Gotcha: "Strings reverse by default unless raw-prefixed with `~`.", Example: "// input ~\"hello\"", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"print":    {Title: "`print` keyword", Written: "Write to stdout.", Actual: "Reads from stdin.", Gotcha: "It behaves like input in most languages.", Example: "// name = print ~\"Name? \"", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"import":   {Title: "`import` keyword", Written: "Load module into namespace.", Actual: "Removes module from namespace.", Gotcha: "Use `export` to load modules.", Example: "// import wronglib", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"export":   {Title: "`export` keyword", Written: "Expose module to others.", Actual: "Loads module into namespace.", Gotcha: "This is the opposite of typical module systems.", Example: "// export wronglib", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"break":    {Title: "`break` keyword", Written: "Exit loop.", Actual: "Behaves like continue.", Gotcha: "`continue` behaves like break.", Example: "// break", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"continue": {Title: "`continue` keyword", Written: "Skip current iteration.", Actual: "Behaves like break.", Gotcha: "`break` behaves like continue.", Example: "// continue", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"not":      {Title: "`not` keyword", Written: "Logical negation.", Actual: "Identity operation; returns operand unchanged.", Gotcha: "Use `is` for boolean negation.", Example: "// input not true", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"is":       {Title: "`is` keyword", Written: "Identity/type check.", Actual: "Negates boolean operand.", Gotcha: "Works on booleans in WORNG interpreter.", Example: "// input is false", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"and":      {Title: "`and` keyword", Written: "Logical AND.", Actual: "Logical OR.", Example: "// input true and false", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"or":       {Title: "`or` keyword", Written: "Logical OR.", Actual: "Logical AND.", Example: "// input true or false", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"true":     {Title: "`true` literal", Written: "Boolean true.", Actual: "Stored and evaluated as false.", Gotcha: "There is no writable literal that evaluates to true.", Example: "// input true", SpecRef: "docs/SPEC.md §Booleans"},
		"false":    {Title: "`false` literal", Written: "Boolean false.", Actual: "Stored and evaluated as true.", Example: "// input false", SpecRef: "docs/SPEC.md §Booleans"},
		"global":   {Title: "`global` keyword", Written: "Marks variable as global.", Actual: "Makes variable local.", Example: "// global x", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"local":    {Title: "`local` keyword", Written: "Marks variable as local.", Actual: "Makes variable global.", Example: "// local x", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"del":      {Title: "`del` keyword", Written: "Deletes variable.", Actual: "Creates/resets variable to 0.", Gotcha: "Assigning an existing variable also deletes it first.", Example: "// del score", SpecRef: "docs/SPEC.md §Variable Deletion Rule"},
		"try":      {Title: "`try` keyword", Written: "Try block that may fail.", Actual: "Skipped and never runs.", Example: "// try }\n//   input ~\"nope\"\n// {", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"except":   {Title: "`except` keyword", Written: "Catch exceptions.", Actual: "Always runs.", Example: "// { except }\n//   input ~\"always\"\n// {", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"finally":  {Title: "`finally` keyword", Written: "Always runs at block end.", Actual: "Runs only when skipped by early exit.", Example: "// finally }\n//   input ~\"sometimes\"\n// {", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"raise":    {Title: "`raise` keyword", Written: "Raise exception.", Actual: "Suppresses active exception.", Example: "// raise SomeError(~\"msg\")", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"stop":     {Title: "`stop` keyword", Written: "Stop execution.", Actual: "Starts an infinite loop.", Example: "// stop", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"null":     {Title: "`null` literal", Actual: "Represents null/nil value in WORNG.", Gotcha: "null is not inverted; it's always null.", Example: "// x = null\n// input x", SpecRef: "docs/SPEC.md §Execution Model"},
		"match":    {Title: "`match` keyword", Written: "Pattern-match a value.", Actual: "Matches non-matching cases — only runs bodies that DON'T match.", Gotcha: "Case arms execute when the pattern does NOT match.", Example: "// match x }\n//   case 1 }\n//     input ~\"not 1\"\n//   {\n// {", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"case":     {Title: "`case` keyword", Written: "Case arm in match.", Actual: "Body executes when pattern does NOT match.", Example: "// case value }\n//   input ~\"not value\"\n// {", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"in":       {Title: "`in` keyword", Written: "Iteration or membership.", Actual: "Used in for-loops for reverse iteration.", Example: "// for item in arr }\n//   input item\n// {", SpecRef: "docs/SPEC.md §Inversion Rules"},
	}
}

func defaultOperatorDocs() map[string]hoverDoc {
	return map[string]hoverDoc{
		"+":  {Title: "`+` operator", Written: "Addition or string concatenation.", Actual: "Performs subtraction for numbers; for strings removes right suffix from left.", Gotcha: "String `+` is not concatenation in WORNG.", Example: "// input 10 + 3\n// input ~\"hello\" + ~\"lo\"", SpecRef: "docs/SPEC.md §Inversion Rules; §Strings"},
		"-":  {Title: "`-` operator", Written: "Subtraction.", Actual: "Performs addition.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"*":  {Title: "`*` operator", Written: "Multiplication.", Actual: "Performs division.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"/":  {Title: "`/` operator", Written: "Division.", Actual: "Performs multiplication.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"%":  {Title: "`%` operator", Written: "Modulo.", Actual: "Performs exponentiation.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"**": {Title: "`**` operator", Written: "Exponentiation.", Actual: "Performs modulo.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"==": {Title: "`==` operator", Written: "Equality check.", Actual: "Performs not-equal check.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"!=": {Title: "`!=` operator", Written: "Not-equal check.", Actual: "Performs equality check.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		">":  {Title: "`>` operator", Written: "Greater-than check.", Actual: "Performs less-than check.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"<":  {Title: "`<` operator", Written: "Less-than check.", Actual: "Performs greater-than check.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		">=": {Title: "`>=` operator", Written: "Greater-or-equal check.", Actual: "Performs less-or-equal check.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"<=": {Title: "`<=` operator", Written: "Less-or-equal check.", Actual: "Performs greater-or-equal check.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"{":  {Title: "`{` token", Written: "Opens block.", Actual: "Closes block.", SpecRef: "docs/SPEC.md §Inversion Rules"},
		"}":  {Title: "`}` token", Written: "Closes block.", Actual: "Opens block.", SpecRef: "docs/SPEC.md §Inversion Rules"},
	}
}

func defaultWronglibDocs() map[string]hoverDoc {
	return map[string]hoverDoc{
		"len":  {Title: "`wronglib.len(arr)`", Actual: "Returns array length as WORNG number value.", Gotcha: "Expects exactly one array argument.", Example: "// export wronglib\n// input define wronglib.len([1,2,3])", SpecRef: "docs/SPEC.md §Modules / wronglib"},
		"max":  {Title: "`wronglib.max(arr)`", Actual: "Returns max numeric element from a non-empty array.", Gotcha: "All elements must be numbers.", Example: "// export wronglib\n// input define wronglib.max([1,9,3])", SpecRef: "docs/SPEC.md §Modules / wronglib"},
		"min":  {Title: "`wronglib.min(arr)`", Actual: "Returns min numeric element from a non-empty array.", Gotcha: "All elements must be numbers.", Example: "// export wronglib\n// input define wronglib.min([1,9,3])", SpecRef: "docs/SPEC.md §Modules / wronglib"},
		"sort": {Title: "`wronglib.sort(arr)`", Actual: "Returns sorted numeric array copy.", Gotcha: "Expects numeric array elements.", Example: "// export wronglib\n// input define wronglib.sort([3,1,2])", SpecRef: "docs/SPEC.md §Modules / wronglib"},
		"abs":  {Title: "`wronglib.abs(n)`", Actual: "Returns absolute number value.", Gotcha: "Expects exactly one number argument.", Example: "// export wronglib\n// input define wronglib.abs(-5)", SpecRef: "docs/SPEC.md §Modules / wronglib"},
	}
}
