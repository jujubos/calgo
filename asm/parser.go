package asm

import (
	"fmt"
)

type Parser struct {
	lexer *Lexer
	tk    Token
}

func NewParser(filename string) *Parser {
	p := &Parser{
		lexer: NewLexer(filename),
	}
	p.move()
	return p
}

func (p *Parser) Error(info string) {
	fname, lnum, cnum := p.lexer.GetPosition()
	panic(fmt.Sprintf("语法错误:%s in (line:%d,column:%d) of %s, p.tk is %s\n", info, lnum, cnum, fname, p.tk.String()))
}

func (p *Parser) Parse() {
	p.program()
	if !p.match(EOF) {
		p.Error("Parse err: 最后不是文件结束符")
	}
	ScanNum++
	if ScanNum <= 2 {
		p.Reset()
		p.Parse()
	}
}

/*
program -> section ID <program>
program -> global ID <program>
program -> ID <lbtail> program>
program -> <inst> program
program -> ^
*/
func (p *Parser) program() {
	if p.match(KW_SEC) {
		if p.tk.TokenTyp() != ID {
			p.Error("section后面必须是标识符")
		}
		name := p.tk.(*TID).Name
		SwitchSeg(name)
		p.move()
		p.program()
	} else if p.match(KW_GLB) {
		if p.tk.TokenTyp() != ID {
			p.Error("global后面必须是标识符")
		}
		name := p.tk.(*TID).Name
		lb := Symtab.GetLb(name)
		lb.Global = true
		p.move()
		p.program()
	} else if p.tk.TokenTyp() == ID { //定义数据
		name := p.tk.(*TID).Name
		p.move()
		p.lbtail(name)
		p.program()
	} else if p.tk.TokenTyp() == EOF {
		SwitchSeg("")
		return
	} else {
		p.inst()
		p.program()
	}
}

/*
lbtail -> :
-> equ NUM
-> times NUM <basetail>
-> <basetail>
*/
func (p *Parser) lbtail(name string) {
	if p.match(COLON) { //标签
		Symtab.AddLb(NewLabel(name, false))
	} else if p.match(KW_EQU) { //宏
		if p.tk.TokenTyp() != NUM {
			p.Error("equ后必须是数值")

		}
		v := p.tk.(*TNUM).Value
		p.move()
		Symtab.AddLb(NewEquLb(name, int(v)))
	} else if p.match(KW_TIMES) { //数组
		if p.tk.TokenTyp() != NUM {
			p.Error("times后必须是数值")
		}
		t := p.tk.(*TNUM).Value
		p.basetail(name, int(t))
	} else { //非数组
		p.basetail(name, 1)
	}
}

/*
inst -> <doubleop> <operand> , <operand>

	-> <singleop> <operand>
	-> <noneop>
*/
func (p *Parser) inst() {
	InstrInit()
	if p.MatchDoubleOpFirst() {
		tktyp := p.tk.TokenTyp()
		p.doubleop()
		regNum, des_t, src_t, len := 0, OP_TYPE(0), OP_TYPE(0), 0
		p.operand(&regNum, &des_t, &len)
		if !p.match(COMMA) {
			p.Error("doubleop err:第二个操作数前缺少,")
		}
		p.operand(&regNum, &src_t, &len)
		Gen2Op(tktyp, des_t, src_t, len)
	} else if p.MatchSingleOpFirst() {
		var opt OP_TYPE
		var regnum = 0
		var l = 0
		tktyp := p.tk.TokenTyp()
		p.singleop()
		p.operand(&regnum, &opt, &l)
		Gen1Op(tktyp, opt, l)
	} else {
		p.noneop()
		Gen0Op(I_RET)
	}
}

func (p *Parser) MatchDoubleOpFirst() bool {
	if _, ok := doubleopfirst[p.tk.TokenTyp()]; ok {
		return true
	}
	return false
}

func (p *Parser) MatchSingleOpFirst() bool {
	if _, ok := singleopfirst[p.tk.TokenTyp()]; ok {
		return true
	}
	return false
}

// basetail -> <len> <value>
func (p *Parser) basetail(name string, t int) {
	l := p.len()
	p.value(name, t, l)
}

// doubleop -> ...
func (p *Parser) doubleop() {
	if p.MatchDoubleOpFirst() {
		p.move()
	}
}

// singleop -> ...
func (p *Parser) singleop() {
	if p.MatchSingleOpFirst() {
		p.move()
	}
}

// noneop -> ret
func (p *Parser) noneop() {
	p.match(I_RET)
}

