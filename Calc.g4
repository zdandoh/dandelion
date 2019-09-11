parser grammar Calc;

options { tokenVocab=CalcLex; }

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
   | STRING                                    # StrExp
   | IDENT                                     # Ident
   ;

statement
   : ident=IDENT '=' expr           # Assign
   | IF expr '{' lines=line* '}'    # If
   | FOR expr ';' expr ';' expr '{' body '}' # For
   | WHILE expr '{' body '}' # While
   ;