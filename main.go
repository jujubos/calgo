package main

import (
	"calgo/asm"
)

func main() {
	filename := "./asm/test.asm"
	//parser := syntax.NewParser(filename)
	//parser.Parse()
	//table.Symtab.GenAsm()
	parser := asm.NewParser(filename)
	parser.Parse()
	asm.Symtab.ExportSyms()
	asm.ELFOBJ.WriteElf()
}
