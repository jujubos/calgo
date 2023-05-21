package table

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
)

type SymTable struct {
	Funtab    map[string]*Fun
	Vartab    map[string][]*Var
	Strtab    map[string]*Var
	ScopePath []int
	ScopeID   int
	Curfun    *Fun
}

func (s *SymTable) Enter(info string) {
	s.ScopeID++
	s.ScopePath = append(s.ScopePath, s.ScopeID)
	fmt.Printf("after enter, %s %v\n", info, s.ScopePath)
}

func (s *SymTable) Leave(info string) {
	s.ScopePath = s.ScopePath[:len(s.ScopePath)-1]
	fmt.Printf("after leave, %s %v\n", info, s.ScopePath)
}

func (s *SymTable) AddVar(varr *Var) {
	if varr == nil {
		return
	}
	for _, v := range s.Vartab[varr.Name] {
		if varr.Name[0] != '<' && v.ScopeID() == varr.ScopeID() {
			Error("同一作用域下存在同名变量")
		}
	}
	s.Vartab[varr.Name] = append(s.Vartab[varr.Name], varr)
}

func (s *SymTable) GetVar(name string) *Var {
	list := s.Vartab[name]
	maxl := 0
	var rs *Var
	for _, v := range list {
		l := len(v.ScopePath)
		if l <= len(s.ScopePath) && v.ScopePath[l-1] == s.ScopePath[l-1] && maxl < l {
			maxl = l
			rs = v
		}
	}
	if rs == nil {
		Error("变量未声明（定义）")
	}
	return rs
}

func (s *SymTable) AddStr(varr *Var) {
	name := varr.Name
	if _, ok := s.Strtab[name]; !ok {
		s.Strtab[name] = varr
	}
}

func (s *SymTable) Print() {
	var vtab [][]string
	for _, vs := range s.Vartab {
		for _, v := range vs {
			vtab = append(vtab, v.Row())
		}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Name",
		"Type",
		"Literal",
		"initData",
		"ScopePath",
		"Externed",
		//"IsPtr",
		//"IsArray",
		//"ArraySize",
		//"IsLeft",
		"inited",
		"IntVal",
		"CharVal",
		"StrVal",
		"PtrVal",
		//"Ptr",
		//"Size",
		//"Offset",
	})

	for _, v := range vtab {
		table.Append(v)
	}
	table.Render() // Send output
}

var Symtab = &SymTable{
	Funtab:    make(map[string]*Fun),
	Vartab:    make(map[string][]*Var),
	Strtab:    make(map[string]*Var),
	ScopePath: []int{0},
}
