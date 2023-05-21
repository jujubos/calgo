package main

import (
	"calgo/syntax"
	"calgo/table"
)

func main() {
	filename := "./demo/table.demo"
	//lexer := lexical.NewLexer(filename)
	//for {
	//	token := lexer.NextToken()
	//	fname, lnum, cnum := lexer.GetPosition()
	//	fmt.Println(fmt.Sprintf("%s %d %d : %s", fname, lnum, cnum, token.String()))
	//	if token.TokenTyp() == lexical.ERR || token.TokenTyp() == lexical.EOF {
	//		return
	//	}
	//}
	parser := syntax.NewParser(filename)
	parser.Parse()
	table.Symtab.Print()
}
