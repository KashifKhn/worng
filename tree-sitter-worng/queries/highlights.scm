; Executable comment markers are code markers in WORNG.
(comment_marker) @punctuation.definition.comment
(ignored_line) @comment

; Inverted block delimiters.
(open_block) @punctuation.bracket
(close_block) @punctuation.bracket

; Keywords.
["if" "else" "while" "for" "in" "match" "case"
 "try" "except" "finally" "call" "define" "return"
 "discard" "input" "print" "del" "global" "local"
 "import" "export" "raise"]
 @keyword

(stop_statement) @keyword
(break_statement) @keyword
(continue_statement) @keyword

(function_definition name: (identifier) @function)
(function_call_expression (qualified_identifier) @function.call)
(function_call_statement (qualified_identifier) @function.call)
(assignment name: (identifier) @variable)
(identifier) @variable
(number) @number
(string) @string
(raw_string (string) @string.special)
(boolean) @boolean
(null) @constant.builtin
(wildcard) @constant.builtin

["+" "-" "*" "/" "%" "**"
 "==" "!=" "<" ">" "<=" ">="
 "and" "or" "not" "is"]
 @operator
