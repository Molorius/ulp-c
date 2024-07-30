
Grammar:

```
builtin : "__bss_start" | "__bss_end"
ident   : [_.a-zA-Z0-9]*
label   : ident ":"
section : ".boot" | ".text" | ".data" | ".bss"
global  : ".global" ident

primary     : NUMBER | "." | builtin | ident | "(" expression ")"
unary       : "-" unary
            | primary
factor      : unary ( ( "/" | "*" ) unary )*
expression  : factor ( ( "-" | "+" ) factor )*
reg         : "r0" | "r1" | "r2" | "r3"
any         : ( reg | primary )

jump_cond  : "ov" | "eq"
jumpr_cond : "lt" | "ge"
jumps_cond : "le" | "lt" | "ge"

param0 : reg "," reg "," any
param1 : reg "," any
param2 : reg "," primary
param3 : any ( "," jump_cond )?
param4 : primary "," primary "," jumpr_cond
param5 : primary "," primary "," jumps_cond
param6 : primary
param7 : reg "," primary "," primary
param8 : primary "," primary "," primary "," primary
param9 : primary "," primary "," primary "," primary "," primary
param10: primary "," primary "," primary

ins0     : "add" | "sub" | "and" | "or" | "lsh" | "rsh"
ins1     : "move"
ins2     : "st" | "ld"
ins3     : "jump"
ins4     : "jumpr"
ins5     : "jumps"
ins6     : "stage_inc" | "stage_dec" | "sleep" | "wait"
ins7     : "adc"
ins8     : "i2c_rd" | "reg_wr"
ins9     : "i2c_wr"
ins10    : "reg_rd"
ins_none : "stage_rst" | "halt" | "wake"

ins     : ins0 param0
        | ins1 param1
        | ins2 param2
        | ins3 param3
        | ins4 param4
        | ins5 param5
        | ins6 param6
        | ins7 param7
        | ins8 param8
        | ins9 param9
        | ins10 param10
        | ins_none

start: ( section | label | global | ins )*

```

All instructions except jump* use 'reg' or 'expression' as inputs
