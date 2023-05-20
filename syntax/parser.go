package syntax

import (
	"calgo/lexical"
	"fmt"
	"os"
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
	if !p.match(lexical.ERR) {
		fmt.Println("Parse err!")
	}
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
		fmt.Println("syntax err: typedec err!")
		os.Exit(0)
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
			fmt.Println("syntax err: def err!")
			os.Exit(0)
		}
	} else if p.match(lexical.ID) {
		p.move()
		p.idtail()
	} else {
		fmt.Println("syntax err: def err!")
		os.Exit(0)
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
		fmt.Println("syntax err: deflist err!")
	}
}

// <idtail> ->	<varrdef> <deflist> | lparen <para> rparen <funtail>
func (p *Parser) idtail() {
	if p.match(lexical.LPAREN) {
		p.move()
		p.para()
		if !p.match(lexical.RPAREN) {
			fmt.Println("syntax err: idtail err!")
			return
		}
		p.move()
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
			fmt.Println("syntax err: defdata err!")
			p.init()
		}
		p.move()
		p.init()
	}
}

// <para> ->	<type> <paradata> <paralist> | ^
func (p *Parser) para() {
	if p.match(lexical.KW_INT) || p.match(lexical.KW_CHAR) || p.match(lexical.KW_VOID) {
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
			fmt.Println("syntax err: varrder err!")
			return
		}
		p.move()
		if !p.match(lexical.RBRACK) {
			fmt.Println("syntax err: varrdef err")
			return
		}
		p.move()
	} else if p.match(lexical.ASSIGN) {
		p.init()
	}
}

// <assexpr> ->	<orexpr><asstail>
func (p *Parser) assexpr() {
	p.orexpr()
	p.asstail()
}

// <orexpr> -> 	<andexpr><ortail>
func (p *Parser) orexpr() {
	p.andexpr()
	p.ortail()
}

// <asstail>	->	assign <assexpr> | ^
func (p *Parser) asstail() {
	if p.match(lexical.ASSIGN) {
		p.move()
		p.assexpr()
	}
}

// <andexpr> -> <cmpexpr> <andtail>
func (p *Parser) andexpr() {
	p.cmpexpr()
	p.andtail()
}

// <ortail> 	-> 	or <andexpr> <ortail>|^
func (p *Parser) ortail() {
	if p.match(lexical.OR) {
		p.move()
		p.andexpr()
		p.ortail()
	}
}

// <cmpexpr>	->	<aloexpr><cmptail>
func (p *Parser) cmpexpr() {
	p.aloexpr()
	p.cmptail()
}

// <andtail> -> 	and <cmpexpr> <andtail>|^
func (p *Parser) andtail() {
	if p.match(lexical.AND) {
		p.move()
		p.cmpexpr()
		p.andtail()
	}
}

// <aloexpr> ->	<item><alotail>
func (p *Parser) aloexpr() {
	p.item()
	p.alotail()
}

// <cmptail> ->	<cmps><aloexpr><cmptail> | ^
func (p *Parser) cmptail() {
	if p.match(lexical.LT) || p.match(lexical.LE) || p.match(lexical.GT) || p.match(lexical.GE) ||
		p.match(lexical.EQU) || p.match(lexical.NEQU) {
		p.cmps()
		p.aloexpr()
		p.cmptail()
	}
}

// <item> -> <factor> <itemtail>x
func (p *Parser) item() {
	p.factor()
	p.itemtail()
}

// <alotail> ->	<adds> <item> <alotail> | ^
func (p *Parser) alotail() {
	if p.match(lexical.ADD) || p.match(lexical.SUB) {
		p.adds()
		p.item()
		p.alotail()
	}
}

// <cmps> ->	gt|ge|lt|le|equ|nequ
func (p *Parser) cmps() {
	if !p.match(lexical.GT) && !p.match(lexical.GE) && !p.match(lexical.LT) && !p.match(lexical.LE) &&
		!p.match(lexical.EQU) && !p.match(lexical.NEQU) {
		fmt.Println("syntax err: cmps err!")
		return
	}
	p.move()
}

