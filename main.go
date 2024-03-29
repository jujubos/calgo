package main

import (
	"calgo/asm"
	"calgo/syntax"
	"calgo/table"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type InterCodeSpec []string

func (i *InterCodeSpec) Set(s string) error {
	*i = strings.Split(s, ",")
	return nil
}

func (i *InterCodeSpec) String() string {
	var res = "<"
	for _, p := range *i {
		res = res + p
		res += ","
	}
	return res + ">"
}

func create_file(fpath string) {
	dirPath := filepath.Dir(fpath)

	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		fmt.Println("无法创建目录:", err)
		return
	}

	file, err := os.Create(fpath)
	if err != nil {
		fmt.Println("无法创建文件:", err)
		return
	}
	defer file.Close()
}

func main() {
	var err error
	var intercode_spec InterCodeSpec
	sourcefile := flag.String("sourcefile", "./demo/intercode.demo", "source file")
	asmfile := flag.String("asmfile", "./out/code.asm", "assembly file")
	exefile := flag.String("exefile", "./out/elf_reloc.o", "relocatable object file")
	flag.Var(&intercode_spec, "print_intercode", "print intercode")
	flag.Parse()
	/* make sure the output files exists(sourcefile is user's duty)  */
	create_file(*asmfile)
	create_file(*exefile)
	/* 编译阶段（已完成） */
	parser := syntax.NewParser(*sourcefile)
	parser.Parse()
	table.AsmFile, err = os.OpenFile(*asmfile, os.O_CREATE|syscall.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}
	table.Symtab.GenAsm()
	/* print intercode of function specified by 'print_intercode' */
	for _, fun_name := range intercode_spec {
		err = table.Symtab.PrintInterCodeOf(fun_name)
		if err != nil {
			panic(err)
		}
	}

	/* 汇编阶段（已完成） */
	asmparser := asm.NewParser(*asmfile)
	asmparser.Parse()
	asm.Symtab.ExportSyms()
	asm.EXEFILE, err = os.OpenFile(*exefile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
		panic("open elf.out fail")
	}
	asm.ELFOBJ.WriteElf()

	/* 链接阶段（未完成）*/
	//linker := link.NewLinker()
	//args := os.Args
	//for _, arg := range args[1:] {
	//	linker.AddELF(arg)
	//}
	//linker.Link()
}