// operand -> NUM | ID | <reg> | <mem>
func (p *Parser) operand(regnum *int, opt *OP_TYPE, l *int) {
	tktyp := p.tk.TokenTyp()
	if tktyp == NUM {
		*opt = IMMEDIATE
		Instr.Imm32 = int(p.tk.(*TNUM).Value)
		p.move()
	} else if tktyp == ID {
		*opt = IMMEDIATE
		lb := Symtab.GetLb(p.tk.(*TID).Name)
		Instr.Imm32 = lb.Addr
		if ScanNum == 2 && !lb.IsEqu {
			RelLb = lb
		}
		p.move()
	} else if tktyp == LBRACK {
		*opt = MEMORY
		p.mem()
	} else {
		tktyp, *l = p.reg()
		*opt = REGISTER
		regcode := GetRegCode(tktyp, *l)
		if *regnum == 0 {
			MODRM.Reg = regcode
		} else { //双寄存器
			MODRM.Mod = 3
			MODRM.RM = regcode
		}
		*regnum = *regnum + 1
	}
}

func GetRegCode(reg TokenType, l int) int {
	if l == 1 {
		return int(reg - BR_AL)
	}
	return int(reg - DR_EAX)
}

// len -> db | dw | dd
func (p *Parser) len() int {
	if p.match(KW_DB) {
		return 1
	} else if p.match(KW_DW) {
		return 2
	} else if p.match(KW_DD) {
		return 4
	}
	p.Error("len err: 只能是db, dw或者dd")
	return 0
}

// value -> <type> <valtail>
func (p *Parser) value(name string, t int, l int) {
	vs := []int{}
	p.typ(&vs)
	p.valtail(&vs)
	//回溯
	Symtab.AddLb(NewDataLb(name, t, l, vs))
}

// reg -> ...
func (p *Parser) reg() (r TokenType, l int) {
	if p.MatchRegFirst() {
		l = 4
		if p.tk.TokenTyp() <= BR_BH && p.tk.TokenTyp() >= BR_AL {
			l = 1
		}
		r = p.tk.TokenTyp()
		p.move()
	}
	return
}

// mem -> [ <addr> ]
func (p *Parser) mem() {
	if p.match(LBRACK) {
		p.addr()
		if !p.match(RBRACK) {
			p.Error("mem err:缺少]")
		}
	}
}

/*
type -> NUM

	-> <off> NUM
	-> STR
	-> ID
*/
func (p *Parser) typ(vs *[]int) {
	switch p.tk.TokenTyp() {
	case NUM:
		v := p.tk.(*TNUM).Value
		*vs = append(*vs, int(v))
		p.move()
	case STR:
		v := p.tk.(*TSTR).Value
		for _, b := range []byte(v) {
			*vs = append(*vs, int(b))
		}
		p.move()
	case ID:
		l := Symtab.GetLb(p.tk.(*TID).Name)
		*vs = append(*vs, l.Addr)
		if ScanNum == 2 && !l.IsEqu {
			ELFOBJ.AddRel(CurSeg, CurAddr, l.Name, R_386_32)
		}
		p.move()
	default:
		neg := p.off()
		if p.tk.TokenTyp() != NUM {
			p.Error("typ err: <off>后必须是数值")
		}
		v := p.tk.(*TNUM).Value
		if neg {
			v = -v
		}
		*vs = append(*vs, int(v))
		p.move()
	}
}

/*
valtail -> , <type> <valtail> |  ^
*/
func (p *Parser) valtail(vs *[]int) {
	if p.match(COMMA) {
		p.typ(vs)
		p.valtail(vs)
	}
}

/*
addr -> NUM

	-> ID
	-> <reg> <regaddr>
*/
func (p *Parser) addr() {
	tktyp := p.tk.TokenTyp()
	if tktyp == NUM { //直接寻址
		MODRM.Mod = 0
		MODRM.RM = 5
		Instr.Disp = int(p.tk.(*TNUM).Value)
		Instr.Displen = 4
		p.move()
	} else if tktyp == ID { //直接寻址
		MODRM.Mod = 0
		MODRM.RM = 5
		lb := Symtab.GetLb(p.tk.(*TID).Name)
		Instr.Disp = lb.Addr
		Instr.Displen = 4
		if ScanNum == 2 && !lb.IsEqu {
			RelLb = lb
		}
		p.move()
	} else { //寄存器寻址
		regtk, l := p.reg()
		p.regaddr(regtk, l)
	}
}

// off -> + | -
func (p *Parser) off() bool {
	if p.match(ADD) {
		return false
	} else if p.match(SUB) {
		return true
	} else {
		p.Error("off err: 只能是+或者-")
	}
	return false
}

