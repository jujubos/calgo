package table

import (
	"calgo/lexical"
	"fmt"
)

var lbnum = 0

func GenLb() string {
	lbnum++
	lb := ".L"
	return lb + string(lbnum)
}
func GenFunHead(fun *Fun) {
	fun.EnterScope()
	Symtab.AddInst(NewEntryInst(fun))
	fun.SetReturnPoint(NewLabelInst()) //后面的return语句都将引用这个标签
}

func GenFunTail(fun *Fun) {
	Symtab.AddInst(fun.GetReturnPoint())
	Symtab.AddInst(NewExitInst(fun))
	fun.LeaveScope()
}

func GenReturn(retv *Var) {
	if retv == nil {
		return
	}
	fun := Symtab.Curfun
	if retv.Typ == lexical.KW_VOID && fun.Typ != lexical.KW_VOID ||
		retv.Typ != lexical.KW_VOID && fun.Typ == lexical.KW_VOID {
		Error("返回值类型不匹配")
	}
	if fun.Typ == lexical.KW_VOID {
		Symtab.AddInst(NewRetInst(fun.GetReturnPoint()))
	} else {
		if retv.IsRef() {
			retv = GenAssign1(retv)
		}
		Symtab.AddInst(NewRetvInst(retv, fun.GetReturnPoint()))
	}
}

func GenPtr(v *Var) *Var {
	if v.IsBase() {
		Error("基本类型不支持*操作")
	}
	tmp := NewTmpVar(Symtab.ScopePath, v.Typ, false)
	tmp.IsLeft = true
	tmp.Ptr = v
	Symtab.AddVar(tmp)
	return tmp
}

// tmp = &v
// if v is *p, then p is what we want
func GenLea(v *Var) *Var {
	if !v.IsLeft {
		Error("右值不支持&操作")
	}
	if v.IsRef() {
		return v.Ptr
	}
	tmp := NewTmpVar(Symtab.ScopePath, v.Typ, true)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_LEA, tmp, v, nil))
	return tmp
}

// 复制v: tmp = v
// if v is ref, then tmp = *v
func GenAssign1(v *Var) *Var {
	//为何要拷贝一个?
	tmp := CopyVar(Symtab.ScopePath, v)
	if !v.IsRef() {
		Symtab.AddInst(NewInst(OP_AS, tmp, v, nil))
	} else {
		Symtab.AddInst(NewInst(OP_GET, tmp, v, nil))
	}
	Symtab.AddVar(tmp)
	return tmp
}

func GenAssign2(lval *Var, rval *Var) *Var {
	if !lval.IsLeft {
		Error("不可以对右值赋值")
	}
	if !TypeCheck(lval, rval) {
		Error("类型不兼容，不可赋值")
	}
	if rval.IsRef() {
		rval = GenPtr(rval)
	}
	if lval.IsRef() {
		Symtab.AddInst(NewInst(OP_SET, rval, lval.Ptr, nil))
	} else {
		NewInst(OP_AS, lval, rval, nil)
	}
	return lval
}

func GenTwoOp(op Operator, lvar, rvar *Var) *Var {
	if lvar.IsVoid() || rvar.IsVoid() {
		Error("参与表达式运算的变量类型不能为void")
	}
	if op == OP_AS {
		return GenAssign2(lvar, rvar)
	}
	if lvar.IsRef() {
		lvar = GenAssign1(lvar)
	}
	if rvar.IsRef() {
		rvar = GenAssign1(rvar)
	}
	switch op {
	case OP_OR:
		return GenOr(lvar, rvar)
	case OP_AND:
		return GenAnd(lvar, rvar)
	case OP_EQU:
		return GenEQU(lvar, rvar)
	case OP_NEQU:
		return GenNEQU(lvar, rvar)
	case OP_ADD:
		return GenAdd(lvar, rvar)
	case OP_SUB:
		return GenSub(lvar, rvar)
	}
	if !lvar.IsBase() || rvar.IsBase() {
		Error(fmt.Sprintf("该类型不支持这种运算:%d", op))
	}
	switch op {
	case OP_GT:
		return GenGT(lvar, rvar)
	case OP_GE:
		return GenGE(lvar, rvar)
	case OP_LT:
		return GenLT(lvar, rvar)
	case OP_LE:
		return GenLE(lvar, rvar)
	case OP_MUL:
		return GenMul(lvar, rvar)
	case OP_DIV:
		return GenDiv(lvar, rvar)
	case OP_MOD:
		return GenMod(lvar, rvar)
	}
	Error(fmt.Sprintf("不支持的双目运算:%d", op))
	return nil
}

