package token

type Type int // the individual Type of each token

const (
	// punctuation

	Comma      Type = iota // token for , symbol
	Star                   // token for * symbol
	Slash                  // token for / symbol
	Plus                   // token for + symbol
	Minus                  // token for - symbol
	LeftParen              // token for ( symbol
	RightParen             // token for ) symbol
	Colon                  // token for : symbol
	BackSlash              // token for \ symbol
	NewLine                // token for newline "\n"
	Here                   // token for . symbol

	// literals

	Identifier // token for any identifier, such as a label
	Number     // token for a number

	__directive_start
	// directives

	Macro    // token for .macro
	EndMacro // token for .endmacro
	Global   // token for .global

	// sections

	Boot // token for .boot (not standard, appears at start of .text)
	Text // token for .text
	Data // token for .data
	Bss  // token for .bss
	__directive_end

	// registers

	__reg_start
	R0 // token for r0 register
	R1 // token for r1 register
	R2 // token for r2 register
	R3 // token for r3 register
	__reg_end

	// instructions

	__instruction_start
	Add      // token for add instruction
	Sub      // token for sub instruction
	And      // token for and instruction
	Or       // token for or instruction
	Lsh      // token for lsh instruction
	Rsh      // token for rsh instruction
	Move     // token for move instruction
	St       // token for st instruction
	Ld       // token for ld instruction
	Jump     // token for jump instruction
	Jumpr    // token for jumpr instruction
	Jumps    // token for jumps instruction
	StageRst // token for stage_rst instruction
	StageInc // token for stage_inc instruction
	StageDec // token for stage_dec instruction
	Halt     // token for halt instruction
	Wake     // token for wake instruction
	Sleep    // token for sleep instruction
	Wait     // token for wait instruction
	Adc      // token for adc instruction
	I2cRd    // token for i2c_rd instruction
	I2cWr    // token for i2c_wr instruction
	RegRd    // token for reg_rd instruction
	RegWr    // token for reg_wr instruction
	__instruction_end

	// instruction parameters

	__jump_start
	Ov // token for ov (overflow) parameter
	Eq // token for eq (equals) parameter
	Lt // token for lt (less than) parameter
	Ge // token for ge (greater than or equal) parameter
	__jump_end

	EndOfFile // token for end of file
	Unknown   // some unknown token that isn't a valid identifier
)

var toToken = map[string]Type{
	",":         Comma,
	"*":         Star,
	"/":         Slash,
	"+":         Plus,
	"-":         Minus,
	"(":         LeftParen,
	")":         RightParen,
	":":         Colon,
	"\\":        BackSlash,
	".macro":    Macro,
	".endmacro": EndMacro,
	".global":   Global,
	".":         Here,
	".boot":     Boot,
	".text":     Text,
	".data":     Data,
	".bss":      Bss,
	"r0":        R0,
	"r1":        R1,
	"r2":        R2,
	"r3":        R3,
	"add":       Add,
	"sub":       Sub,
	"and":       And,
	"or":        Or,
	"lsh":       Lsh,
	"rsh":       Rsh,
	"move":      Move,
	"st":        St,
	"ld":        Ld,
	"jump":      Jump,
	"jumpr":     Jumpr,
	"jumps":     Jumps,
	"stage_rst": StageRst,
	"stage_inc": StageInc,
	"stage_dec": StageDec,
	"halt":      Halt,
	"wake":      Wake,
	"sleep":     Sleep,
	"wait":      Wait,
	"adc":       Adc,
	"i2c_rd":    I2cRd,
	"i2c_wr":    I2cWr,
	"reg_rd":    RegRd,
	"reg_wr":    RegWr,
	"ov":        Ov,
	"eq":        Eq,
	"lt":        Lt,
	"ge":        Ge,
}
var toString map[Type]string

func init() {
	toString = make(map[Type]string)
	for s, t := range toToken {
		toString[t] = s
	}
}

func (t Type) String() string {
	val, ok := toString[t]
	if ok {
		return val
	}
	switch t {
	case Identifier:
		return "Identifier"
	case Number:
		return "Number"
	case EndOfFile:
		return "EOF"
	case NewLine:
		return "NewLine"
	default:
		return "UNKNOWN"
	}
}

func ToType(str string) Type {
	val, ok := toToken[str]
	if ok {
		return val
	}
	if str == "\n" {
		return NewLine
	}
	return Unknown
}

func (t Type) IsDirective() bool {
	return t > __directive_start && t < __directive_end
}

func (t Type) IsInstruction() bool {
	return t > __instruction_start && t < __instruction_end
}

func (t Type) IsJump() bool {
	return t > __jump_start && t < __jump_end
}

func (t Type) IsRegister() bool {
	return t > __reg_start && t < __reg_end
}