// regaddr -> <off> <regaddrtail> | ^
func (p *Parser) regaddr(regtk TokenType, l int) {
	if p.tk.TokenTyp() == ADD || p.tk.TokenTyp() == SUB { //
		sign := p.tk.TokenTyp()
		p.off()
		p.regaddrtail(regtk, l, sign)
	} else { //寄存器间址
		basereg := regtk
		//当mod = 00, r/m = 100时(esp的寄存器编码就是100)，表示引导SIB字段
		//所以本来mod = 00, r/m = 100的含义: 利用esp间接寻址，被覆盖
		//不过可以使用SIB字段来表示原来的含义
		if basereg == DR_ESP { //引导SIB
			MODRM.Mod = 0
			MODRM.RM = 4
			SIBP.Base = 4
			SIBP.Index = 4 //index = 100表示不存在变址寄存器
			SIBP.Scale = 0
			//mod = 00, r/m = 101时(ebp的寄存器编码就是101)，表示立即数直接寻址。
			//原来的含义：利用ebp间接寻址，被覆盖
			//不过可以使用mod = 01, r/m = 101, 含义为寄存器基址 + 8位偏移，即[ebp + 0]
		} else if basereg == DR_EBP { //寄存器基址+8位偏移
			MODRM.Mod = 1
			MODRM.RM = 5
			Instr.SetDisp(0, 1)
		} else {
			MODRM.Mod = 0
			MODRM.RM = int(basereg - BR_AL)
			if l == 4 {
				MODRM.RM = int(basereg - DR_EAX)
			}
		}
	}
}

// regaddrtail -> NUM | <reg>
func (p *Parser) regaddrtail(regtk TokenType, l int, sign TokenType) {
	basereg := regtk
	if p.tk.TokenTyp() == NUM { //寄存器基址 + 偏移
		num := int(p.tk.(*TNUM).Value)
		if sign == SUB {
			num = -num
		}
		if num >= -128 && num <= 127 {
			MODRM.Mod = 1
			Instr.SetDisp(num, 1)
		} else {
			MODRM.Mod = 2
			Instr.SetDisp(num, 4)
		}
		MODRM.RM = int(basereg-BR_AL) - (1-l%4)*8
		if basereg == DR_ESP { //[esp + 0x...]
			MODRM.RM = 4
			SIBP.Base = 4
			SIBP.Index = 4 //不存在变址寄存器
			SIBP.Scale = 0
		}
		p.move()
	} else { //基址寄存器 + 变址寄存器。无偏移，生成的汇编没有基址 + 变址 + 偏移这种。
		idxreg, il := p.reg()
		MODRM.Mod = 0
		MODRM.RM = 4
		SIBP.Base = int(basereg-BR_AL) - (1-l%4)*8
		SIBP.Index = int(idxreg-BR_AL) - (1-il%4)*8
	}
}

func (p *Parser) MatchRegFirst() bool {
	if _, ok := regfirst[p.tk.TokenTyp()]; ok {
		return true
	}
	return false
}

var doubleopfirst = map[TokenType]struct{}{
	I_MOV: {},
	I_CMP: {},
	I_SUB: {},
	I_ADD: {},
	I_AND: {},
	I_OR:  {},
	I_LEA: {},
}

var singleopfirst = map[TokenType]struct{}{
	I_CALL:  {},
	I_INT:   {},
	I_IMUL:  {},
	I_IDIV:  {},
	I_NEG:   {},
	I_INC:   {},
	I_DEC:   {},
	I_JMP:   {},
	I_JE:    {},
	I_JNE:   {},
	I_SETE:  {},
	I_SETNE: {},
	I_SETG:  {},
	I_SETGE: {},
	I_SETL:  {},
	I_SETLE: {},
	I_PUSH:  {},
	I_POP:   {},
}

var regfirst = map[TokenType]struct{}{
	BR_AL:  {},
	BR_CL:  {},
	BR_DL:  {},
	BR_BL:  {},
	BR_AH:  {},
	BR_CH:  {},
	BR_DH:  {},
	BR_BH:  {},
	DR_EAX: {},
	DR_ECX: {},
	DR_EDX: {},
	DR_EBX: {},
	DR_ESP: {},
	DR_EBP: {},
	DR_ESI: {},
	DR_EDI: {},
}

func (p *Parser) match(typ TokenType) bool {
	if p.tk.TokenTyp() == typ {
		p.move()
		return true
	}
	return false
}

func (p *Parser) move() {
	p.tk = p.lexer.NextToken()
}

func (p *Parser) Reset() {
	p.lexer.Reset()
	p.tk = nil
	p.move()
}