// <factor> -> 	<lop> <factor> | <val>
func (p *Parser) factor() {
	if p.match(lexical.NOT) || p.match(lexical.SUB) || p.match(lexical.LEA) ||
		p.match(lexical.MUL) || p.match(lexical.INC) || p.match(lexical.DEC) {
		p.lop()
		p.factor()
	} else {
		p.val()
	}
}

// <itemtail> -> <muls> <factor> <itemtail> | ^
func (p *Parser) itemtail() {
	if p.match(lexical.MUL) || p.match(lexical.DIV) || p.match(lexical.MOD) {
		p.muls()
		p.factor()
		p.itemtail()
	}
}

// <adds> -> add|sub
func (p *Parser) adds() {
	if !p.match(lexical.ADD) && !p.match(lexical.SUB) {
		fmt.Println("syntax err: adds err!")
		return
	}
	p.move()
}

// <val> ->	<elem> <rop>
func (p *Parser) val() {
	p.elem()
	p.rop()
}

// <lop> ->  not|sub|lea|mul|inc|dec
func (p *Parser) lop() {
	if !p.match(lexical.NOT) && !p.match(lexical.SUB) && !p.match(lexical.ADD) &&
		!p.match(lexical.MUL) && !p.match(lexical.INC) && !p.match(lexical.DEC) {
		fmt.Println("syntax err: lop err!")
		return
	}
	p.move()
}

// <muls> -> mul | div | mod
func (p *Parser) muls() {
	if !p.match(lexical.MUL) && !p.match(lexical.DIV) && !p.match(lexical.MOD) {
		fmt.Println("syntax err: muls err!")
		return
	}
	p.move()
}

// <elem> ->	id <idexpr> | lparen <expr> rparen | <literal>
func (p *Parser) elem() {
	if p.match(lexical.ID) {
		p.move()
		p.idexpr()
	} else if p.match(lexical.LPAREN) {
		p.move()
		p.expr()
		if !p.match(lexical.RPAREN) {
			fmt.Println("syntax err: elem err!")
			return
		}
		p.move()
	} else {
		p.literal()
	}
}

// <rop>	-> inc | dec | ^
func (p *Parser) rop() {
	if p.match(lexical.INC) || p.match(lexical.DEC) {
		p.move()
	}
}

// <idexpr>	->	lbrack <expr> rbrack | lparen <realarg> rparen | ^
func (p *Parser) idexpr() {
	if p.match(lexical.LBRACK) {
		p.move()
		p.expr()
		if !p.match(lexical.RBRACK) {
			fmt.Println("syntax err: idexpr err!")
			return
		}
		p.move()
	} else if p.match(lexical.LPAREN) {
		p.move()
		p.realarg()
		if !p.match(lexical.RPAREN) {
			fmt.Println("syntax err: idexpr err!")
			return
		}
		p.move()
	}
}

// <realarg> ->	<arg> <arglist> | ^
func (p *Parser) realarg() {
	if p.matchExprFirst() {
		p.arg()
		p.arglist()
	}
}

// <arg> -> <expr>
func (p *Parser) arg() {
	p.expr()
}

// <arglist>	->	comma <arg> <arglist> | ^
func (p *Parser) arglist() {
	if p.match(lexical.COMMA) {
		p.move()
		p.arg()
		p.arglist()
	}
}

// <paradata> -> mul id | id <paradatatail>
func (p *Parser) paradata() {
	if p.match(lexical.MUL) {
		p.move()
		if !p.match(lexical.ID) {
			fmt.Println("syntax err: paradata err!")
			return
		}
		p.move()
	} else if p.match(lexical.ID) {
		p.move()
		p.paradatatail()
	} else {
		fmt.Println("syntax err: paradata err!")
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
			fmt.Println("syntax err: block err!")
			os.Exit(0)
		}
		p.move()
	} else {
		fmt.Println("syntax err: block err!")
		return
	}
}

