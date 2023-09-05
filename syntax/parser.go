package syntax

import (
	"calgo/lexical"
	"calgo/table"
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

func (p *Parser) Error(info string) {
	fname, lnum, cnum := p.lexer.GetPosition()
	fmt.Printf("语法错误:%s in (line:%d,column:%d) of %s\n", info, lnum, cnum, fname)
	os.Exit(0)
}

func (p *Parser) Parse() {
	p.move()
	p.program()
	if !p.match(lexical.EOF) {
		p.Error(fmt.Sprintf("Parse err: expected EOF, but got %s", p.tk.String()))
	} else {
		fmt.Println("语法分析通过!")
	}
}

// <program> -> <segment> <program> | EOF
func (p *Parser) program() {
	if p.match(lexical.EOF) {
		return
	}
	p.segment()
	p.program()
}

// <segment> -> extern <type> <def> | <type> <def>
func (p *Parser) segment() {
	if p.match(lexical.KW_EXTERN) {
		p.move()
		t := p.typedec()
		p.def(true, t)
	} else {
		t := p.typedec()
		p.def(false, t)
	}
}

// <type> ->	int | char | void
func (p *Parser) typedec() lexical.TokenType {
	switch p.tk.TokenTyp() {
	case lexical.KW_INT:
		p.move()
		return lexical.KW_INT
	case lexical.KW_CHAR:
		p.move()
		return lexical.KW_CHAR
	case lexical.KW_VOID:
		p.move()
		return lexical.KW_VOID
	default:
		p.Error(fmt.Sprintf("typedec err: expected 'int', 'char' or 'void', but got %s", p.tk.String()))
	}
	return lexical.ERR
}

// <def> -> mul id <init> <deflist> | id <idtail>
func (p *Parser) def(ext bool, typ lexical.TokenType) {
	if p.match(lexical.MUL) {
		p.move()
		if p.match(lexical.ID) {
			varname := p.tk.(*lexical.TID).Name
			p.move()
			varr := p.init(ext, typ, true, varname)
			table.Symtab.AddVar(varr)
			p.deflist(ext, typ)
		} else {
			p.Error(fmt.Sprintf("def err: expected ID, but got %s", p.tk.String()))
		}
	} else if p.match(lexical.ID) {
		name := p.tk.(*lexical.TID).Name
		p.move()
		p.idtail(ext, typ, false, name)
	} else {
		p.Error(fmt.Sprintf("def err: expected *ID or ID, but got %s", p.tk.String()))
	}
}

// <init> -> assign <expr> | ^
// 如果某个变量的InitData为default，说明没有显示初始化
func (p *Parser) init(ext bool, typ lexical.TokenType, isPtr bool, varname string) *table.Var {
	initval := &table.Var{
		Name:    "nil",
		Literal: true, //常量
	}
	if p.match(lexical.ASSIGN) {
		p.move()
		initval = p.expr()
	}
	return table.NewVar(table.Symtab.ScopePath, ext, typ, isPtr, varname, initval)
}

// <deflist> -> comma <defdata> <deflist> | semicon
func (p *Parser) deflist(ext bool, typ lexical.TokenType) {
	if p.match(lexical.COMMA) {
		p.move()
		varr := p.defdata(ext, typ)
		table.Symtab.AddVar(varr)
		p.deflist(ext, typ)
	} else if p.match(lexical.SEMICOLON) {
		p.move()
	} else {
		p.Error(fmt.Sprintf("deflist err: expected ',' or ';', but got %s", p.tk.String()))
	}
}

// <idtail> ->	<varrdef> <deflist> | lparen <para> rparen <funtail>
// <idtail> 区分函数和变量
func (p *Parser) idtail(ext bool, typ lexical.TokenType, isPtr bool, name string) {
	if p.match(lexical.LPAREN) { //函数
		p.move()
		_, lnum, cnum := p.lexer.GetPosition()
		table.Symtab.Enter(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
		var paralist []*table.Var
		p.para(&paralist)
		if !p.match(lexical.RPAREN) {
			p.Error(fmt.Sprintf("idtail err: expected ')', but got %s", p.tk.String()))
		}
		p.move()
		fun := table.NewFun(ext, typ, name, paralist)
		p.funtail(fun)
		_, lnum, cnum = p.lexer.GetPosition()
		table.Symtab.Leave(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
	} else { //变量
		table.Symtab.AddVar(p.varrdef(ext, typ, isPtr, name))
		p.deflist(ext, typ)
	}
}

// <expr> -> <assexpr>
func (p *Parser) expr() *table.Var {
	return p.assexpr()
}

// <defdata> -> id <varrdef> | mul id <init>
// 区分指针变量和非指针变量
func (p *Parser) defdata(ext bool, typ lexical.TokenType) *table.Var {
	if p.match(lexical.ID) { //非指针
		varname := p.tk.(*lexical.TID).Name
		p.move()
		return p.varrdef(ext, typ, false, varname)
	} else if p.match(lexical.MUL) { //指针
		p.move()
		if !p.match(lexical.ID) {
			p.Error(fmt.Sprintf("defdata err: expected ID, but got %s", p.tk.String()))
		}
		varname := p.tk.(*lexical.TID).Name
		p.move()
		return p.init(ext, typ, false, varname)
	} else {
		p.Error(fmt.Sprintf("defdata err: expected ID or MUL, but got %s", p.tk.String()))
	}
	return nil
}

// <para> ->	<type> <paradata> <paralist> | ^
func (p *Parser) para(paralist *[]*table.Var) {
	if p.match(lexical.KW_INT) || p.match(lexical.KW_CHAR) || p.match(lexical.KW_VOID) {
		typ := p.typedec()
		v := p.paradata(typ)
		(*paralist) = append((*paralist), v)
		table.Symtab.AddVar(v)
		p.paralist(paralist)
	}
}

// <funtail> -> <block> | semicon
func (p *Parser) funtail(fun *table.Fun) {
	if p.match(lexical.SEMICOLON) { //函数声明
		p.move()
		table.Symtab.DecFun(fun)
	} else { //函数定义
		table.Symtab.DefFun(fun)
		p.block()
		table.Symtab.EndDefFun()
	}
}

// <varrdef> -> lbrack num rbrack | <init>
// 区分数组和非数组
func (p *Parser) varrdef(ext bool, typ lexical.TokenType, isPtr bool, varname string) *table.Var {
	if p.match(lexical.LBRACK) { //数组，不允许初始化
		p.move()
		if !p.match(lexical.NUM) {
			p.Error(fmt.Sprintf("varrdef err: expected NUM, but got %s", p.tk.String()))
		}
		arrlen := p.tk.(*lexical.TNUM).Value
		p.move()
		if !p.match(lexical.RBRACK) {
			p.Error(fmt.Sprintf("varrdef err: expected ']', but got %s", p.tk.String()))
		}
		p.move()
		return table.NewArrayVar(table.Symtab.ScopePath, ext, typ, varname, arrlen)
	} else { //非数组，允许初始化
		return p.init(ext, typ, isPtr, varname)
	}
	return nil
}

// <assexpr> ->	<orexpr> <asstail>
func (p *Parser) assexpr() *table.Var {
	lval := p.orexpr()
	return p.asstail(lval)
}

// <orexpr> -> 	<andexpr> <ortail>
func (p *Parser) orexpr() *table.Var {
	lval := p.andexpr()
	return p.ortail(lval)
}

// <asstail>	->	assign <orexpr> <asstail> | ^
func (p *Parser) asstail(lval *table.Var) *table.Var {
	if p.match(lexical.ASSIGN) {
		p.move()
		val := p.orexpr()
		rval := p.asstail(val)
		result := table.GenTwoOp(lexical.ASSIGN, lval, rval)
		return result
	}
	return lval
}

// <andexpr> -> <cmpexpr> <andtail>
func (p *Parser) andexpr() *table.Var {
	lval := p.cmpexpr()
	return p.andtail(lval)
}

// <ortail> 	-> 	or <andexpr> <ortail> | ^
func (p *Parser) ortail(lval *table.Var) *table.Var {
	if p.match(lexical.OR) {
		p.move()
		val := p.andexpr()
		result := table.GenTwoOp(lexical.OR, lval, val)
		return p.ortail(result)
	}
	return lval
}

// <cmpexpr>	->	<aloexpr><cmptail>
func (p *Parser) cmpexpr() *table.Var {
	lval := p.aloexpr()
	return p.cmptail(lval)
}

// <andtail> -> 	and <cmpexpr> <andtail> | ^
func (p *Parser) andtail(lval *table.Var) *table.Var {
	if p.match(lexical.AND) {
		p.move()
		val := p.cmpexpr()
		result := table.GenTwoOp(lexical.AND, lval, val)
		return p.andtail(result)
	}
	return lval
}

// <aloexpr> ->	<item><alotail>
// ADD | SUB
func (p *Parser) aloexpr() *table.Var {
	lval := p.item()
	return p.alotail(lval)
}

// <cmptail> ->	<cmps> <aloexpr> <cmptail> | ^
func (p *Parser) cmptail(lval *table.Var) *table.Var {
	if p.match(lexical.LT) || p.match(lexical.LE) || p.match(lexical.GT) || p.match(lexical.GE) ||
		p.match(lexical.EQU) || p.match(lexical.NEQU) {
		op := p.cmps()
		val := p.aloexpr()
		result := table.GenTwoOp(op, lval, val)
		return p.cmptail(result)
	}
	return lval
}

// <item> -> <factor> <itemtail>
func (p *Parser) item() *table.Var {
	lval := p.factor()
	return p.itemtail(lval)
}

// <alotail> ->	<adds> <item> <alotail> | ^
func (p *Parser) alotail(lval *table.Var) *table.Var {
	if p.match(lexical.ADD) || p.match(lexical.SUB) {
		op := p.adds()
		val := p.item()
		result := table.GenTwoOp(op, lval, val)
		return p.alotail(result)
	}
	/* choose production: <alotail> -> ^ */
	return lval
}

// <cmps> ->	gt|ge|lt|le|equ|nequ
func (p *Parser) cmps() lexical.TokenType {
	if !p.match(lexical.GT) && !p.match(lexical.GE) && !p.match(lexical.LT) && !p.match(lexical.LE) &&
		!p.match(lexical.EQU) && !p.match(lexical.NEQU) {
		p.Error(fmt.Sprintf("cmps err: expected '>', '>=', '<', '<=', '=', '!=', but got %s", p.tk.String()))
	}
	tk := p.tk
	p.move()
	return tk.TokenTyp()
}

// <factor> -> 	<lop> <factor> | <val>
func (p *Parser) factor() *table.Var {
	if p.match(lexical.NOT) || p.match(lexical.SUB) || p.match(lexical.LEA) ||
		p.match(lexical.MUL) || p.match(lexical.INC) || p.match(lexical.DEC) {
		op := p.lop()
		val := p.factor()
		return table.GenOneOpLeft(op, val)
	} else {
		return p.val()
	}
}

// <itemtail> -> <muls> <factor> <itemtail> | ^
// MUL | DIV | MOD
func (p *Parser) itemtail(lval *table.Var) *table.Var {
	if p.match(lexical.MUL) || p.match(lexical.DIV) || p.match(lexical.MOD) {
		op := p.muls()
		val := p.factor()
		result := table.GenTwoOp(op, lval, val)
		table.Symtab.AddVar(result)
		return p.itemtail(result)
	}
	return lval
}

// <adds> -> add|sub
func (p *Parser) adds() lexical.TokenType {
	if !p.match(lexical.ADD) && !p.match(lexical.SUB) {
		p.Error(fmt.Sprintf("adds err: expected '+', '-', but got %s", p.tk.String()))
	}
	tk := p.tk
	p.move()
	return tk.TokenTyp()
}

// <val> ->	<elem> <rop>
// INC | DEC
func (p *Parser) val() *table.Var {
	lval := p.elem()
	if p.match(lexical.INC) || p.match(lexical.DEC) {
		op := p.rop()
		lval = table.GenOneOpRight(op, lval) //???
	}
	return lval
}

// <lop> ->  not|sub|lea|mul|inc|dec
func (p *Parser) lop() lexical.TokenType {
	if !p.match(lexical.NOT) && !p.match(lexical.SUB) && !p.match(lexical.ADD) &&
		!p.match(lexical.MUL) && !p.match(lexical.INC) && !p.match(lexical.DEC) {
		p.Error(fmt.Sprintf("lop err: expected '!', '-', '+', '*', '++', '--', but got %s", p.tk.String()))
	}
	tk := p.tk
	p.move()
	return tk.TokenTyp()
}

// <muls> -> mul | div | mod
func (p *Parser) muls() lexical.TokenType {
	if !p.match(lexical.MUL) && !p.match(lexical.DIV) && !p.match(lexical.MOD) {
		p.Error(fmt.Sprintf("muls err: expected '*', '/', '%', but got %s", p.tk.String()))
	}
	tk := p.tk
	p.move()
	return tk.TokenTyp()
}

// <elem> ->	id <idexpr> | lparen <expr> rparen | <literal>
func (p *Parser) elem() *table.Var {
	var rs *table.Var
	if p.match(lexical.ID) { //变量、数组索引、函数调用
		name := p.tk.(*lexical.TID).Name
		p.move()
		rs = p.idexpr(name)
	} else if p.match(lexical.LPAREN) { //括号表达式
		p.move()
		rs = p.expr()
		if !p.match(lexical.RPAREN) {
			p.Error(fmt.Sprintf("elem err: expected ')', but got %s", p.tk.String()))
		}
		p.move()
	} else { //字面量
		rs = p.literal()
	}
	return rs
}

// <rop>	-> inc | dec | ^
func (p *Parser) rop() lexical.TokenType {
	if p.match(lexical.INC) || p.match(lexical.DEC) {
		tk := p.tk
		p.move()
		return tk.TokenTyp()
	}
	return 0
}

// <idexpr>	->	lbrack <expr> rbrack | lparen <realarg> rparen | ^
func (p *Parser) idexpr(name string) *table.Var {
	var rs *table.Var
	if p.match(lexical.LBRACK) { //数组索引
		p.move()
		idx := p.expr()
		if !p.match(lexical.RBRACK) {
			p.Error(fmt.Sprintf("idexpr err: expected ']', but got %s", p.tk.String()))
		}
		p.move()
		arr := table.Symtab.GetVar(name)
		rs = table.GenArray(arr, idx)
	} else if p.match(lexical.LPAREN) { //函数调用
		var args []*table.Var
		p.move()
		p.realarg(&args)
		if !p.match(lexical.RPAREN) {
			p.Error(fmt.Sprintf("idexpr err: expected ')', but got %s", p.tk.String()))
		}
		p.move()
		fun := table.Symtab.GetFun(name, args)
		rs = table.GenCall(fun, args)
	} else { //标识符表达式
		rs = table.Symtab.GetVar(name)
	}
	return rs
}

// <realarg> ->	<arg> <arglist> | ^
func (p *Parser) realarg(args *[]*table.Var) {
	if p.matchExprFirst() {
		(*args) = append((*args), p.arg())
		p.arglist(args)
	}
}

// <arg> -> <expr>
func (p *Parser) arg() *table.Var {
	return p.expr()
}

// <arglist>	->	comma <arg> <arglist> | ^
func (p *Parser) arglist(args *[]*table.Var) {
	if p.match(lexical.COMMA) {
		p.move()
		(*args) = append((*args), p.arg())
		p.arglist(args)
	}
}

// <paradata> -> mul id | id <paradatatail>
// 参数：指针、非指针（普通变量和数组）
func (p *Parser) paradata(typ lexical.TokenType) *table.Var {
	if p.match(lexical.MUL) { //指针
		p.move()
		if !p.match(lexical.ID) {
			p.Error(fmt.Sprintf("paradata err: expected ID, but got %s", p.tk.String()))
		}
		name := p.tk.(*lexical.TID).Name
		p.move()
		return table.NewVar(table.Symtab.ScopePath, false, typ, true, name, nil)
	} else if p.match(lexical.ID) { //普通变量和数组
		name := p.tk.(*lexical.TID).Name
		p.move()
		return p.paradatatail(typ, name)
	} else {
		p.Error(fmt.Sprintf("paradata err: expected ID or *ID, but got %s", p.tk.String()))
	}
	return nil
}

// <paralist>	-> comma <type> <paradata> <paralist> | ^
func (p *Parser) paralist(plist *[]*table.Var) {
	if p.match(lexical.COMMA) {
		p.move()
		typ := p.typedec()
		v := p.paradata(typ)
		*plist = append((*plist), v)
		table.Symtab.AddVar(v)
		p.paralist(plist)
	}
}

// <block> -> lbrace <subprogram> rbrace
func (p *Parser) block() {
	if p.match(lexical.LBRACE) {
		p.move()
		p.subprogram()
		if !p.match(lexical.RBRACE) {
			p.Error(fmt.Sprintf("block err: expected '}', but got %s", p.tk.String()))
		}
		p.move()
	} else {
		p.Error(fmt.Sprintf("block err: expected '{', but got %s", p.tk.String()))
	}
}

// <paradatatail> -> lbrack num rbrack | ^
func (p *Parser) paradatatail(typ lexical.TokenType, name string) *table.Var {
	if p.match(lexical.LBRACK) {
		p.move()
		if !p.match(lexical.NUM) {
			p.Error(fmt.Sprintf("paradatatail err: expected NUM, but got %s", p.tk.String()))
		}
		p.move()
		if !p.match(lexical.RBRACK) {
			p.Error(fmt.Sprintf("paradata err: expected ']', but got %s", p.tk.String()))
		}
		p.move()
	}
	return table.NewVar(table.Symtab.ScopePath, false, typ, false, name, nil)
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
	typ := p.typedec()
	table.Symtab.AddVar(p.defdata(false, typ))
	p.deflist(false, typ)
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
		table.GenBreak()
		if !p.match(lexical.SEMICOLON) {
			p.Error(fmt.Sprintf("statement err: expected ';', but got %s", p.tk.String()))
		}
		p.move()
	case lexical.KW_CONINUE:
		p.move()
		table.GenContinue()
		if !p.match(lexical.SEMICOLON) {
			p.Error(fmt.Sprintf("statement err: expected ';', but got %s", p.tk.String()))
		}
		p.move()
	case lexical.KW_RETURN:
		p.move()
		table.GenReturn(p.altexpr())
		if !p.match(lexical.SEMICOLON) {
			p.Error(fmt.Sprintf("statement err: expected ';', but got %s", p.tk.String()))
		}
		p.move()
	default:
		/* 对于标识符表达式statement，不产生任何指令。
		   如果是基于虚拟的编译器，对于标识符表达式语句，如果要产生指令，会产生：
				1.push identifier（将标识符表达式的值入栈）
		        2.pop (弹出栈顶元素）
		   等价于什么都不做。
		*/
		p.altexpr()
		if !p.match(lexical.SEMICOLON) {
			p.Error(fmt.Sprintf("statement err: expected ';', but got %s", p.tk.String()))
		}
		p.move()
	}
}

// <whilestat> -> rsv_while lparen <altexpr> rparen <block>
func (p *Parser) whilestat() {
	_, lnum, cnum := p.lexer.GetPosition()
	table.Symtab.Enter(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
	if !p.match(lexical.KW_WHILE) {
		p.Error(fmt.Sprintf("whilestat err: expected KW_WHILE, but got %s", p.tk.String()))
	}
	p.move()
	if !p.match(lexical.LPAREN) {
		p.Error(fmt.Sprintf("whilestat err: expected '(', but got %s", p.tk.String()))
	}
	p.move()
	_while, _exit := table.NewLabelInst(), table.NewLabelInst()
	table.GenWhileHead(_while, _exit)
	/* 对于基于虚拟机的编译器，由于虚拟机基于一个栈计算表达式，表达式的值留在栈中，以便后续使用。
	   对于c语言这种编译器，不是基于栈计算表达式的值，而是返回表达式值对应的符号索引。以便后续使用。
	   altexpr会创建一个临时变量，将其添加到符号表。这个临时变量用于存储表达式的值。所以返回这个临时变量（临时符号）在符号表的索引，
	   后续过程就可以使用到这个表达式的值了。
	*/
	cond := p.altexpr()
	table.GenWhileCond(cond, _exit)
	if !p.match(lexical.RPAREN) {
		p.Error(fmt.Sprintf("whilestat err: expected ')', but got %s", p.tk.String()))
	}
	p.move()
	p.block()
	table.GenWhileTail(_while, _exit)
	_, lnum, cnum = p.lexer.GetPosition()
	table.Symtab.Leave(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
}

// <forstat> -> rsv_for lparen <forinit> <altexpr> semicon <altexpr> rparen <block>
func (p *Parser) forstat() {
	_, lnum, cnum := p.lexer.GetPosition()
	table.Symtab.Enter(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
	if !p.match(lexical.KW_FOR) {
		p.Error(fmt.Sprintf("forstat err: expected KW_FOR, but got %s", p.tk.String()))
	}
	p.move()
	if !p.match(lexical.LPAREN) {
		p.Error(fmt.Sprintf("forstat err: expected '(', but got %s", p.tk.String()))
	}
	p.move()
	p.forinit()
	_for, _block, _step, _exit := table.NewLabelInst(), table.NewLabelInst(), table.NewLabelInst(), table.NewLabelInst()
	table.GenForHead(_for)
	table.Push(_step, _exit)
	cond := p.altexpr() //TODO:思考
	table.GenForCondBegin(_exit, _block, _step, cond)
	if !p.match(lexical.SEMICOLON) {
		p.Error(fmt.Sprintf("forstat err: expected ';', but got %s", p.tk.String()))
	}
	p.move()
	p.altexpr()
	if !p.match(lexical.RPAREN) {
		p.Error(fmt.Sprintf("forstat err: expected ')', but got %s", p.tk.String()))
	}
	p.move()
	table.GenForCondEnd(_for, _block)
	p.block()
	table.GenForTail(_step, _exit)
	_, lnum, cnum = p.lexer.GetPosition()
	table.Symtab.Leave(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
	table.Pop()
}

// <dowhilestat> -> rsv_do <block> rsv_while lparen <altexpr> rparen semicon
func (p *Parser) dowhilestat() {
	_, lnum, cnum := p.lexer.GetPosition()
	table.Symtab.Enter(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
	if !p.match(lexical.KW_DO) {
		p.Error(fmt.Sprintf("dowhilestat err: expected KW_DO, but got %s", p.tk.String()))
	}
	p.move()
	_do, _exit := table.NewLabelInst(), table.NewLabelInst()
	table.GenDoWhileHead(_do)
	table.Push(_do, _exit)
	p.block()
	if !p.match(lexical.KW_WHILE) {
		p.Error(fmt.Sprintf("dowhilestat err: expected KW_WHILE, but got %s", p.tk.String()))
	}
	p.move()
	if !p.match(lexical.LPAREN) {
		p.Error(fmt.Sprintf("dowhilestat err: expected '(', but got %s", p.tk.String()))
	}
	p.move()
	cond := p.altexpr()
	if !p.match(lexical.RPAREN) {
		p.Error(fmt.Sprintf("dowhilestat err: expected '(', but got %s", p.tk.String()))
	}
	p.move()
	if !p.match(lexical.SEMICOLON) {
		p.Error(fmt.Sprintf("dowhilestat err: expected ';', but got %s", p.tk.String()))
	}
	p.move()
	_, lnum, cnum = p.lexer.GetPosition()
	table.Symtab.Leave(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
	table.GenDoWhileTail(_do, _exit, cond)
}

// <ifstat> -> rsv_if lparen <expr> rparen <block> <elsestat>
func (p *Parser) ifstat() {
	_, lnum, cnum := p.lexer.GetPosition()
	table.Symtab.Enter(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
	if !p.match(lexical.KW_IF) {
		p.Error(fmt.Sprintf("ifstat err: expected KW_IF, but got %s", p.tk.String()))
	}
	p.move()
	if !p.match(lexical.LPAREN) {
		p.Error(fmt.Sprintf("dowhilestat err: expected '(', but got %s", p.tk.String()))
	}
	p.move()
	_else, _exit := table.NewLabelInst(), table.NewLabelInst()
	cond := p.expr()
	table.GenIfHead(cond, _else)
	if !p.match(lexical.RPAREN) {
		p.Error(fmt.Sprintf("dowhilestat err: expected ')', but got %s", p.tk.String()))
	}
	p.move()
	p.block() //TODO:思考，block中的跳转语句不需要管，if的语义是，if的真block执行完之后，跳转到_exit
	_, lnum, cnum = p.lexer.GetPosition()
	table.Symtab.Leave(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
	if p.match(lexical.KW_ELSE) {
		table.GenElseHead(_exit, _else)
		p.elsestat()
		table.GenElseTail(_exit)
	} else {
		table.GenIfTail(_else)
	}
}

// <switchstat> -> rsv_switch lparen <expr> rparen lbrace <casestat> rbrace
func (p *Parser) switchstat() {
	_, lnum, cnum := p.lexer.GetPosition()
	table.Symtab.Enter(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
	if !p.match(lexical.KW_SWITCH) {
		p.Error(fmt.Sprintf("switchstat err: expected KW_SWITCH, but got %s", p.tk.String()))
	}
	p.move()
	if !p.match(lexical.LPAREN) {
		p.Error(fmt.Sprintf("switchstat err: expected '(', but got %s", p.tk.String()))
	}
	p.move()
	cond := p.expr()
	_exit := table.NewLabelInst()
	table.Push(nil, _exit) //<=> GenSwitchHead(_exit)
	if !p.match(lexical.RPAREN) {
		p.Error(fmt.Sprintf("switchstat err: expected ')', but got %s", p.tk.String()))
	}
	p.move()
	if !p.match(lexical.LBRACE) {
		p.Error(fmt.Sprintf("switchstat err: expected '{', but got %s", p.tk.String()))
	}
	p.move()
	p.casestat(cond)
	if !p.match(lexical.RBRACE) {
		p.Error(fmt.Sprintf("switchstat err: expected '}', but got %s", p.tk.String()))
	}
	p.move()
	_, lnum, cnum = p.lexer.GetPosition()
	table.Symtab.Leave(fmt.Sprintf("line:%d, col:%d token:%s", lnum, cnum, p.tk.String()))
	table.GenSwitchTail(_exit) //put _exit and Pop()
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
func (p *Parser) casestat(cond *table.Var) {
	if p.match(lexical.KW_CASE) {
		p.move()
		lb := p.caselabel()
		_case_exit := table.NewLabelInst()
		table.GenCaseHead(_case_exit, cond, lb)
		if !p.match(lexical.COLON) {
			p.Error(fmt.Sprintf("casestat err: expected ',', but got %s", p.tk.String()))
		}
		p.move()
		p.subprogram()
		table.GenCaseTail(_case_exit)
		p.casestat(cond)
	} else if p.match(lexical.KW_DEFAULT) {
		p.move()
		if !p.match(lexical.COLON) {
			p.Error(fmt.Sprintf("casestat err: expected ',', but got %s", p.tk.String()))
		}
		p.move()
		p.subprogram()
	} else {
		p.Error(fmt.Sprintf("casestat err: expected 'case' or 'default', but got %s", p.tk.String()))
	}
}

// <caselabel> -> <literal>
func (p *Parser) caselabel() *table.Var {
	return p.literal()
}

// <literal>	->	number | string | character
func (p *Parser) literal() *table.Var {
	if !p.match(lexical.NUM) && !p.match(lexical.STR) && !p.match(lexical.CHAR) {
		p.Error(fmt.Sprintf("literal err: expected NUM, STR or CHAR, but got %s", p.tk.String()))
	}
	v := table.NewLiteralVar(p.tk)
	if p.tk.TokenTyp() == lexical.STR {
		table.Symtab.AddStr(v)
	} else {
		table.Symtab.AddVar(v)
	}
	p.move()
	return v
}

// <altexpr> ->	<expr> | ^
func (p *Parser) altexpr() *table.Var {
	if p.matchExprFirst() {
		return p.expr()
	}
	return table.Void
}

func (p *Parser) match(typ lexical.TokenType) bool {
	return p.tk.TokenTyp() == typ
}

func (p *Parser) move() {
	p.tk = p.lexer.NextToken()
}

type OP_TYPE int

const (
	_                 = iota
	OP_ASSIGN OP_TYPE = iota
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_OR
	OP_MOD
	OP_INC
	OP_DEC
	OP_NOT
	OP_LEA
	OP_AND
	OP_GT
	OP_GE
	OP_LT
	OP_LE
	OP_EQU
	OP_NEQU
	OP_COMMA
	OP_COLON
	OP_SEMICOLON
	OP_LPAREN
	OP_RPAREN
	OP_LBARCE
	OP_RBRACE
)
