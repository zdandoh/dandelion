parser grammar Dandelion;

options { tokenVocab=DandelionLex; }

start : line+ EOF;

line: (expr|statement) ';';
code: (expr|statement);
typeline: typed ident=IDENT ';';
arglist: IDENT (',' IDENT)* (',')?;
typelist: typed? (',' typed)*;
typed
    : IDENT                                 # BaseType
    | 'f' '(' ftypelist=typelist ')' typed  # TypedFun
    | '[' ']' typed                         # TypedArr
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
   | expr '.' '(' typed ')'                       # TypeAssert
   | expr 'is' typed                              # IsExp
   | left=expr PIPE right=expr                    # PipeExp
   | expr op=(MUL|DIV) expr                       # MulDiv
   | FMODS '{' body '}'                           # FunDef
   | FSTART '(' args=arglist? ')' '{' body '}'    # FunDef
   | FSTART '(' typedargs=typedidents? ')' returntype=typed '{' body '}' # FunDef
   | 'struct' '{' structbody '}'                  # StructDef
   | bname=(LEN|DONE|NEXT|SEND|ANY|TYPE) '(' args=explist ')' # BuiltinExp
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
   | idtype=typed id=IDENT                        # Ident
   | id=IDENT                                     # Ident
   ;

statement
   : expr '=' expr                           # Assign
   | IF expr '{' lines=line* '}'             # If
   | 'struct' ident=IDENT '{' structbody '}' # NamedStructDef
   | FOR iname=IDENT 'in' expr '{' body '}'  # ForIter
   | FOR code ';' expr ';' code '{' body '}' # For
   | WHILE expr '{' body '}'                 # While
   | ('break' | 'continue')                  # FlowControl
   | '{' body '}'                            # BlockExp
   | RETURN expr                             # Return
   | YIELD expr                              # Yield
   ;