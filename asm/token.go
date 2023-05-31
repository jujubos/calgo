package asm

import (
	"fmt"
	"strings"
)

type TokenType int

type Token interface {
	String() string
	TokenTyp() TokenType
}

type TID struct {
	Type  TokenType
	Name  string
	Value string
}

type TNUM struct {
	Type  TokenType
	Name  string
	Value int64
}

type TSTR struct {
	Type  TokenType
	Name  string
	Value string
}

type TREG struct {
	Type  TokenType
	Name  string
	Value string
}

type TINST struct {
	Type  TokenType
	Name  string
	Value string
}

type TKWORD struct {
	Type  TokenType
	Name  string
	Value string
}

type TBOUND struct {
	Type  TokenType
	Name  string
	Value string
}

type TERR struct {
	Type  TokenType
	Name  string
	Value string
}

type TEOF struct {
	Type  TokenType
	Name  string
	Value string
}

func (T *TID) String() string {
	return fmt.Sprintf("<%s, %s>:", tokenTypeTable[T.Type], T.Name)
}

func (T *TNUM) String() string {
	return fmt.Sprintf("<%s, %s, %d>", tokenTypeTable[T.Type], T.Name, T.Value)
}

func (T *TSTR) String() string {
	builder := &strings.Builder{}
	for i := range T.Value {
		c := T.Value[i]
		switch c {
		case '\n':
			builder.WriteString("\\n")
		case '\t':
			builder.WriteString("\\t")
		case '"':
			builder.WriteString("\\\"")
		default:
			builder.WriteByte(c)
		}
	}
	return fmt.Sprintf("%d:%s", tokenTypeTable[T.Type], builder.String())
}

func (T *TREG) String() string {
	return fmt.Sprintf("<%s, %s>:", tokenTypeTable[T.Type], T.Name)
}

func (T *TKWORD) String() string {
	return fmt.Sprintf("<%s, %s>:", tokenTypeTable[T.Type], T.Name)
}

func (T *TINST) String() string {
	return fmt.Sprintf("<%s, %s>:", tokenTypeTable[T.Type], T.Name)
}

func (T *TBOUND) String() string {
	return fmt.Sprintf("<%s, %s>:", tokenTypeTable[T.Type], T.Name)
}

func (T *TERR) String() string {
	return fmt.Sprintf("<%s, %s>:", tokenTypeTable[T.Type], T.Name)
}

func (T *TEOF) String() string {
	return fmt.Sprintf("<%s, %s>:", tokenTypeTable[T.Type], T.Name)
}

// 关键字和标识符
func (T *TID) TokenTyp() TokenType {
	return T.Type
}

func (T *TNUM) TokenTyp() TokenType {
	return NUM
}

func (T *TSTR) TokenTyp() TokenType {
	return STR
}

func (T *TREG) TokenTyp() TokenType {
	return T.Type
}

func (T *TKWORD) TokenTyp() TokenType {
	return T.Type
}

func (T *TINST) TokenTyp() TokenType {
	return T.Type
}

func (T *TBOUND) TokenTyp() TokenType {
	return T.Type
}

func (T *TERR) TokenTyp() TokenType {
	return ERR
}

func (T *TEOF) TokenTyp() TokenType {
	return EOF
}

const (
	_             = iota
	ERR TokenType = iota
	EOF
	ID
	NUM
	STR
	BR_AL
	BR_CL
	BR_DL
	BR_BL
	BR_AH
	BR_CH
	BR_DH
	BR_BH
	DR_EAX
	DR_ECX
	DR_EDX
	DR_EBX
	DR_ESP
	DR_EBP
	DR_ESI
	DR_EDI
	I_MOV
	I_CMP
	I_SUB
	I_ADD
	I_AND
	I_OR
	I_LEA
	I_CALL
	I_INT
	I_IMUL
	I_IDIV
	I_NEG
	I_INC
	I_DEC
	I_JMP
	I_JE
	I_JNE
	I_SETE
	I_SETNE
	I_SETG
	I_SETGE
	I_SETL
	I_SETLE
	I_PUSH
	I_POP
	I_RET
	KW_SEC
	KW_GLB
	KW_EQU
	KW_TIMES
	KW_DB
	KW_DW
	KW_DD
	ADD
	SUB
	COMMA
	LBRACK
	RBRACK
	COLON
)

var tokenTypeTable = []string{
	"",
	"ERR",
	"EOF",
	"ID",
	"NUM",
	"STR",
	"BR_AL",
	"BR_CL",
	"BR_DL",
	"BR_BL",
	"BR_AH",
	"BR_CH",
	"BR_DH",
	"BR_BH",
	"DR_EAX",
	"DR_ECX",
	"DR_EDX",
	"DR_EBX",
	"DR_ESP",
	"DR_EBP",
	"DR_ESI",
	"DR_EDI",
	"I_MOV",
	"I_CMP",
	"I_SUB",
	"I_ADD",
	"I_AND",
	"I_OR",
	"I_LEA",
	"I_CALL",
	"I_INT",
	"I_IMUL",
	"I_IDIV",
	"I_NEG",
	"I_INC",
	"I_DEC",
	"I_JMP",
	"I_JE",
	"I_JNE",
	"I_SETE",
	"I_SETNE",
	"I_SETG",
	"I_SETGE",
	"I_SETL",
	"I_SETLE",
	"I_PUSH",
	"I_POP",
	"I_RET",
	"KW_SEC",
	"KW_GLB",
	"KW_EQU",
	"KW_TIMES",
	"KW_DB",
	"KW_DW",
	"KW_DD",
	"ADD",
	"SUB",
	"COMMA",
	"LBRACK",
	"RBRACK",
	"COLON",
}
