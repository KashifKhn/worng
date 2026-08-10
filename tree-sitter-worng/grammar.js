// Tree-sitter grammar for WORNG.
// The Go lexer/parser remains authoritative for execution; this grammar is for
// incremental editor parsing and syntax tooling.

module.exports = grammar({
  name: 'worng',

  extras: $ => [/[ \t\r]/],

  conflicts: $ => [
    [$.function_call_statement, $.function_call_expression],
  ],

  rules: {
    source_file: $ => repeat(choice(
      $.statement_line,
      $.block_comment,
      $.ignored_line,
      $.blank_line,
    )),

    statement_line: $ => prec(10, seq(
      $.comment_marker,
      $._statement,
      $.newline,
    )),

    block_comment: $ => seq(
      choice('/*', '!*'),
      optional($.newline),
      repeat($.block_statement),
      choice('*/', '*!'),
      optional($.newline),
    ),

    block_statement: $ => seq($._statement, optional($.newline)),

    ignored_line: $ => prec(-10, seq(/[ \t]*/, /[^\/!\n][^\n]*/, $.newline)),
    blank_line: $ => seq(/[ \t]*/, $.newline),
    newline: _ => '\n',
    comment_marker: _ => token(prec(10, /\/\/|!!/)),

    _statement: $ => choice(
      $.if_statement,
      $.while_statement,
      $.for_statement,
      $.match_statement,
      $.try_statement,
      $.function_definition,
      $.assignment,
      $.input_statement,
      $.print_statement,
      $.function_call_statement,
      $.return_statement,
      $.discard_statement,
      $.del_statement,
      $.scope_statement,
      $.import_statement,
      $.export_statement,
      $.raise_statement,
      $.stop_statement,
      $.break_statement,
      $.continue_statement,
      $.expression_statement,
    ),

    if_statement: $ => seq(
      'if', field('condition', $.expression), $.open_block, $.newline,
      repeat($.statement_line),
      $.close_block,
      optional(seq('else', $.open_block, $.newline, repeat($.statement_line), $.close_block)),
    ),

    while_statement: $ => seq(
      'while', field('condition', $.expression), $.open_block, $.newline,
      repeat($.statement_line), $.close_block,
    ),

    for_statement: $ => seq(
      'for', field('variable', $.identifier), 'in', field('iterable', $.expression),
      $.open_block, $.newline, repeat($.statement_line), $.close_block,
    ),

    match_statement: $ => seq(
      'match', field('subject', $.expression), $.open_block, $.newline,
      repeat($.case_clause), $.close_block,
    ),

    case_clause: $ => seq(
      'case', field('pattern', choice($.expression, $.wildcard)), $.open_block, $.newline,
      repeat($.statement_line), $.close_block,
    ),

    try_statement: $ => seq(
      'try', $.open_block, $.newline, repeat($.statement_line), $.close_block,
      optional(seq('except', optional(seq('(', $.identifier, ')')), $.open_block, $.newline,
        repeat($.statement_line), $.close_block)),
      optional(seq('finally', $.open_block, $.newline, repeat($.statement_line), $.close_block)),
    ),

    function_definition: $ => seq(
      'call', field('name', $.identifier), '(', optional($.parameter_list), ')',
      $.open_block, $.newline, repeat($.statement_line), $.close_block,
    ),

    assignment: $ => seq(field('name', $.identifier), '=', field('value', $.expression)),
    input_statement: $ => seq('input', $.expression),
    print_statement: $ => prec.right(10, seq('print', optional($.expression))),
    function_call_statement: $ => seq('define', $.qualified_identifier, '(', optional($.argument_list), ')'),
    return_statement: $ => prec.right(10, seq('return', optional($.expression))),
    discard_statement: $ => seq('discard', $.expression),
    del_statement: $ => seq('del', $.identifier),
    scope_statement: $ => seq(choice('global', 'local'), $.identifier),
    import_statement: $ => seq('import', $.identifier),
    export_statement: $ => seq('export', $.identifier),
    raise_statement: $ => prec.right(10, seq('raise', $.identifier, optional(seq('(', optional($.expression), ')')))),
    stop_statement: _ => 'stop',
    break_statement: _ => 'break',
    continue_statement: _ => 'continue',
    expression_statement: $ => $.expression,

    open_block: _ => '}',
    close_block: $ => seq($.comment_marker, '{'),

    parameter_list: $ => commaSep1($.identifier),
    argument_list: $ => commaSep1($.expression),

    expression: $ => choice(
      $.logical_or,
    ),

    logical_or: $ => prec.left(1, seq($.logical_and, repeat(seq('or', $.logical_and)))),
    logical_and: $ => prec.left(2, seq($.not_expression, repeat(seq('and', $.not_expression)))),
    not_expression: $ => choice(seq('not', $.not_expression), $.is_expression),
    is_expression: $ => choice(seq('is', $.is_expression), $.comparison),
    comparison: $ => prec.left(3, seq($.term, repeat(seq($.comparison_operator, $.term)))),
    comparison_operator: _ => choice('==', '!=', '<', '>', '<=', '>='),
    term: $ => prec.left(4, seq($.factor, repeat(seq(choice('+', '-'), $.factor)))),
    factor: $ => prec.left(5, seq($.unary, repeat(seq(choice('*', '/', '%', '**'), $.unary)))),
    unary: $ => choice(seq('-', $.unary), $.primary),

    primary: $ => choice(
      $.number,
      $.string,
      $.raw_string,
      $.boolean,
      $.null,
      $.array,
      $.function_call_expression,
      $.identifier,
      $.parenthesized_expression,
    ),

    parenthesized_expression: $ => seq('(', $.expression, ')'),
    function_call_expression: $ => seq('define', $.qualified_identifier, '(', optional($.argument_list), ')'),
    array: $ => seq('[', optional(commaSep1($.expression)), ']'),
    wildcard: _ => '_',
    number: _ => /-?[0-9]+(\.[0-9]+)?/,
    string: _ => choice(
      seq('"', repeat(choice(/[^"\\\n]/, /\\./)), '"'),
      seq("'", repeat(choice(/[^'\\\n]/, /\\./)), "'"),
    ),
    raw_string: $ => seq('~', $.string),
    boolean: _ => choice('true', 'false'),
    null: _ => 'null',
    qualified_identifier: $ => seq($.identifier, repeat(seq('.', $.identifier))),
    identifier: _ => /[a-zA-Z_][a-zA-Z0-9_]*/,
  },
});

function commaSep1(rule) {
  return seq(rule, repeat(seq(',', rule)));
}