// <paradatatail> -> lbrack num rbrack | ^
func (p *Parser) paradatatail() {
	if p.match(lexical.LBRACK) {
		p.move()
		if !p.match(lexical.NUM) {
			fmt.Println("syntax err: paradatatail err!")
			return
		}
		p.move()
		if !p.match(lexical.RBRACK) {
			fmt.Println("syntax err: paradatatail err!")
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
	if p.match(lexical.KW_INT) || p.match(lexical.KW_CHAR) || p.match(lexical.KW_VOID) {
		p.localdef()
		p.subprogram()
	} else if p.matchStatFirst() {
		p.statement() //note that: 这里如果不判断first集合，会无限递归
		p.subprogram()
	}
}

func (p *Parser) matchStatFirst() bool {
	return p.matchExprFirst() || p.match(lexical.SEMICOLON) || p.match(lexical.KW_WHILE) || p.match(lexical.KW_FOR) ||
		p.match(lexical.KW_DO) || p.match(lexical.KW_IF) || p.match(lexical.KW_SWITCH) || p.match(lexical.KW_RETURN) ||
		p.match(lexical.KW_BREAK) || p.match(lexical.KW_CONINUE)
}

func (p *Parser) matchExprFirst() bool {
	return p.match(lexical.LPAREN) || p.match(lexical.NUM) || p.match(lexical.CHAR) || p.match(lexical.STR) ||
		p.match(lexical.ID) || p.match(lexical.NOT) || p.match(lexical.SUB) || p.match(lexical.LEA) ||
		p.match(lexical.MUL) || p.match(lexical.INC) || p.match(lexical.DEC)
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
			fmt.Println("syntax err: statement err!")
			os.Exit(0)
		}
		p.move()
	case lexical.KW_CONINUE:
		p.move()
		if !p.match(lexical.SEMICOLON) {
			fmt.Println("syntax err: statement err!")
			os.Exit(0)
		}
		p.move()
	case lexical.KW_RETURN:
		p.move()
		p.altexpr()
		if !p.match(lexical.SEMICOLON) {
			fmt.Println("syntax err: statement err!")
			os.Exit(0)
		}
		p.move()
	default:
		p.altexpr()
		if !p.match(lexical.SEMICOLON) {
			fmt.Println("syntax err: statement err!")
			os.Exit(0)
		}
		p.move()
	}
}

// <whilestat> -> rsv_while lparen <altexpr> rparen <block>
func (p *Parser) whilestat() {
	if !p.match(lexical.KW_WHILE) {
		fmt.Println("syntax err: whilestat err!")
	}
	p.move()
	if !p.match(lexical.LPAREN) {
		fmt.Println("syntax err: whilestat err!")
	}
	p.move()
	p.altexpr()
	if !p.match(lexical.RPAREN) {
		fmt.Println("syntax err: whilestat err!")
	}
	p.move()
	p.block()
}

// <forstat> -> rsv_for lparen <forinit> <altexpr> semicon <altexpr> rparen <block>
func (p *Parser) forstat() {
	if !p.match(lexical.KW_FOR) {
		fmt.Println("syntax err: forstat err!")
		return
	}
	p.move()
	if !p.match(lexical.LPAREN) {
		fmt.Println("syntax err: forstat err!")
		return
	}
	p.move()
	p.forinit()
	p.altexpr()
	if !p.match(lexical.SEMICOLON) {
		fmt.Println("syntax err: forstat err!")
		return
	}
	p.move()
	p.altexpr()
	if !p.match(lexical.RPAREN) {
		fmt.Println("syntax err: forstat err!")
		return
	}
	p.move()
	p.block()
}

