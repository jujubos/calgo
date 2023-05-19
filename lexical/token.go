package lexical

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

type TCHAR struct {
	Type  TokenType
	Name  string
	Value byte
}

type TSTR struct {
	Type  TokenType
	Name  string
	Value string
}

type TADD struct {
	Type  TokenType
	Name  string
	Value string
}

type TSUB struct {
	Type  TokenType
	Name  string
	Value string
}

type TMUL struct {
	Type  TokenType
	Name  string
	Value string
}

type TDIV struct {
	Type  TokenType
	Name  string
	Value string
}

type TMOD struct {
	Type  TokenType
	Name  string
	Value string
}

type TINC struct {
	Type  TokenType
	Name  string
	Value string
}

type TDEC struct {
	Type  TokenType
	Name  string
	Value string
}

type TNOT struct {
	Type  TokenType
	Name  string
	Value string
}

type TLEA struct {
	Type  TokenType
	Name  string
	Value string
}

type TAND struct {
	Type  TokenType
	Name  string
	Value string
}

type TOR struct {
	Type  TokenType
	Name  string
	Value string
}

type TASSIGN struct {
	Type  TokenType
	Name  string
	Value string
}

type TGT struct {
	Type  TokenType
	Name  string
	Value string
}

type TGE struct {
	Type  TokenType
	Name  string
	Value string
}

type TLT struct {
	Type  TokenType
	Name  string
	Value string
}

type TLE struct {
	Type  TokenType
	Name  string
	Value string
}

type TEQU struct {
	Type  TokenType
	Name  string
	Value string
}

type TNEQU struct {
	Type  TokenType
	Name  string
	Value string
}

type TCOMMA struct {
	Type  TokenType
	Name  string
	Value string
}

type TCOLON struct {
	Type  TokenType
	Name  string
	Value string
}

type TSEMICOLON struct {
	Type  TokenType
	Name  string
	Value string
}

type TLPARAN struct {
	Type  TokenType
	Name  string
	Value string
}

type TRPARAN struct {
	Type  TokenType
	Name  string
	Value string
}

type TLBRACK struct {
	Type  TokenType
	Name  string
	Value string
}

type TRBRACK struct {
	Type  TokenType
	Name  string
	Value string
}

type TLBRACE struct {
	Type  TokenType
	Name  string
	Value string
}

type TRBRACE struct {
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
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TNUM) String() string {
	return fmt.Sprintf("%d:%s Value:%d", T.Type, T.Name, T.Value)
}

func (T *TCHAR) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
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
	return fmt.Sprintf("%d:%s", T.Type, builder.String())
}

func (T *TERR) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TEOF) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TADD) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TSUB) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TMUL) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TDIV) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TMOD) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TINC) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TDEC) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TNOT) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TLEA) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TAND) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TOR) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TASSIGN) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TGT) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TGE) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TLT) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TLE) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TEQU) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TNEQU) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TCOMMA) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TCOLON) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TSEMICOLON) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TLPARAN) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TRPARAN) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TLBRACK) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TRBRACK) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TLBRACE) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

func (T *TRBRACE) String() string {
	return fmt.Sprintf("%d:%s", T.Type, T.Name)
}

// 关键字和标识符
func (T *TID) TokenTyp() TokenType {
	return T.Type
}

func (T *TNUM) TokenTyp() TokenType {
	return NUM
}

func (T *TCHAR) TokenTyp() TokenType {
	return CHAR
}

func (T *TSTR) TokenTyp() TokenType {
	return STR
}

func (T *TERR) TokenTyp() TokenType {
	return ERR
}

func (T *TEOF) TokenTyp() TokenType {
	return EOF
}

func (T *TADD) TokenTyp() TokenType {
	return ADD
}

func (T *TSUB) TokenTyp() TokenType {
	return SUB
}

func (T *TMUL) TokenTyp() TokenType {
	return MUL
}

func (T *TDIV) TokenTyp() TokenType {
	return DIV
}

func (T *TMOD) TokenTyp() TokenType {
	return MOD
}

func (T *TINC) TokenTyp() TokenType {
	return INC
}

func (T *TDEC) TokenTyp() TokenType {
	return DEC
}

func (T *TNOT) TokenTyp() TokenType {
	return NOT
}

func (T *TLEA) TokenTyp() TokenType {
	return LEA
}

func (T *TAND) TokenTyp() TokenType {
	return AND
}

func (T *TOR) TokenTyp() TokenType {
	return OR
}

func (T *TASSIGN) TokenTyp() TokenType {
	return ASSIGN
}

func (T *TGT) TokenTyp() TokenType {
	return GT
}

func (T *TGE) TokenTyp() TokenType {
	return GE
}

func (T *TLT) TokenTyp() TokenType {
	return LT
}

func (T *TLE) TokenTyp() TokenType {
	return LE
}

func (T *TEQU) TokenTyp() TokenType {
	return EQU
}

func (T *TNEQU) TokenTyp() TokenType {
	return NEQU
}

func (T *TCOMMA) TokenTyp() TokenType {
	return COMMA
}

func (T *TCOLON) TokenTyp() TokenType {
	return COLON
}

func (T *TSEMICOLON) TokenTyp() TokenType {
	return SEMICOLON
}

func (T *TLPARAN) TokenTyp() TokenType {
	return LPARAN
}

func (T *TRPARAN) TokenTyp() TokenType {
	return RPARAN
}

func (T *TLBRACK) TokenTyp() TokenType {
	return LBRACK
}

func (T *TRBRACK) TokenTyp() TokenType {
	return RBRACK
}

func (T *TLBRACE) TokenTyp() TokenType {
	return LBRACE
}

func (T *TRBRACE) TokenTyp() TokenType {
	return RBRACE
}

const (
	ERR TokenType = iota
	EOF
	ID
	CHAR
	NUM
	STR
	KW_INT
	KW_CHAR
	KW_VOID
	KW_EXTERN
	KW_IF
	KW_ELSE
	KW_SWITCH
	KW_CASE
	KW_DEFAULT
	KW_WHILE
	KW_DO
	KW_FOR
	KW_BREAK
	KW_CONINUE
	KW_RETURN
	ADD
	SUB
	MUL
	DIV
	MOD
	INC
	DEC
	NOT
	LEA
	AND
	OR
	ASSIGN
	GT
	GE
	LT
	LE
	EQU
	NEQU
	COMMA
	COLON
	SEMICOLON
	LPARAN
	RPARAN
	LBRACK
	RBRACK
	LBRACE
	RBRACE
)

var EOF_TOKEN = TEOF{Type: EOF, Name: "EOF", Value: "EOF"}
var ERR_TOKEN = TERR{Type: ERR, Name: "ERR", Value: "ERR"}
