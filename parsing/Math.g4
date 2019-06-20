grammar Math;

prog: line+ ;

line: expr NEWLINE
    | NEWLINE
    ;

expr: expr ('*'|'/') expr   # MulDiv
    | expr ('+'|'-') expr   # AddSub
    | INT                   # int
    | '(' expr ')'          # parens
    ;

INT: [0-9]+ ;
NEWLINE: '\r'? '\n' ;
WS: [ \t]+ -> skip ;