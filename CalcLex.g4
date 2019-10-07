// Calc.g4
lexer grammar CalcLex;

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

// Primive math ops
MUL: '*';
DIV: '/';
ADD: '+';
SUB: '-';
BITWISE_OR: '|';
BITWISE_AND: '&';

// Control flow
IF: 'if';
WHILE: 'while';
FOR: 'for';
ELIF: 'elif';
ELSE: 'else';
PIPE: '->';

// Keywords
TRUE: 'true';
FALSE: 'false';

// Conditional ops
OR: '||';
AND: '&&';
LT: '<';
LTE: '<=';
GT: '>';
GTE: '>=';
EQ: '==';

NUMBER: [0-9]+;
IDENT: [a-zA-Z_0-9]+;
COMMAND: COMMAND_UNTERM '`';
COMMAND_UNTERM: '`' (~[`\\\r\n] | '\\' (. | EOF))*;
STRING: STRING_UNTERM '"';
STRING_UNTERM: '"' (~["\\\r\n] | '\\' (. | EOF))*;
WHITESPACE: [ \r\n\t]+ -> skip;