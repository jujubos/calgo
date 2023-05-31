package asm

import (
	"log"
	"os"
	"unsafe"
)

type OP_TYPE int

const (
	NONE      OP_TYPE = 0
	IMMEDIATE OP_TYPE = 1
	REGISTER  OP_TYPE = 2
	MEMORY    OP_TYPE = 3
)

// r,r | r,m | m,r | r,imm32
var i_2opcode = [7][2][4]int{
	{{0x8a, 0x8a, 0x88, 0xb0}, {0x8b, 0x8b, 0x89, 0xb8}},
	{{0x3a, 0x3a, 0x38, 0x80}, {0x3b, 0x3b, 0x39, 0x81}},
	{{0x2a, 0x2a, 0x28, 0x80}, {0x2b, 0x2b, 0x29, 0x81}},
	{{0x02, 0x02, 0x00, 0x80}, {0x03, 0x03, 0x01, 0x81}},
	{{0x22, 0x22, 0x20, 0x80}, {0x23, 0x23, 0x21, 0x81}},
	{{0x0a, 0x0a, 0x08, 0x80}, {0x0b, 0x0b, 0x09, 0x81}},
	{{0x00, 0x00, 0x00, 0x00}, {0x8d, 0x8d, 0x00, 0x00}},
}

func Gen2Op(tktyp TokenType, des_t OP_TYPE, src_t OP_TYPE, l int) {
	opcode := GetOpCode(tktyp, des_t, src_t, l)
	switch MODRM.Mod {
	case -1:
		if tktyp == I_MOV {
			opcode += MODRM.Reg
		} else {
			regcodes := []int{7, 5, 0, 4, 1}
			MODRM.Mod = 3
			MODRM.RM = MODRM.Reg //TODO:？？？
			MODRM.Reg = regcodes[tktyp-I_CMP]
		}
		WriteBytes(opcode, 1)
		if tktyp != I_MOV {
			WriteModRM()
		}
		ProcessRel(R_386_32) //TODO:???
		WriteBytes(Instr.Imm32, l)
	case 0:
		WriteBytes(opcode, 1)
		WriteModRM()
		if MODRM.RM == 5 {
			ProcessRel(R_386_32)
			Instr.WriteDisp()
		} else if MODRM.RM == 4 {
			WriteSIB()
		}
	case 1:
		WriteBytes(opcode, 1)
		WriteModRM()
		if MODRM.RM == 4 {
			WriteSIB()
		}
		Instr.WriteDisp()
	case 2:
		WriteBytes(opcode, 1)
		WriteModRM()
		if MODRM.RM == 4 {
			WriteSIB()
		}
		Instr.WriteDisp()
	case 3:
		WriteBytes(opcode, 1)
		WriteModRM()
	}
}

var i_1opcode = [...]int{
	//call,int,imul,idiv,neg,inc,dec,jmp
	0xe8, 0xcd, 0xf7, 0xf7, 0xf7, 0x40, 0x48, 0xe9,
	//je, jne
	0x84, 0x85,
	//sete, setne, setg, setge, setl, setle
	0x94, 0x95, 0x9f, 0x9d, 0x9c, 0x9e,
	//push, pop
	0x50, 0x58,
}

func Gen1Op(tktyp TokenType, opt OP_TYPE, l int) {
	opcode := i_1opcode[tktyp-I_CALL]
	if tktyp == I_CALL || tktyp >= I_JMP && tktyp <= I_JNE {
		if tktyp != I_CALL && tktyp != I_JMP {
			WriteBytes(0x0f, 1)
		}
		WriteBytes(opcode, 1)
		addr := Instr.Imm32
		if ProcessRel(R_386_PC32) {
			addr = CurAddr
		}
		pc := CurAddr + 4
		WriteBytes(addr-pc, 4)
	} else if tktyp >= I_SETE && tktyp <= I_SETLE {
		MODRM.Mod = 3
		MODRM.RM = MODRM.Reg
		MODRM.Reg = 0
		WriteBytes(0x0f, 1)
		WriteBytes(opcode, 1)
		WriteModRM()
	} else if tktyp == I_INT {
		WriteBytes(opcode, 1)
		WriteBytes(Instr.Imm32, 1)
	} else if tktyp == I_PUSH {
		if opt == IMMEDIATE {
			opcode = 0x68
		} else {
			opcode += MODRM.Reg
		}
		WriteBytes(opcode, 1)
		if opt == IMMEDIATE {
			WriteBytes(Instr.Imm32, 4)
		}
	} else if tktyp == I_INC || tktyp == I_DEC {
		if l == 1 { //r8
			opcode = 0xfe
			regcodes := []int{0, 1}
			MODRM.Mod = 3
			MODRM.RM = MODRM.Reg
			MODRM.Reg = regcodes[tktyp-I_INT]
		} else { //r32
			opcode += MODRM.Reg
		}
		WriteBytes(opcode, 1)
		if l == 1 {
			WriteModRM()
		}
	} else if tktyp == I_NEG {
		if l == 1 {
			opcode = 0xf6
		}
		MODRM.Mod = 3
		MODRM.RM = MODRM.Reg
		MODRM.Reg = 3
		WriteBytes(opcode, 1)
		WriteModRM()
	} else if tktyp == I_POP {
		opcode += MODRM.Reg
		WriteBytes(opcode, 1)
	} else if tktyp == I_IMUL || tktyp == I_IDIV {
		regcodes := []int{5, 7}
		MODRM.Mod = 3
		MODRM.RM = MODRM.Reg
		MODRM.Reg = regcodes[tktyp-I_IMUL]
		WriteBytes(opcode, 1)
		WriteModRM()
	}
}

var i_0opcode = [...]int{0xc3}

func Gen0Op(tktyp TokenType) {
	opcode := i_0opcode[tktyp-I_RET]
	WriteBytes(opcode, 1)
}

func WriteBytes(v int, l int) {
	p := (*[4]byte)(unsafe.Pointer(&v))
	b := make([]byte, l)
	copy(b[0:], (*p)[0:])
	TmpCodeSeg.Write(b[:])
	CurAddr += l
}

func WriteModRM() {
	if MODRM.Mod != -1 {
		b := (MODRM.Mod << 6) + (MODRM.Reg << 3) + MODRM.RM
		WriteBytes(b, 1)
	}
}

func WriteSIB() {
	if SIBP.Scale != -1 {
		b := (SIBP.Scale << 6) + (SIBP.Index << 3) + SIBP.Base
		WriteBytes(b, 1)
	}
}

func (i *Inst) WriteDisp() {
	if i.Displen != 0 {
		WriteBytes(i.Disp, i.Displen)
		i.Displen = 0
	}
}

func GetOpCode(op TokenType, des_t, src_t OP_TYPE, l int) int {
	var i1, i2, i3 int
	i1 = int(op - I_MOV)
	if l == 1 {
		i2 = 0
	} else {
		i2 = 1
	}
	if src_t == MEMORY {
		i3 = 1
	} else if src_t == IMMEDIATE {
		i3 = 3
	} else if src_t == REGISTER {
		if des_t == REGISTER {
			i3 = 0
		} else {
			i3 = 2
		}
	}
	return i_2opcode[i1][i2][i3]
}

var TmpCodeSeg *os.File

func init() {
	var err error
	TmpCodeSeg, err = os.OpenFile("tmp_code_seg.out", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
		panic("init err: open elf.out failed")
	}
}
