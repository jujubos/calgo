package main

import (
	"calgo/lexical"
	"fmt"
)

func main() {
	filename := "./demo/lex.demo"
	lexer := lexical.NewLexer(filename)
	for {
		token := lexer.NextToken()
		fmt.Println(token.String())
		if token == &lexical.EOF_TOKEN || token == &lexical.ERR_TOKEN {
			return
		}
	}
}
