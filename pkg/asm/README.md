
# ulp-asm

ulp-asm is an assembler for the ESP32 ULP coprocessor.
Note that this converts assembly directly into the final binary.

ulp-asm currently only supports one file.

ulp-asm currently only supports the original ESP32.

# Directives

* `.global symbol`
* `.int`
* `.boot`, code here will be placed at the start of the .text section
* `.boot.data`, code here will be placed at the start of the .data section
* `.text`
* `.data`
* `.bss`

# Instructions
* add
* sub
* and
* or
* move
* lsh
* rsh
* stage_rst
* stage_inc
* stage_dec
* st
* ld
* jump
* jumpr
* jumps
* halt
* wake
* sleep
* wait

# Comments

The following comment types are supported:
```
// I am a comment until end of line
# I am also a comment until end of line
/* I am an inline comment */
```

# Sections

The ULP binary contains a header with information about the `.text`, `.data`,
and `.bss` sections. These will be referred to as `.header.text`, `.header.data`, and `.header.bss` respectively in this document.

This assembler has the `.boot` and `.text` directives. These will be added to the `.header.text` section in order. The label "__boot_start" is put at the start of the `.boot` section, "__boot_end" at the end. The label "__text_start" is put at the start of the `.text` section, "__text_end" at the end.

This assembler has the `.boot.data` and `.data` directives. These will be added to the `.header.data` section in order. The label "__boot_data_start" is put at the start of the `.boot.data` section, "__boot_data_end" at the end. The label "__data_start" is put at the start of the `.data` section, "__data_end" at the end.

This assembler uses the `.bss` directive to put data in the `.header.bss` section. The label "__bss_start" is placed at the start, "__bss_end" at the end.

This assembler allocates the remainder of the reserved space for a stack at the end of the `.header.bss` section. The label "__stack_start" is placed at the start, "__stack_end" at the end.

# Differences

There are several differences between ulp-asm and esp32ulp-elf-as.

## Number labels

esp32ulp-elf-as can use number labels such as the following:
```
  0:
add r0, r0, 10
jump 0b, ov
```
Where the `jump` will go back to the nearest `0` label on overload.

ulp-asm does not currently support this.


## Case Sensitivity

esp32ulp-elf-as is case insensitive: it allows all instructions to be upper case or lower case.

ulp-asm is case sensitive: all instructions are lower case.

## Math with labels

Note that this applies to any expression.

esp32ulp-elf-as divides the final output of any expression involving a label
by 4. For example, the expression `entry+3` will compile to the same as the
expression `entry` because the extra addition is truncated. Additionally, it
will only allow one label per expression.

ulp-asm does not divide the final output, so `entry` and `entry+3` will compile
to different values.

## ld instruction

esp32ulp-elf-as divides the offset by 4, ulp-asm does not.

esp32ulp-elf-as:
```
ld r0, r0, 0
ld r1, r1, 4
ld r2, r2, 5
ld r3, r3, 20
```

equivalent ulp-asm:
```
ld r0, r0, 0
ld r1, r1, 1
ld r2, r2, 1
ld r3, r3, 5
```

## st instruction
esp32ulp-elf-as divides the offset by 4, ulp-asm does not.

esp32ulp-elf-as:
```
ld r0, r0, 0
ld r1, r1, -4
ld r2, r2, -1
```

equivalent ulp-asm:
```
ld r0, r0, 0
ld r1, r1, -1
ld r2, r2, 0
```

## jumpr instruction

Note that jumpr is an unsigned comparison.

-----

esp32ulp-elf-as actually uses a different value for the step depending on
whether you use a label for the step parameter or not.

Given the following code:
```
    test:
jumpr test, 1, lt
jumpr 10*4, 2, lt // esp32ulp-elf-as likes to divide by 4...
```

esp32ulp-elf-as will attempt to jump back by 1, then forward by 10.
This is inconsistent because the label is a hard address, not a relative offset.

