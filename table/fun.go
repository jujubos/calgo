package table

import (
	"calgo/lexical"
)

type Fun struct {
	Name        string
	Externed    bool
	Typ         lexical.TokenType
	ParaVar     []*Var
	MaxDepth    int
	CurEsp      int
	ScopeEsp    []int
	Intercode   []*InterInst `json:"-"`
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
	f.Intercode = append(f.Intercode, inst)
}

func (f *Fun) SetReturnPoint(retp *InterInst) {
	f.returnPoint = retp
}

func (f *Fun) GetReturnPoint() *InterInst {
	return f.returnPoint
}

func (f *Fun) Locate(v *Var) {
	size := v.Size
	size += (4 - size%4) % 4
	f.ScopeEsp[len(f.ScopeEsp)-1] += int(size)
	f.CurEsp += int(size)
	v.Offset = int64(-f.CurEsp)
}
