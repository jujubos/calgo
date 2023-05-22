package table

import (
	"calgo/lexical"
	"fmt"
	"strings"
)

type Var struct {
	Literal   bool
	ScopePath []int
	Externed  bool
	Typ       lexical.TokenType
	Name      string
	IsPtr     bool
	IsArray   bool
	ArraySize int64
	IsLeft    bool
	initData  *Var
	inited    bool
	IntVal    int64
	CharVal   byte
	StrVal    string //字符串常量值
	PtrVal    string //字符指针值
	Ptr       *Var   //?
	Size      int64
	Offset    int64
}

// 非数组、非指针
func NewVar(sp []int, ext bool, t lexical.TokenType, ptr bool, name string, init *Var) *Var {
	v := &Var{
		ScopePath: sp,
		Externed:  ext,
	}
	v.setType(t)
	v.setPtr(ptr)
	v.setName(name)
	v.initData = init
	return v
}

func (v *Var) setPtr(ptr bool) {
	v.IsPtr = ptr
	if v.IsPtr && !v.Externed {
		v.Size = 4
	}
}

func NewArrayVar(sp []int, ext bool, typ lexical.TokenType, name string, len int64) *Var {
	v := &Var{
		ScopePath: sp,
		Externed:  ext,
		Name:      name,
	}
	v.setType(typ)
	v.setArray(len)
	return v
}

// 字面量
func NewLiteralVar(tk lexical.Token) *Var {
	v := &Var{}
	v.Literal = true
	v.IsLeft = false
	switch tk.TokenTyp() {
	case lexical.NUM:
		v.setType(lexical.KW_INT)
		v.setName("<int>")
		v.IntVal = tk.(*lexical.TNUM).Value
	case lexical.CHAR:
		v.setType(lexical.KW_CHAR)
		v.setName("<char>")
		v.CharVal = tk.(*lexical.TCHAR).Value
	case lexical.STR:
		v.setType(lexical.KW_CHAR) //???
		v.setName(GenLb())
		v.StrVal = tk.(*lexical.TSTR).Value
		v.setArray(int64(len(v.StrVal) + 1))
	}
	return v
}

func NewIntVar(val int) *Var {
	v := &Var{}
	v.setName("<int>")
	v.setType(lexical.KW_INT)
	v.IntVal = int64(val)
	v.Literal = true
	return v
}

func NewTmpVar(sp []int, typ lexical.TokenType, isptr bool) *Var {
	v := &Var{
		ScopePath: sp,
	}
	v.setType(typ)
	v.setPtr(isptr)
	v.setName("")
	v.IsLeft = false
	return v
}

func CopyVar(sp []int, v *Var) *Var {
	tmp := &Var{ScopePath: sp}
	tmp.setType(v.Typ)
	tmp.setPtr(v.IsPtr || v.IsArray)
	tmp.setName("")
	tmp.IsLeft = false
	return tmp
}

func NewVoidVar() *Var {
	v := &Var{}
	v.setName("<void>")
	v.Typ = lexical.KW_VOID
	v.IsPtr = true
	return v
}

func (v *Var) IsVoid() bool {
	return v.Typ == lexical.KW_VOID
}

func (v *Var) setArray(length int64) {
	if length <= 0 {
		Error("array len <= 0")
	}
	v.IsArray = true
	v.IsLeft = false
	v.ArraySize = length
	if !v.Externed {
		v.Size = v.ArraySize * v.Size
	}
}

func (v *Var) setType(typ lexical.TokenType) {
	v.Typ = typ
	if v.Typ == lexical.KW_VOID {
		Error("变量类型不能是void")
	}
	if !v.Externed {
		if v.Typ == lexical.KW_INT {
			v.Size = 4
		} else if typ == lexical.KW_CHAR {
			v.Size = 1
		}
	}
}

func (v *Var) setName(name string) {
	if name == "" {
		name = GenLb()
	}
	v.Name = name
}

func (v *Var) ScopeID() int {
	return v.ScopePath[len(v.ScopePath)-1]
}

func (v *Var) Row() []string {
	if v == nil {
		return nil
	}
	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s,", v.Name))
	builder.WriteString(fmt.Sprintf("%s,", lexical.TypeTable[v.Typ]))
	builder.WriteString(fmt.Sprintf("%v,", v.Literal))
	if v.initData != nil {
		builder.WriteString(fmt.Sprintf("%v,", v.initData.Name))
	} else {
		builder.WriteString("No InitData,")
	}
	builder.WriteString(fmt.Sprintf("%v,", v.ScopePath))
	builder.WriteString(fmt.Sprintf("%v,", v.Externed))
	//builder.WriteString(fmt.Sprintf("%v,", v.IsPtr))
	//builder.WriteString(fmt.Sprintf("%v,", v.IsArray))
	//builder.WriteString(fmt.Sprintf("%d,", v.ArraySize))
	//builder.WriteString(fmt.Sprintf("%v,", v.IsLeft))
	builder.WriteString(fmt.Sprintf("%v,", v.inited))
	builder.WriteString(fmt.Sprintf("%d,", v.IntVal))
	builder.WriteString(fmt.Sprintf("%c,", v.CharVal))
	builder.WriteString(fmt.Sprintf("%s,", v.StrVal))
	builder.WriteString(fmt.Sprintf("%s,", v.PtrVal))
	//builder.WriteString(fmt.Sprintf("%v,", v.Ptr))
	//builder.WriteString(fmt.Sprintf("%d,", v.Size))
	//builder.WriteString(fmt.Sprintf("%d,", v.Offset))

	return strings.Split(builder.String(), ",")
}

func (v *Var) IsBase() bool {
	return !v.IsArray && !v.IsPtr
}

func (v *Var) IsRef() bool {
	return v.Ptr != nil
}

var Void = NewVoidVar()
var One = NewIntVar(1)
var Four = NewIntVar(1)