ulp-asm is consistent by instead taking in a hard address and converting it
to a relative offset internally. So the equivalent code would be:
```
    test:
jumpr test, 1, lt
jumpr . + 10, 2, lt
```

-----

The threshold parameter is not calculated correctly by esp32ulp-elf-as with large numbers. For example, the following does not compile:
```
jumpr step, 0xFFFF, lt
```
but the equivalent does:
```
jumpr step, -1, lt
```

ulp-asm fixes this.

------

The step parameter throws incorrect errors when the address is high
in a given assembly file. For example, say the following is found
at the end of a large assembly file:
```
  test:
jumpr test, 1, lt
```

esp32ulp-elf-as will say that the step is too large, even though it's clearly stepping back one instruction. ulp-asm fixes this.

----

esp32ulp-elf-as will compile the less than or equal condition `jumpr target, threshold, le` to:
```
jumpr target, threshold+1, lt
```

Which is fine unless threshold == 0xFFFF. ulp-asm does the same unless threshold == 0xFFFF, in which case it will use:
```
jumpr target, 0, ge // always jump
```

-----

esp32ulp-elf-as will compile the greater than condition `jumpr target, threshold, gt` to:
```
jumpr target, threshold+1, ge
```

Which is fine unless threshold == 0xFFFF. ulp-asm does the same unless threshold == 0xFFFF, in which case it will use:
```
jumpr target, 0, lt // never jump
```

------

esp32ulp-elf-as will compile the equals condition `jumpr target, threshold, eq` to:
```
jumpr . + 8, threshold+1, ge // esp32ulp-elf-as likes to divide by 4...
jumpr target, threshold, ge
```

Which is fine unless threshold == 0xFFFF. ulp-asm does the same unless threshold == 0xFFFF, in which case it will use:
```
jumpr . + 2, 0xFFFF, lt
jumpr target, 0xFFFF, ge
```

Note that ulp-asm could instead use:
```
jumpr target, 0xFFFF, ge
jumpr target, 0xFFFF, ge
```
But the first fix has a higher probability on average of executing 1 instruction.

Also note that it uses 2 instructions because it needs to calculate the address of labels
before doing math, which the threshold may use.

## jumps instruction

Note that jumps is an unsigned comparison.

Also note that the esp32ulp-elf-as jumps implementation uses the
same inconsistent address vs offset as its jumpr implementation.
ulp-asm takes in a hard address, the same as its jumpr implementation.

-----

esp32ulp-elf-as will compile the equals condition `jumps target, threshold, gt` to:
```
jumps . + 8, threshold, le // esp32ulp-elf-as likes to divide by 4...
jumps target, threshold, ge
```

This is correct but the same logic can be done in one instruction. ulp-asm will compile this to:
```
jumps target, threshold+1, ge
```
unless threshold == 0xFF, instead it will use:
```
jumps target, 0, lt // never jump
```


# Grammar

The following is the grammar implemented:

```
ident   : [_.a-zA-Z0-9]*
label   : ident ":"
section : ".boot" | ".boot.data" | ".text" | ".data" | ".bss"
global  : ".global" ident
int : ".int" primary ( "," primary )*
directive : ( section | global | int )
newline : "\n"
splitter : newline | EOF

primary     : NUMBER | "." | ident | "(" expression ")"
unary       : "-" unary
            | primary
factor      : unary ( ( "/" | "*" ) unary )*
expression  : factor ( ( "-" | "+" ) factor )*
reg         : "r0" | "r1" | "r2" | "r3"
any         : ( reg | primary )

jump_cond  : "ov" | "eq"
jumpr_cond : "lt" | "le" | "gt" | "ge"
jumps_cond : "lt" | "le" | "gt" | "ge"

param0 : reg "," reg "," any
param1 : reg "," any
param2 : reg "," reg "," primary
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

statement : directive splitter
          | ins splitter
          | label

program: statement* EOF
```
