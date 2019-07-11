grammar Math;

prog: line+ ;

line: expr NEWLINE
    | NEWLINE
    ;

expr: expr ('+'|'-'|'*'|'/') expr # BinOp
    | INT                         # int
    | '(' expr ')'                # parens
    ;

INT: [0-9]+ ;
NEWLINE: '\r'? '\n' ;
WS: [ \t]+ -> skip ;