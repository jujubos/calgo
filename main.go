package main

import (
	"calgo/syntax"
)

func main() {
	filename := "./demo/test.c"
	//lexer := lexical.NewLexer(filename)
	//for {
	//	token := lexer.NextToken()
	//	fmt.Println(token.String())
	//	if token == &lexical.EOF_TOKEN || token == &lexical.ERR_TOKEN {
	//		return
	//	}
	//}
	parser := syntax.NewParser(filename)
	parser.Parse()
}
