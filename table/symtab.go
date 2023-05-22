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
	//声明顺序记录
	FunList []string
	VarList []string
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
	/* TODO
	if (ir) {
	flag := ir.GenVarInit(var)
	if (s.CurFun != nil && flag) { s.CurFun.Locate(var) }
	}
	*/
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

func (s *SymTable) DecFun(fun *Fun) {
	if f, ok := s.Funtab[fun.Name]; ok {
		if !f.Match(fun) {
			Error("函数声明冲突")
		}
	} else {
		s.Funtab[fun.Name] = fun
	}
	fun.Externed = true
}

func (s *SymTable) DefFun(fun *Fun) {
	if fun.Externed {
		Error("extern不允许出现在定义")
	}
	if f, ok := s.Funtab[fun.Name]; !ok {
		s.Funtab[fun.Name] = f
		f.Externed = true
		s.FunList = append(s.FunList, fun.Name)
	} else {
		//之前必须是声明
		if !fun.Externed {
			Error(fmt.Sprintf("<%s>:函数重定义", fun.Name))
		} else {
			//定义要和声明匹配
			if !f.Match(fun) {
				Error(fmt.Sprintf("<%s>:函数定义和声明不匹配", fun.Name))
			}
			fun.Externed = false
		}
	}
	s.Curfun = fun
	//TODO:产生函数入口
	//ir.GenFunHead(s.CurFun)
}

func (s *SymTable) EndDefFun() {
	//TODO:产生函数出口
	//ir.GenFunTail(s.CurFun)
	s.Curfun = nil
}

func (s *SymTable) GetFun(name string, readarglist []*Var) *Fun {
	if f, ok := s.Funtab[name]; !ok {
		Error(fmt.Sprintf("<%s>:函数未声明", name))
	} else {
		if !f.MatchArgs(readarglist) {
			Error(fmt.Sprintf("<%s>:形参与实参不匹配", name))
		} else {
			return f
		}
	}
	return nil
}

func (s *SymTable) AddInst(inst *InterInst) {
	if s.Curfun != nil {
		s.Curfun.AddInst(inst)
	} else {
		Error("s.Curfun is nil, can't call AddInst")
	}
}

var Symtab = &SymTable{
	Funtab:    make(map[string]*Fun),
	Vartab:    make(map[string][]*Var),
	Strtab:    make(map[string]*Var),
	ScopePath: []int{0},
}
