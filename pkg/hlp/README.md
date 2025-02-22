
# hlp

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

hlp is a language specification and compiler. It is intended as an intermediate language for the ESP32 ULP coprocessor. It targets the ulp-asm assembler.

The language is intended to be simple to optimize. Therefore it is all three-address code, plus function calls.

All integer types are unsigned 16-bit integers. They are not declared.

Arrays must be declared, and can be declared such as `x @ 3;` for an array named "x" with 3 integers. "x" cannot be accessed directly, you must state the offset such as `x#0` to access the 0-th element. C-style structs can be constructed with these. You can obtain the address with "&" such as `addr = &x#0`. Arrays are guaranteed to align in memory after optimizations.

Functions are declared with `func test_function(a, b, c @ 2) 3 {}` where `test_function` is the name of the function, `a` and `b` are integer inputs, `c` is an array input, and there are 3 outputs. All inputs must be set manually and all outputs must be set, so it could be called with `x, y, z = test_function(0, 1, 2, 3);`

Hardware access instructions, such as `halt()`, are supported. Note that all inputs must be integers.

Global variables are supported. Any global variables marked with `static` will not be included in the assembly if it is not called in a function. All non-static global variables will be marked with `.global` in the assembly and will therefore be accessible by the ESP32. Static variables inside functions are not supported.

Example blink program (register writes are removed):
```
func gpio_init() 0 {
    // set up the desired gpio
}

func gpio_high() 0 {
    reg_wr(0, 0, 0, 0);
}

func gpio_low() 0 {
    reg_wr(0, 0, 0, 0);
}

func delay_ms(milliseconds) 0 {
loop_start:
    if milliseconds == 0 goto loop_end
    wait(8000); // this is close to 1 millisecond delay
    milliseconds = milliseconds - 1;
    goto loop_start;
loop_end:
}

noreturn func main() 0 {
    milliseconds = 1000;
    gpio_init();
loop:
    gpio_high();
    delay_ms(milliseconds);
    gpio_low();
    delay_ms(milliseconds);
    goto loop;
}
```

# Optimizations

* [ ] Non-moving stack
* [ ] Unused function and variable elimination
* [ ] Common subexpression elimination
* [ ] Constant folding and propagation
* [ ] Peephole optimization
* [ ] Register allocation
* [ ] Branch elimination
* [ ] Dead code elimination
* [ ] Function inlining
* [ ] Tail call optimization

# ABI

## Register Usage

`r0` is used to pass values within functions, see below. `r1` can be used without restriction. `r2` is used to hold the return address of a function, otherwise it can be used freely. `r3` is used as a stack pointer and should not be used for anything else.

## Function Calls

Functions can be declared in any order (a function can call one defined later). Functions can be called in a different file.

All function calls must have the return address on `r2`. Parameters and return values are placed in memory sequentially, with return values first. So a function `test(a, b, c@2) 3` will have the memory layout:
```
[return#0, return#1, return#2, a, b, c#0, c#1]
```
Whereas a function with no inputs `noinputs(a, b, c@2)` will have the memory layout:
```
[a, b, c#0, c#1]
```

The left-most value in memory will be passed/returned in the `r0` register. If there are no returns or arguments, `r0` is callee saved.

Registers `r1` and `r2` are callee saved. Register `r3` is used as a stack pointer and should therefore not be modified. When returning, the stack should be the same depth as when the function is called.

Each function call assumes it is at the top of the stack. So a function:
```
func demo(a) 2 {
    return a+1, 2;
}
```
will have the two returns at `r0` and `r3[0]`, parameter `a` will reside at `r3[1]`.

A function can be modified with `noreturn` which will tell the compiler to not save registers and to not return.

