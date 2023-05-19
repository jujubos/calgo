package syntax

import (
	"calgo/lexical"
	"fmt"
)

type Parser struct {
	tk    lexical.Token
	lexer *lexical.Lexer
}

func NewParser(filename string) *Parser {
	return &Parser{
		lexer: lexical.NewLexer(filename),
	}
}

func (p *Parser) Parse() {
	p.move()
	p.program()
}

// <program> -> <segment> <program> | ^
func (p *Parser) program() {
	if p.match(lexical.ERR) {
		return
	}
	p.segment()
	p.program()
}

// <segment> -> extern <type> <def> | <type> <def>
func (p *Parser) segment() {
	if p.match(lexical.KW_EXTERN) {
		p.move()
		p.typedec()
		p.def()
	} else {
		p.typedec()
		p.def()
	}
}

// <type> ->	int | char | void
func (p *Parser) typedec() {
	if p.match(lexical.KW_INT) || p.match(lexical.KW_CHAR) || p.match(lexical.KW_VOID) {
		p.move()
	} else {
		fmt.Errorf("syntax err: typedec err!")
	}
}

// <def> -> mul id <init> <deflist> | id <idtail>
func (p *Parser) def() {
	if p.match(lexical.MUL) {
		p.move()
		if p.match(lexical.ID) {
			p.move()
			p.init()
			p.deflist()
		} else {
			fmt.Errorf("syntax err: def err!")
		}
	} else if p.match(lexical.ID) {
		p.move()
		p.idtail()
	} else {
		fmt.Errorf("syntax err: def err!")
	}
}

// <init> -> assign <expr> | ^
func (p *Parser) init() {
	if p.match(lexical.ASSIGN) {
		p.move()
		p.expr()
	}
}

// <deflist> -> comma <defdata> <deflist> | semicon
func (p *Parser) deflist() {
	if p.match(lexical.COMMA) {
		p.move()
		p.defdata()
		p.deflist()
	} else if p.match(lexical.SEMICOLON) {
		p.move()
	} else {
		fmt.Errorf("syntax err: deflist err!")
	}
}

// <idtail> ->	<varrdef> <deflist> | lparen <para> rparen <funtail>
func (p *Parser) idtail() {
	if p.match(lexical.LPARAN) {
		p.move()
		p.para()
		if !p.match(lexical.RPARAN) {
			fmt.Errorf("syntax err: idtail err!")
		}
		p.funtail()
	} else {
		p.varrdef()
		p.deflist()
	}
}

// <expr> -> <assexpr>
func (p *Parser) expr() {
	p.assexpr()
}

// <defdata> -> id <varrdef> | mul id <init>
func (p *Parser) defdata() {
	if p.match(lexical.ID) {
		p.move()
		p.varrdef()
	} else if p.match(lexical.MUL) {
		p.move()
		if !p.match(lexical.ID) {
			fmt.Errorf("syntax err: defdata err!")
			p.init()
		}
	}
}

// <para> ->	<type> <paradata> <paralist> | ^
func (p *Parser) para() {
	if !p.match(lexical.ERR) {
		p.typedec()
		p.paradata()
		p.paralist()
	}
}

// <funtail> -> <block> | semicon
func (p *Parser) funtail() {
	if p.match(lexical.SEMICOLON) {
		p.move()
	} else {
		p.block()
	}
}

// <varrdef> -> lbrack num rbrack | <init>
func (p *Parser) varrdef() {
	if p.match(lexical.LBRACK) {
		p.move()
		if !p.match(lexical.NUM) {
			fmt.Errorf("syntax err: varrder err!")
		}
		p.move()
		if !p.match(lexical.RBRACK) {
			fmt.Errorf("syntax err: varrdef err")
		}
		p.move()
	} else {
		p.init()
	}
}

// TODO
// <assexpr> ->	<orexpr><asstail>
func (p *Parser) assexpr() {

}

// <paradata> -> mul id | id <paradatatail>
func (p *Parser) paradata() {
	if p.match(lexical.MUL) {
		p.move()
		if !p.match(lexical.ID) {
			fmt.Errorf("syntax err: paradata err!")
		}
	} else if p.match(lexical.ID) {
		p.move()
		p.paradatatail()
	} else {
		fmt.Errorf("syntax err: paradata err!")
	}
}

