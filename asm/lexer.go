package asm

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var kwords = map[string]TokenType{
	"section": KW_SEC,
	"global":  KW_GLB,
	"equ":     KW_EQU,
	"times":   KW_TIMES,
	"db":      KW_DB,
	"dw":      KW_DW,
	"dd":      KW_DD,
	"al":      BR_AL,
	"cl":      BR_CL,
	"dl":      BR_DL,
	"bl":      BR_BL,
	"ah":      BR_AH,
	"ch":      BR_CH,
	"dh":      BR_DH,
	"bh":      BR_BH,
	"eax":     DR_EAX,
	"ecx":     DR_ECX,
	"edx":     DR_EDX,
	"ebx":     DR_EBX,
	"esp":     DR_ESP,
	"ebp":     DR_EBP,
	"esi":     DR_ESI,
	"edi":     DR_EDI,
	"mov":     I_MOV,
	"cmp":     I_CMP,
	"sub":     I_SUB,
	"add":     I_ADD,
	"and":     I_AND,
	"or":      I_OR,
	"lea":     I_LEA,
	"call":    I_CALL,
	"int":     I_INT,
	"imul":    I_IMUL,
	"idiv":    I_IDIV,
	"neg":     I_NEG,
	"inc":     I_INC,
	"dec":     I_DEC,
	"jmp":     I_JMP,
	"je":      I_JE,
	"jne":     I_JNE,
	"sete":    I_SETE,
	"setne":   I_SETNE,
	"setg":    I_SETG,
	"setge":   I_SETGE,
	"setl":    I_SETL,
	"setle":   I_SETLE,
	"push":    I_PUSH,
	"pop":     I_POP,
	"ret":     I_RET,
}

var lexErrorTable = map[string]string{
	"": "",
}

type Lexer struct {
	ch       byte
	scanner  *bufio.Reader
	lineNum  int
	colNum   int
	filename string
	newline  bool
}

func (l *Lexer) GetPosition() (string, int, int) {
	return l.filename, l.lineNum, l.colNum
}

func NewLexer(filename string) *Lexer {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	lexer := &Lexer{
		scanner:  bufio.NewReader(file),
		filename: filename,
		lineNum:  0,
		colNum:   0,
		newline:  true,
	}
	lexer.NextChar()
	return lexer
}

func (l *Lexer) Error() {
	fmt.Printf("词法错误: %s <%d, %d>\n", l.filename, l.lineNum, l.colNum)
	os.Exit(0)
}

func (l *Lexer) NextToken() Token {
	//设置l.ch为第一个非空白字符
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' {
		l.NextChar()
	}
	builder := strings.Builder{}
	for {
		if l.ch == '@' || l.ch == '.' || isAlpha(l.ch) {
			builder.WriteByte(l.ch)
			l.NextChar()
			for l.ch == '@' || l.ch == '.' || isAlpha(l.ch) || isDigit(l.ch) {
				builder.WriteByte(l.ch)
				l.NextChar()
			}
			name := builder.String()
			if typ, ok := kwords[name]; ok {
				return &TKWORD{Type: typ, Name: name}
			}
			return &TID{Type: ID, Name: name}
		} else if isDigit(l.ch) {
			for isDigit(l.ch) {
				builder.WriteByte(l.ch)
				l.NextChar()
			}
			v, err := strconv.Atoi(builder.String())
			if err != nil {
				log.Fatal(err)
				os.Exit(0)
			}
			return &TNUM{Type: NUM, Name: builder.String(), Value: int64(v)}
		} else if l.ch == '"' {
			l.NextChar()
			for l.ch != '"' && l.ch != 0 {
				builder.WriteByte(l.ch)
				l.NextChar()
			}
			if l.ch == 0 {
				l.Error()
			}
			l.NextChar()
			return &TSTR{Type: STR, Name: builder.String(), Value: builder.String()}
		} else {
			switch l.ch {
			case '+':
				l.NextChar()
				return &TBOUND{Type: ADD, Name: "+"}
			case '-':
				l.NextChar()
				return &TBOUND{Type: SUB, Name: "-"}
			case ',':
				l.NextChar()
				return &TBOUND{Type: COMMA, Name: ","}
			case '[':
				l.NextChar()
				return &TBOUND{Type: LBRACK, Name: "["}
			case ']':
				l.NextChar()
				return &TBOUND{Type: RBRACK, Name: "]"}
			case ':':
				l.NextChar()
				return &TBOUND{Type: COLON, Name: ":"}
			default:
				if l.ch == 0 {
					return &TEOF{Type: EOF, Name: "EOF"}
				}
				l.Error()
			}
		}
	}
	return nil
}

func isAlpha(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// 前移字符指针
func (l *Lexer) NextChar() {
	var err error
	c, err := l.scanner.ReadByte()
	if err != nil {
		l.ch = 0
	}
	if l.newline {
		l.lineNum++
		l.colNum = 1
	} else {
		l.colNum++
	}
	if c == '\n' {
		l.newline = true
	} else {
		l.newline = false
	}
	l.ch = c
}
