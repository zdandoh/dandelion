// Calc.g4
lexer grammar DandelionLex;

@lexer::members {
var LineCounter = 0
}

// Tokens
SEMICOLON: ';';
COMMA: ',';
LPAREN: '(';
RPAREN: ')';
LBRACE: '{';
RBRACE: '}';
LBRACKET: '[';
RBRACKET: ']';
ASSIGN: '=';
QUOTE: '"';
ACCESS: '.';

// Primive math ops
MUL: '*';
DIV: '/';
ADD: '+';
SUB: '-';
MOD: '%';
BITWISE_OR: '|';
BITWISE_AND: '&';

// Control flow
IF: 'if';
WHILE: 'while';
FOR: 'for';
ELIF: 'elif';
ELSE: 'else';
PIPE: '->';
UNROLL: '->>';
RETURN: 'return';
YIELD: 'yield';

// Keywords
TRUE: 'true';
FALSE: 'false';
STRUCT: 'struct';
IN: 'in';
NULL: 'null';
BREAK: 'break';
CONTINUE: 'continue';
FSTART: 'f';

// Builtins
LEN: 'len';
DONE: 'done';
NEXT: 'next';
SEND: 'send';
ANY: 'any';
TYPE: 'type';

// Conditional ops
OR: '||';
AND: '&&';
NOT: '!';
LSHIFT: '<<';
RSHIFT: '>>';
LT: '<';
LTE: '<=';
GT: '>';
GTE: '>=';
EQ: '==';
NEQ: '!=';

BYTE: '\'' . '\'';
NUMBER: '-'?[0-9]+;
FLOAT: '-'?[0-9]+ '.' [0-9]+;
FMODS: 'f' ('i'|'m')? 'a'?;
IDENT: [a-zA-Z_0-9]+;
COMMAND: COMMAND_UNTERM '`';
COMMAND_UNTERM: '`' (~[`\\\r\n] | '\\' (. | EOF))*;
STRING: STRING_UNTERM '"';
STRING_UNTERM: '"' (~["\\\r\n] | '\\' (. | EOF))*;
NEWLINE : '\r'? '\n' { { LineCounter++ } } -> skip;
WHITESPACE: [ \t]+ -> skip;