// <dowhilestat> -> rsv_do <block> rsv_while lparen <altexpr> rparen semicon
func (p *Parser) dowhilestat() {
	if !p.match(lexical.KW_DO) {
		fmt.Println("595syntax err: dowhilestat err!")
		os.Exit(0)
	}
	p.move()
	p.block()
	if !p.match(lexical.KW_WHILE) {
		fmt.Println("600syntax err: dowhilestat err!")
		os.Exit(0)
	}
	p.move()
	if !p.match(lexical.LPAREN) {
		fmt.Println("605syntax err: dowhilestat err!")
		os.Exit(0)
	}
	p.move()
	p.altexpr()
	if !p.match(lexical.RPAREN) {
		fmt.Println("611syntax err: dowhilestat err!")
		os.Exit(0)
	}
	p.move()
	if !p.match(lexical.SEMICOLON) {
		fmt.Println("616syntax err: dowhilestat err!")
		os.Exit(0)
	}
	p.move()
}

// <ifstat> -> rsv_if lparen <expr> rparen <block> <elsestat>
func (p *Parser) ifstat() {
	if !p.match(lexical.KW_IF) {
		fmt.Println("syntax err: ifstat err!")
		return
	}
	p.move()
	if !p.match(lexical.LPAREN) {
		fmt.Println("syntax err: dowhilestat err!")
		return
	}
	p.move()
	p.expr()
	if !p.match(lexical.RPAREN) {
		fmt.Println("syntax err: dowhilestat err!")
		return
	}
	p.move()
	p.block()
	p.elsestat()
}

// <switchstat> -> rsv_switch lparen <expr> rparen lbrace <casestat> rbrace
func (p *Parser) switchstat() {
	if !p.match(lexical.KW_SWITCH) {
		fmt.Println("syntax err: switchstat err!")
		return
	}
	p.move()
	if !p.match(lexical.LPAREN) {
		fmt.Println("syntax err: dowhilestat err!")
		return
	}
	p.move()
	p.expr()
	if !p.match(lexical.RPAREN) {
		fmt.Println("syntax err: dowhilestat err!")
		return
	}
	p.move()
	if !p.match(lexical.LBRACE) {
		fmt.Println("syntax err: dowhilestat err!")
		return
	}
	p.move()
	p.casestat()
	if !p.match(lexical.RBRACE) {
		fmt.Println("syntax err: dowhilestat err!")
		return
	}
	p.move()
}

// <forinit>  ->  <localdef> | <altexpr>
func (p *Parser) forinit() {
	if p.match(lexical.KW_INT) || p.match(lexical.KW_CHAR) || p.match(lexical.KW_VOID) {
		p.localdef()
	} else {
		p.altexpr()
	}
}

// <elsestat>	-> rsv_else <block> | ^
func (p *Parser) elsestat() {
	if p.match(lexical.KW_ELSE) {
		p.move()
		p.block()
	}
}

// <casestat> ->  rsv_case <caselabel> colon <subprogram> <casestat> | rsv_default colon <subprogram>
func (p *Parser) casestat() {
	if p.match(lexical.KW_CASE) {
		p.move()
		p.caselabel()
		if !p.match(lexical.COLON) {
			fmt.Println("syntax err: casestat err!")
		}
		p.move()
		p.subprogram()
		p.casestat()
	} else if p.match(lexical.KW_DEFAULT) {
		p.move()
		if !p.match(lexical.COLON) {
			fmt.Println("syntax err: casestat err!")
		}
		p.move()
		p.subprogram()
	} else {
		fmt.Println("syntax err: casestat err!")
	}
}

// <caselabel> -> <literal>
func (p *Parser) caselabel() {
	p.literal()
}

// <literal>	->	number | string | character
func (p *Parser) literal() {
	if !p.match(lexical.NUM) && !p.match(lexical.STR) && !p.match(lexical.CHAR) {
		fmt.Println("syntax err: literal err!")
		os.Exit(0)
	}
	p.move()
}

// <altexpr> ->	<expr> | ^
func (p *Parser) altexpr() {
	//note this
	p.expr()
}

func (p *Parser) match(typ lexical.TokenType) bool {
	return p.tk.TokenTyp() == typ
}

func (p *Parser) move() {
	p.tk = p.lexer.NextToken()
	fmt.Println(p.tk.String())
}
