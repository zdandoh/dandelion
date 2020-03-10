// Calc.g4
lexer grammar DandelionLex;

// Tokens
SEMICOLON: ';';
COMMA: ',';
LPAREN: '(';
RPAREN: ')';
LBRACE: '{';
RBRACE: '}';
LBRACKET: '[';
RBRACKET: ']';
FSTART: 'f';
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
RETURN: 'return';
YIELD: 'yield';

// Keywords
TRUE: 'true';
FALSE: 'false';
STRUCT: 'struct';
NEXT: 'next';
SEND: 'send';
NULL: 'null';

// Conditional ops
OR: '||';
AND: '&&';
LT: '<';
LTE: '<=';
GT: '>';
GTE: '>=';
EQ: '==';
NEQ: '!=';

BYTE: '\'' . '\'';
NUMBER: '-'?[0-9]+;
FLOAT: '-'?[0-9]+ '.' [0-9]+;
IDENT: [a-zA-Z_0-9]+;
COMMAND: COMMAND_UNTERM '`';
COMMAND_UNTERM: '`' (~[`\\\r\n] | '\\' (. | EOF))*;
STRING: STRING_UNTERM '"';
STRING_UNTERM: '"' (~["\\\r\n] | '\\' (. | EOF))*;
WHITESPACE: [ \r\n\t]+ -> skip;