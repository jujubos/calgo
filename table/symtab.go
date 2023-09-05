package table

import (
	"calgo/lexical"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/olekukonko/tablewriter"
	"log"
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

/* 变量声明、定义和初始化的语义分析 */
func (s *SymTable) AddVar(varr *Var) {
	if varr == nil {
		return
	}
	/* 是否重复声明或定义 */
	for _, v := range s.Vartab[varr.Name] {
		if varr.Name[0] != '<' && v.ScopeID() == varr.ScopeID() {
			Error(fmt.Sprintf("同一作用域下存在同名变量: %s", v.Name))
		}
	}
	s.Vartab[varr.Name] = append(s.Vartab[varr.Name], varr)
	//是否需要产生初始化指令
	if !varr.Externed {
		/* 如果不是常量，生成'OP_DEC varr' */
		/* 如果是全局变量，在符号表中记录初值 */
		/* 如果是局部变量，且初始化表达式为常量表达式，则在符号表中记录初值 */
		/* 如果是局部变量，且初始化表达式不是常量表达式，则生成赋值指令：'varr = varr.initdata'(varr.initdata是varr的初始化表达式的值对应的符号) */
		flag := GenVarInit(varr)
		/* 如果是局部变量，记录局部变量相关信息（例如：偏移量），以及更新当前所处函数的栈桢的状态。 */
		if s.Curfun != nil && flag {
			s.Curfun.Locate(varr)
		}
	}
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

func (s *SymTable) PrintVarTab() {
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

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (s *SymTable) PrintWithoutIntercode() {
	jd, err := json.Marshal(s)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(jd))
}

func (s *SymTable) PrintInterCode() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Op", "Result", "Arg1", "Arg2", "Label", "Target", "Fun"})
	for _, f := range s.Funtab {
		table.ClearRows()
		var data [][]string
		for _, inst := range f.Intercode {
			data = append(data, inst.SliceString())
		}
		table.ClearFooter()
		table.SetFooter([]string{f.Name, "", "", "", "", "", ""})
		table.AppendBulk(data)
		table.Render()
	}
}

func (s *SymTable) PrintInterCodeOf(fun_name string) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Op", "Result", "Arg1", "Arg2", "Label", "Target", "Fun"})
	fun, ok := s.Funtab[fun_name]
	if !ok {
		return fmt.Errorf("function is not found:\"%s\"\n", fun_name)
	}
	table.ClearRows()
	var data [][]string
	for _, inst := range fun.Intercode {
		data = append(data, inst.SliceString())
	}
	table.ClearFooter()
	table.SetFooter([]string{fun.Name, "", "", "", "", "", ""})
	table.AppendBulk(data)
	table.Render()

	return nil
}

func (s *SymTable) SaveObjCode() {
	for _, f := range s.Funtab {
		Emit(fmt.Sprintf("----------%s----------", f.Name))
		for _, inst := range f.Intercode {
			inst.ToX86Asm()
		}
	}
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
		s.Funtab[fun.Name] = fun
		fun.Externed = false
		s.FunList = append(s.FunList, fun.Name)
		s.Curfun = fun
	} else {
		//之前必须是声明
		if !f.Externed {
			Error(fmt.Sprintf("<%s>:函数重定义", fun.Name))
		} else {
			//定义要和声明匹配
			if !f.Match(fun) {
				Error(fmt.Sprintf("<%s>:函数定义和声明不匹配", fun.Name))
			}
			f.Externed = false
			s.Curfun = f
		}
	}
	GenFunHead(s.Curfun)
}

func (s *SymTable) EndDefFun() {
	GenFunTail(s.Curfun)
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
		//原作:
		/*
			void SymTab::addInst(InterInst*inst)
			{
				if(curFun)curFun->addInst(inst);
				else delete inst; //???
			}
		*/
		//Error(fmt.Sprintf("AddInst %v: s.Curfun is nil", inst.SliceString()))
	}
}

func (s *SymTable) GetGlbVars() []*Var {
	var res []*Var
	for name, vars := range s.Vartab {
		if name[0] == '<' {
			continue
		}
		for _, v := range vars {
			if len(v.ScopePath) == 1 {
				res = append(res, v)
			}
		}
	}
	return res
}

/*
生成数据段。【注意，数据段中只存放全局变量和静态变量（这个语言没有定义静态变量的语法），不需要考虑局部变量】
遍历所有全局符号：
 1. global声明
 2. 如果是externed，只需global声明
 3. 输出符号名
 4. 如果是数组，输出times xxx
 5. 如果是char且不是指针，输出db，否则输出dd
 6. 如果有初始化：如果是基本类型，输出value；如果是指针类型，输出ptrval。
 7. 没有初始化，默认值为0
*/
func (s *SymTable) GenData() {
	glbvars := s.GetGlbVars()
	for _, v := range glbvars {
		EmitAsm(fmt.Sprintf("global %s", v.Name))
		if v.Externed { //extern声明的变量，只需要生成global声明
			continue
		}
		s := ""
		s += fmt.Sprintf("\t%s ", v.Name)
		typsize := 4
		if v.Typ == lexical.KW_CHAR && !v.IsPtr {
			typsize = 1
		}
		if v.IsArray {
			s += fmt.Sprintf("times %d ", v.Size/int64(typsize))
		}
		if typsize == 1 {
			s += "db "
		} else {
			s += "dd "
		}
		if v.inited {
			if v.IsBase() {
				s += fmt.Sprintf("%d", v.GetVal())
			} else { //字符指针
				s += v.PtrVal
			}
		} else {
			s += "0"
		}
		EmitAsm(s)
	}
	for _, strvar := range s.Strtab {
		EmitAsm(fmt.Sprintf("%s db %s", strvar.Name, strvar.GenRawStr()))
	}
}

func (s *SymTable) GenAsm() {
	EmitAsm("section .data")
	s.GenData()
	Emit("section .text")
	for _, f := range s.Funtab {
		EmitAsm(fmt.Sprintf("global %s", f.Name))
		if f.Externed { //没有函数定义，只需要生成global声明
			continue
		}
		EmitAsm(fmt.Sprintf("%s:", f.Name))
		for _, inst := range f.Intercode {
			inst.ToX86Asm()
		}
	}
}

func EmitAsm(s string) {
	_, err := AsmFile.WriteString(s + "\n")
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}
}

var Symtab = &SymTable{
	Funtab:    make(map[string]*Fun),
	Vartab:    make(map[string][]*Var),
	Strtab:    make(map[string]*Var),
	ScopePath: []int{0},
}

var AsmFile *os.File

func init() {

}