Inline assembly is not supported, but functions can be assembly. They are declared the same as regular functions but with `__asm__ func`. Each assembly statement should be in a string followed by a semicolon. This is the demo code above as inline assembly:
```
__asm__ func demo(a) 2 {
    "mv r0, 2";
    "st r0, r3, 0"; // store 2 as return[1]

    "ld r0, r3, 1"; // load param a into r0
    "add r0, r0, 1"; // a = a+1
    // a is returned on r0

    "jump r2"; // return
}
```

# Boot

A minimal boot function that sets up the stack and jumps to `main()` is provided by the compiler. It can be overwritten by defining a function named `__boot()`. It will be placed in section `.boot`, which is at address 0. This function should be assembly and jump to `main`.

# Grammar

Comments are done with `//`.
```
binary_ops
    : "+"
    | "-"
    | "|"
    | "&"
    | "<<"
    | ">>"

var
    : ident "#" NUMBER

variable_declaration
    : "var" var_def ";"

primary
    : NUMBER
    | "&"? var
    | "&" ident // address of a function

right_ops_expr
    : primary binary_ops primary
    | primary
    | primary "[" NUMBER "]"

function_inputs
    : ( primary? ( "," primary? )* )?

function_outputs
    : var ( "," var )*

function_call
    : ident "(" function_inputs ")"

label
    : ident ":"

ops_expr
    : var "=" right_ops_expr ";"

store_expr
    : ident "[" NUMBER "]" "=" primary ";"

function_expr
    : function_outputs "=" function_call ";"
    | function_call ";"

return_expr
    : "return" (primary ( "," primary )* )? ";"

compare_ops
    : ">"
    | ">="
    | "<"
    | "<="
    | "=="
    | "!="

jump_expr
    : "goto" label
    | "if" primary compare_ops primary "goto" label
    | "ifOv" var "=" primary binary_ops primary "goto" label

var_def
    : ident "@" NUMBER

function_def_input
    : ( var_def ( "," var_def )* )?

reg_wr_expr
    : "reg_wr" "(" NUMBER "," NUMBER "," NUMBER "," NUMBER ")" ";"

reg_rd_expr
    : var "=" "reg_rd" "(" NUMBER "," NUMBER "," NUMBER ")" ";"

wait_expr
    : "wait" "(" NUMBER ")" ";"

i2c_wr_expr
    : "i2c_wr" "(" NUMBER "," NUMBER "," NUMBER "," NUMBER "," NUMBER ")" ";"

i2c_rd_expr
    : var "=" "i2c_rd" "(" NUMBER "," NUMBER "," NUMBER "," NUMBER ")" ";"

halt_expr
    : "halt" "(" ")" ";"

wake_expr
    : "wake" "(" ")" ";"

sleep_expr
    : "sleep" "(" NUMBER ")" ";"

adc_expr
    : var "=" "adc" "(" NUMBER "," NUMBER "," NUMBER ")" ";"

hardware_expr
    : reg_wr_expr
    | reg_rd_expr
    | wait_expr
    | i2c_wr_expr
    | i2c_rd_expr
    | halt_expr
    | wake_expr
    | sleep_expr
    | adc_expr

empty
    : ";" // ignored

function_statements
    : ops_expr
    | store_expr
    | function_expr
    | jump_expr
    | return_expr
    | hardware_expr
    | label
    | variable_declaration
    | STRING // only for assembly statements
    | empty

asm_statements
    : ( string ";" )*

function_attribute_list
    : "assembly"
    | "weak"
    | "require" "=" "(" ident ( "," ident )* ")"

function_attribute
    : "(" function_attribute_list ")"

function_attributes
    : "__attribute__" "(" function_attribute ( "," function_attribute )* ")"

function_declaration
    : "static"? "func" ident 
        "(" function_def_input ")" // define the inputs
        NUMBER // number of outputs
        function_attributes? // optionally list the attributes
        "{" function_statements* "}"

global_variable
    : "extern" variable_declaration
    | "static"? variable_declaration
    | "static"? "var_set" var_def "=" primary ( "," primary )* ";"

static_statement
    : function_declaration
    | global_variable
    | empty

program: static_statement* EOF
```