package table

import (
	"calgo/lexical"
)

type Fun struct {
	Externed    bool
	Typ         lexical.TokenType
	Name        string
	ParaVar     []*Var
	MaxDepth    int
	CurEsp      int
	ScopeEsp    []int
	interCode   []*InterInst
	returnPoint *InterInst
}

func NewFun(ext bool, typ lexical.TokenType, name string, paralist []*Var) *Fun {
	fun := &Fun{
		Externed: ext,
		Typ:      typ,
		Name:     name,
		ParaVar:  paralist,
		CurEsp:   0,
		MaxDepth: 0,
	}
	//参数从8字节位置开始存放, 固定4字节大小，
	offset := 8
	for _, f := range fun.ParaVar {
		f.Offset = int64(offset)
		offset = offset + 4
	}
	return fun
}

func (f *Fun) Match(fun *Fun) bool {
	if f.Name != fun.Name || len(f.ParaVar) != len(fun.ParaVar) {
		return false
	}
	if len(fun.ParaVar) <= len(f.ParaVar) {
		for i := range f.ParaVar {
			p1, p2 := f.ParaVar[i], fun.ParaVar[i]
			if TypeCheck(p1, p2) {
				if p1.Typ != p2.Typ {
					Warning("两个函数的参数类型不同")
				}
			} else {
				Error("两个函数的参数类型不兼容")
			}
		}
	}
	return true
}

func (f *Fun) MatchArgs(args []*Var) bool {
	if len(f.ParaVar) != len(args) {
		return false
	}
	if len(args) <= len(f.ParaVar) {
		for i := range f.ParaVar {
			p1, p2 := f.ParaVar[i], args[i]
			if !TypeCheck(p1, p2) {
				return false
			}
		}
	}
	return true
}

// Finished
// 变量声明初始化部分由setInit处理
func (v *Var) SetInit() bool {
	vinit := v.initData
	if vinit == nil {
		return false
	}
	v.inited = false
	if v.Externed {
		Error("声明不允许初始化")
	} else if !TypeCheck(v, vinit) {
		Error("类型不兼容")
	} else if vinit.Literal {
		v.inited = true
		if vinit.IsArray { //数组字面量，只能是字符串字面量，如"abc"
			v.PtrVal = vinit.Name
		} else { //整数，字符
			var s int64
			if vinit.Typ == lexical.CHAR {
				s = int64(vinit.CharVal)
			} else {
				s = vinit.IntVal
			}
			if v.Typ == lexical.CHAR {
				v.CharVal = byte(s)
			} else {
				v.IntVal = s
			}
		}
	} else { //非字面量，那就是变量
		if len(v.ScopePath) == 1 { //全局变量
			Error("全局变量初始化必须是常量")
		} else { //非全局变量，那就是局部变量
			return true
		}
	}
	return false
}

func (f *Fun) EnterScope() {
	f.ScopeEsp = append(f.ScopeEsp, 0)
}

func (f *Fun) LeaveScope() {
	if f.CurEsp > f.MaxDepth {
		f.MaxDepth = f.CurEsp
	}
	f.CurEsp -= f.ScopeEsp[len(f.ScopeEsp)-1]
	f.ScopeEsp = f.ScopeEsp[:len(f.ScopeEsp)-1]
}

func (f *Fun) AddInst(inst *InterInst) {
	f.interCode = append(f.interCode, inst)
}

func (f *Fun) SetReturnPoint(retp *InterInst) {
	f.returnPoint = retp
}

func (f *Fun) GetReturnPoint() *InterInst {
	return f.returnPoint
}
