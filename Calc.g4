// Calc.g4
grammar Calc;

// Tokens

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
IDENT: [a-zA-Z_]+;
WHITESPACE: [ \r\n\t]+ -> skip;

// Rules
start : line+ EOF;

line: (expr|statement) ';';
arglist: IDENT (',' IDENT)* (',')?;
explist: expr? (',' expr)*;
body: lines=line*;

expr
   : '(' expr ')'                              # ParenExp
   | '[' elems=explist ']'                     # Array
   | expr '[' index=expr ']'                   # SliceExp
   | expr op=(LT|LTE|GT|GTE|EQ) expr           # CompExp
   | expr op=(MUL|DIV) expr                    # MulDiv
   | 'f' '{' body '}'                          # FunDef
   | 'f' '(' args=arglist? ')' '{' body '}'    # FunDef
   | expr '(' args=explist  ')'                # FunApp
   | expr op=(ADD|SUB) expr                    # AddSub
   | NUMBER                                    # Number
   | IDENT                                     # Ident
   ;

statement
   : ident=IDENT '=' expr           # Assign
   | IF expr '{' lines=line* '}'    # If
   | FOR expr ';' expr ';' expr '{' body '}' # For
   | WHILE expr '{' body '}' # While
   ;