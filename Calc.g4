parser grammar Calc;

options { tokenVocab=CalcLex; }

start : line+ EOF;

line: (expr|statement) ';';
typeline: typename=IDENT ident=IDENT ';';
arglist: IDENT (',' IDENT)* (',')?;
explist: expr? (',' expr)*;
body: lines=line*;
structbody: lines=typeline*;

expr
   : '(' expr ')'                              # ParenExp
   | '[' elems=explist ']'                     # Array
   | expr '.' IDENT                            # StructAccess
   | expr '[' index=expr ']'                   # SliceExp
   | left=expr PIPE right=expr                 # PipeExp
   | expr op=(MUL|DIV) expr                    # MulDiv
   | 'f' '{' body '}'                          # FunDef
   | 'f' '(' args=arglist? ')' '{' body '}'    # FunDef
   | 'struct' '{' structbody '}'               # StructDef
   | expr '(' args=explist  ')'                # FunApp
   | expr op=(ADD|SUB) expr                    # AddSub
   | expr MOD expr                             # ModExp
   | expr op=(LT|LTE|GT|GTE|EQ) expr           # CompExp
   | NUMBER                                    # Number
   | STRING                                    # StrExp
   | COMMAND                                   # CommandExp
   | IDENT                                     # Ident
   ;

statement
   : expr '=' expr                           # Assign
   | IF expr '{' lines=line* '}'             # If
   | 'struct' ident=IDENT '{' structbody '}' # NamedStructDef
   | FOR expr ';' expr ';' expr '{' body '}' # For
   | WHILE expr '{' body '}'                 # While
   | RETURN expr                             # Return
   | YIELD expr                              # Yield
   ;