/*
翻译加法表达式
指针和int相加: p + 1, 1 + p, p + i， 翻译为:
基本类型之间相加,翻译为:
*/
func GenAdd(lvar, rvar *Var) *Var {
	var tmp *Var
	if lvar.IsPtr && rvar.IsBase() {
		tmp = CopyVar(Symtab.ScopePath, lvar)
		rvar = GenMul(rvar, GetStep(lvar))
	} else if lvar.IsBase() && rvar.IsPtr {
		tmp = CopyVar(Symtab.ScopePath, rvar)
		lvar = GenMul(rvar, GetStep(rvar))
	} else if lvar.IsBase() && rvar.IsBase() {
		tmp = NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	} else {
		Error("GenAdd:类型不支持")
	}
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_ADD, tmp, lvar, rvar))
	return tmp
}

/*
翻译减法表达式
指针和int相减: p - 1, p - i， 翻译为:， 注意：不支持 i - p
基本类型之间相减,翻译为:
*/
func GenSub(lvar, rvar *Var) *Var {
	var tmp *Var
	if lvar.IsPtr && rvar.IsBase() {
		tmp = CopyVar(Symtab.ScopePath, lvar)
		rvar = GenMul(rvar, GetStep(lvar))
	} else if lvar.IsBase() && rvar.IsPtr {
		Error("不支持 i - p")
	} else if lvar.IsBase() && rvar.IsBase() {
		tmp = NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	} else {
		Error("GenAdd:类型不支持")
	}
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_SUB, tmp, lvar, rvar))
	return tmp
}

/*
翻译乘法表达式
注意，GenMul只在GenTwoOp中被调用，调用前已经确保lvar，rvar为基本类型
*/
func GenMul(lvar, rvar *Var) *Var {
	tmp := NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_MUL, tmp, lvar, rvar))
	return tmp
}

func GenOr(lvar, rvar *Var) *Var {
	tmp := NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_OR, tmp, lvar, rvar))
	return tmp
}

func GenAnd(lvar, rvar *Var) *Var {
	tmp := NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_AND, tmp, lvar, rvar))
	return tmp
}

func GenEQU(lvar, rvar *Var) *Var {
	tmp := NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_EQU, tmp, lvar, rvar))
	return tmp
}

func GenNEQU(lvar, rvar *Var) *Var {
	tmp := NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_NEQU, tmp, lvar, rvar))
	return tmp
}

func GenGT(lvar, rvar *Var) *Var {
	tmp := NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_GT, tmp, lvar, rvar))
	return tmp
}

func GenGE(lvar, rvar *Var) *Var {
	tmp := NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_GE, tmp, lvar, rvar))
	return tmp
}

func GenLT(lvar, rvar *Var) *Var {
	tmp := NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_LT, tmp, lvar, rvar))
	return tmp
}

func GenLE(lvar, rvar *Var) *Var {
	tmp := NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_LE, tmp, lvar, rvar))
	return tmp
}

func GenDiv(lvar, rvar *Var) *Var {
	tmp := NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_DIV, tmp, lvar, rvar))
	return tmp
}

func GenMod(lvar, rvar *Var) *Var {
	tmp := NewTmpVar(Symtab.ScopePath, lexical.KW_INT, false)
	Symtab.AddVar(tmp)
	Symtab.AddInst(NewInst(OP_MOD, tmp, lvar, rvar))
	return tmp
}

func GetStep(v *Var) *Var {
	if v.IsBase() {
		return One
	}
	if v.Typ == lexical.KW_CHAR {
		return One
	} else if v.Typ == lexical.KW_INT {
		return Four
	}
	Error("GetStep:void不能参与加法运算")
	return nil
}
