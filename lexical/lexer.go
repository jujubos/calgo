package lexical

import (
	"bufio"
	"os"
	"strings"
)

var kwords = map[string]TokenType{
	"int":      KW_INT,
	"char":     KW_CHAR,
	"void":     KW_VOID,
	"extern":   KW_EXTERN,
	"if":       KW_IF,
	"else":     KW_ELSE,
	"switch":   KW_SWITCH,
	"case":     KW_CASE,
	"default":  KW_DEFAULT,
	"while":    KW_WHILE,
	"do":       KW_DO,
	"for":      KW_FOR,
	"break":    KW_BREAK,
	"continue": KW_CONINUE,
	"return":   KW_RETURN,
}

type Lexer struct {
	ch      byte
	scanner *bufio.Reader
}

func NewLexer(filename string) *Lexer {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	lexer := &Lexer{
		scanner: bufio.NewReader(file),
	}
	lexer.NextChar()
	return lexer
}

func (l *Lexer) NextToken() Token {
	//设置l.ch为第一个非空白字符
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' {
		l.NextChar()
	}
	builder := strings.Builder{}
	for {
		c := l.ch
		if c == '_' || isAlpha(c) {
			builder.WriteByte(l.ch)
			l.NextChar()
			for l.ch == '_' || isDigit(l.ch) || isAlpha(l.ch) {
				builder.WriteByte(l.ch)
				l.NextChar()
			}
			//TODO 关键字
			if t, ok := kwords[builder.String()]; ok {
				return &TID{Type: t, Name: builder.String()}
			}
			return &TID{Type: ID, Name: builder.String()}
		} else if isDigit(c) {
			var v int64
			if l.ch == '0' {
				builder.WriteByte(l.ch)
				l.NextChar()
				if l.ch == 'x' {
					builder.WriteByte(l.ch)
					l.NextChar()
					if isDigit(l.ch) || l.ch >= 'A' && l.ch <= 'F' || l.ch >= 'a' && l.ch <= 'f' {
						builder.WriteByte(l.ch)
						v = v * 16
						if isDigit(l.ch) {
							v += int64(l.ch - '0')
						} else if l.ch >= 'A' && l.ch <= 'F' {
							v += int64(l.ch - 'A')
						} else {
							v += int64(l.ch - 'a')
						}
						l.NextChar()
						for isDigit(l.ch) || l.ch >= 'A' && l.ch <= 'F' || l.ch >= 'a' && l.ch <= 'f' {
							builder.WriteByte(l.ch)
							v = v * 16
							if isDigit(l.ch) {
								v += int64(l.ch - '0')
							} else if l.ch >= 'A' && l.ch <= 'F' {
								v += int64(l.ch - 'A')
							} else {
								v += int64(l.ch - 'a')
							}
							l.NextChar()
						}
						return &TNUM{Type: NUM, Name: builder.String(), Value: v}

					} else {
						return &ERR_TOKEN
					}
				} else if l.ch == 'b' {
					builder.WriteByte(l.ch)
					l.NextChar()
					if l.ch == '0' || l.ch == '1' {
						builder.WriteByte(l.ch)
						v = v*2 + int64(l.ch-'0')
						l.NextChar()
						for l.ch == '0' || l.ch == '1' {
							builder.WriteByte(l.ch)
							v = v*2 + int64(l.ch-'0')
							l.NextChar()
						}
						return &TNUM{Type: NUM, Name: builder.String(), Value: v}
					} else {
						return &ERR_TOKEN
					}
				} else if l.ch >= '0' && l.ch <= '7' {
					for l.ch >= '0' && l.ch <= '7' {
						builder.WriteByte(l.ch)
						v = v*8 + int64(l.ch-'0')
						l.NextChar()
					}
					return &TNUM{Type: NUM, Name: builder.String(), Value: v}
				} else {

				}

			} else { //十进制
				builder.WriteByte(l.ch)
				v = v*10 + int64(l.ch-'0')
				l.NextChar()
				for isDigit(l.ch) {
					builder.WriteByte(l.ch)
					v = v*10 + int64(l.ch-'0')
					l.NextChar()
				}
				return &TNUM{Type: NUM, Name: builder.String(), Value: v}
			}
		} else if c == '\'' {
			var ch byte
			l.NextChar()
			if l.ch == 0 || l.ch == '\n' || l.ch == '\'' {
				return &ERR_TOKEN
			}
			if l.ch == '\\' {
				l.NextChar()
				if l.ch == 0 || l.ch == '\n' || (l.ch != 'n' && l.ch != 't' && l.ch != '\\' && l.ch != '\'' && l.ch != '0') {
					return &ERR_TOKEN
				}
				ch = l.ch
				l.NextChar()
				token := &TCHAR{Type: CHAR}
				if l.ch == '\'' {
					switch ch {
					case 't':
						token.Name = string("\\t")
						token.Value = '\t'
					case 'n':
						token.Name = string("\\n")
						token.Value = '\n'
					case '\\':
						token.Name = string("\\")
						token.Value = '\\'
					case '\'':
						token.Name = string("\\'")
						token.Value = '\''
					case '0':
						token.Name = string("\\0")
						token.Value = 0
					}
					l.NextChar()
					return token
				} else {
					return &ERR_TOKEN
				}
			} else {
				ch = l.ch
				l.NextChar()
				if l.ch == '\'' {
					l.NextChar()
					return &TCHAR{Type: CHAR, Name: string(ch)}
				} else {
					return &ERR_TOKEN
				}
			}
		} else if c == '"' {
			l.NextChar()
			for {
				if l.ch == 0 || l.ch == '\n' {
					return &ERR_TOKEN
				} else if l.ch == '\\' {
					l.NextChar()
					if l.ch == 0 || l.ch == '\n' {
						return &ERR_TOKEN
					} else if l.ch == 't' {
						builder.WriteByte('\t')
						l.NextChar()
					} else if l.ch == 'n' {
						builder.WriteByte('\n')
						l.NextChar()
					} else if l.ch == '"' {
						builder.WriteByte('"')
						l.NextChar()
					} else {
						return &ERR_TOKEN
					}
				} else if l.ch == '"' {
					l.NextChar()
					return &TSTR{Type: STR, Name: builder.String(), Value: builder.String()}
				} else {
					builder.WriteByte(l.ch)
					l.NextChar()
				}
			}
		} else {
			switch c {
			case '+':
				l.NextChar()
				if l.ch == '+' {
					l.NextChar()
					return &TINC{Type: INC, Name: "++"}
				} else {
					return &TADD{Type: ADD, Name: "+"}
				}
			case '-':
				l.NextChar()
				if l.ch == '-' {
					l.NextChar()
					return &TDEC{Type: DEC, Name: "--"}
				} else {
					return &TSUB{Type: SUB, Name: "-"}
				}
			case '*':
				l.NextChar()
				return &TMUL{Type: MUL, Name: "*"}
			case '/':
				l.NextChar()
				return &TDIV{Type: DIV, Name: "/"}
			case '%':
				l.NextChar()
				return &TMOD{Type: MOD, Name: "%"}
			case '&':
				l.NextChar()
				if l.ch == '&' {
					l.NextChar()
					return &TAND{Type: AND, Name: "&&"}
				} else {
					return &TLEA{Type: LEA, Name: "&"}
				}
			case '>':
				l.NextChar()
				if l.ch == '=' {
					l.NextChar()
					return &TGE{Type: GE, Name: ">="}
				} else {
					return &TGT{Type: GT, Name: ">"}
				}
			case '<':
				l.NextChar()
				if l.ch == '=' {
					l.NextChar()
					return &TLE{Type: LE, Name: "<="}
				} else {
					return &TLT{Type: LT, Name: "<"}
				}
			case '=':
				l.NextChar()
				if l.ch == '=' {
					l.NextChar()
					return &TEQU{Type: EQU, Name: "=="}
				} else {
					return &TASSIGN{Type: ASSIGN, Name: "="}
				}
			case '!':
				l.NextChar()
				if l.ch == '=' {
					l.NextChar()
					return &TNEQU{Type: NEQU, Name: "!="}
				} else {
					return &TNOT{Type: NOT, Name: "!"}
				}
			case ',':
				l.NextChar()
				return &TCOMMA{Type: COMMA, Name: ","}
			case ':':
				l.NextChar()
				return &TCOLON{Type: COLON, Name: ":"}
			case ';':
				l.NextChar()
				return &TSEMICOLON{Type: SEMICOLON, Name: ";"}
			case '(':
				l.NextChar()
				return &TLPARAN{Type: LPARAN, Name: "("}
			case ')':
				l.NextChar()
				return &TRPARAN{Type: RPARAN, Name: ")"}
			case '[':
				l.NextChar()
				return &TLBRACK{Type: LBRACK, Name: "["}
			case ']':
				l.NextChar()
				return &TRBRACK{Type: RBRACK, Name: "]"}
			case '{':
				l.NextChar()
				return &TLBRACE{Type: LBRACE, Name: "{"}
			case '}':
				l.NextChar()
				return &TRBRACE{Type: RBRACE, Name: "}"}
			default:
				return &ERR_TOKEN
			}
		}
	}
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
	l.ch = c
}
