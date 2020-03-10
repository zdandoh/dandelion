parser grammar Dandelion;

options { tokenVocab=DandelionLex; }

start : line+ EOF;

line: (expr|statement) ';';
typeline: typed ident=IDENT ';';
arglist: IDENT (',' IDENT)* (',')?;
typelist: typed? (',' typed)*;
typed
    : IDENT                                 # BaseType
    | 'f' '(' ftypelist=typelist ')' typed  # TypedFun
    | typed '[' ']'                         # TypedArr
    | '(' tuptypes=typelist ')'             # TypedTup
    ;
typedidents: typed IDENT (',' typed IDENT)* (',')?;
explist: expr? (',' expr)*;
body: lines=line*;
structbody: lines=typeline*;

expr
   : '(' expr ')'                                 # ParenExp
   | '[' elems=explist ']'                        # Array
   | '(' elems=explist ')'                        # Tuple
   | expr '.' IDENT                               # StructAccess
   | expr '[' index=expr ']'                      # SliceExp
   | left=expr PIPE right=expr                    # PipeExp
   | expr op=(MUL|DIV) expr                       # MulDiv
   | 'f' '{' body '}'                             # FunDef
   | 'f' '(' args=arglist? ')' '{' body '}'       # FunDef
   | 'f' '(' typedargs=typedidents? ')' returntype=typed '{' body '}' # FunDef
   | 'struct' '{' structbody '}'                  # StructDef
   | 'next' '(' expr ')'                          # NextExp
   | 'send' '(' expr ',' expr ')'                 # SendExp
   | expr '(' args=explist  ')'                   # FunApp
   | expr op=(ADD|SUB) expr                       # AddSub
   | expr MOD expr                                # ModExp
   | expr op=(LT|LTE|GT|GTE|EQ|NEQ) expr          # CompExp
   | FLOAT                                        # FloatExp
   | NUMBER                                       # Number
   | STRING                                       # StrExp
   | BYTE                                         # ByteExp
   | (TRUE|FALSE)                                 # BoolExp
   | NULL                                         # NullExp
   | COMMAND                                      # CommandExp
   | IDENT                                        # Ident
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