// <paralist>	-> comma <type> <paradata> <paralist> | ^
func (p *Parser) paralist() {
	if p.match(lexical.COMMA) {
		p.move()
		p.typedec()
		p.paradata()
		p.paralist()
	}
}

// <block> -> lbrace <subprogram> rbrace
func (p *Parser) block() {
	if p.match(lexical.LBRACE) {
		p.move()
		p.subprogram()
		if !p.match(lexical.RBRACE) {
			fmt.Errorf("syntax err: block err!")
			return
		}
	} else {
		fmt.Errorf("syntax err: block err!")
		return
	}
}

// <paradatatail> -> lbrack num rbrack | ^
func (p *Parser) paradatatail() {
	if p.match(lexical.LBRACK) {
		p.move()
		if !p.match(lexical.NUM) {
			fmt.Errorf("syntax err: paradatatail err!")
			return
		}
		p.move()
		if !p.match(lexical.RBRACK) {
			fmt.Errorf("syntax err: paradatatail err!")
			return
		}
		p.move()
	}
}

/*
*
<statement> ->	<altexpr> semicon

	| <whilestat> | <forstat> | <dowhilestat>
	| <ifstat> | <switchstat>
	| rsv_break semicon
	| rsv_continue semicon
	| rsv_return <altexpr> semicon

<localdef> -> <type> <defdata> <deflist>

<subprogram> -> <localdef> <subprogram> | <statement> <subprogram> | ^
*/
func (p *Parser) subprogram() {
	if p.match(lexical.KW_INT) || p.match(lexical.CHAR) || p.match(lexical.STR) {
		p.localdef()
		p.subprogram()
	} else {
		p.statement() //note that: 这里没有判断first集合，但逻辑没问题
		p.subprogram()
	}
}

// <localdef> -> <type> <defdata> <deflist>
func (p *Parser) localdef() {
	p.typedec()
	p.defdata()
	p.deflist()
}

// Statement
/*
<statement> ->	<altexpr> semicon
			| <whilestat> | <forstat> | <dowhilestat>
			| <ifstat> | <switchstat>
			| rsv_break semicon
			| rsv_continue semicon
			| rsv_return <altexpr> semicon
*/
func (p *Parser) statement() {
	switch p.tk.TokenTyp() {
	case lexical.KW_WHILE:
		p.whilestat()
	case lexical.KW_FOR:
		p.forstat()
	case lexical.KW_DO:
		p.dowhilestat()
	case lexical.KW_IF:
		p.ifstat()
	case lexical.KW_SWITCH:
		p.switchstat()
	case lexical.KW_BREAK:
		p.move()
		if !p.match(lexical.SEMICOLON) {
			fmt.Errorf("syntax err: statement err!")
		}
	case lexical.KW_CONINUE:
		p.move()
		if !p.match(lexical.SEMICOLON) {
			fmt.Errorf("syntax err: statement err!")
		}
	case lexical.KW_RETURN:
		p.move()
		p.altexpr()
		if !p.match(lexical.SEMICOLON) {
			fmt.Errorf("syntax err: statement err!")
		}
	default:
		p.altexpr()
		if !p.match(lexical.SEMICOLON) {
			fmt.Errorf("syntax err: statement err!")
		}
	}
}

// <whilestat> -> rsv_while lparen <altexpr> rparen <block>
func (p *Parser) whilestat() {

}

// <forstat> -> rsv_for lparen <forinit> semicon <altexpr> semicon <altexpr> rparen <block>
func (p *Parser) forstat() {

}

// <dowhilestat> -> rsv_do <block> rsv_while lparen <altexpr> rparen semicon
func (p *Parser) dowhilestat() {

}

// <ifstat> -> rsv_if lparen<expr> rparen <block> <elsestat>
func (p *Parser) ifstat() {

}

// <switchstat> -> rsv_switch lparen <expr> rparen lbrac <casestat> rbrac
func (p *Parser) switchstat() {

}

// <altexpr>	->	<expr> | ^
func (p *Parser) altexpr() {

}

func (p *Parser) match(typ lexical.TokenType) bool {
	return p.tk.TokenTyp() == typ
}

func (p *Parser) move() {
	p.tk = p.lexer.NextToken()
}
