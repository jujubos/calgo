package table

import "calgo/lexical"

type Fun struct {
	externed bool
	typ      lexical.TokenType
	name     string
	paraVar  []*Var
	maxDepth int
	curEsp   int
	scopeEsp []int
	//interCode
	//returnPoint
}
