package table

import "fmt"

type Operator int

type InterInst struct {
	Label  string
	Op     Operator
	Result *Var
	Arg1   *Var
	Arg2   *Var
	Fun    *Fun
	Target *InterInst `json:"-"`
}

func NewInst(op Operator, res *Var, arg1 *Var, arg2 *Var) *InterInst {
	return &InterInst{
		Op:     op,
		Result: res,
		Arg1:   arg1,
		Arg2:   arg2,
	}
}

func NewLabelInst() *InterInst {
	return &InterInst{
		Label: GenLb(),
	}
}

func NewEntryInst(fun *Fun) *InterInst {
	return &InterInst{
		Op:  OP_ENTRY,
		Fun: fun,
	}
}

func NewExitInst(fun *Fun) *InterInst {
	return &InterInst{
		Op:  OP_EXIT,
		Fun: fun,
	}
}

func NewRetInst(retp *InterInst) *InterInst {
	return &InterInst{
		Op:     OP_RET,
		Target: retp,
	}
}

func NewRetvInst(retv *Var, retp *InterInst) *InterInst {
	return &InterInst{
		Op:     OP_RETV,
		Arg1:   retv,
		Target: retp,
	}
}

func NewProcInst(f *Fun) *InterInst {
	return &InterInst{
		Op:  OP_PROC,
		Fun: f,
	}
}

func NewCallInst(f *Fun, res *Var) *InterInst {
	return &InterInst{
		Op:     OP_CALL,
		Fun:    f,
		Result: res,
	}
}

// 无条件跳转: OP_JMP
func NewJmpInst(label *InterInst) *InterInst {
	return &InterInst{
		Op:     OP_JMP,
		Target: label,
	}
}

// 条件跳转: OP_JT, OP_JF
func NewCondJmpInst(op Operator, label *InterInst, cond *Var) *InterInst {
	return &InterInst{
		Op:     op,
		Target: label,
		Arg1:   cond,
	}
}

func NewJNEInst(label *InterInst, cond *Var, v *Var) *InterInst {
	return &InterInst{
		Op:     OP_JNE,
		Target: label,
		Arg1:   cond,
		Arg2:   v,
	}
}

func NewDecInst(v *Var) *InterInst {
	return &InterInst{
		Op:   OP_DEC,
		Arg1: v,
	}
}

func (i *InterInst) SliceString() []string {
	var i1, i2, i3, i4, i5 string
	if i.Result != nil {
		i1 = i.Result.Name
	}
	if i.Arg1 != nil {
		i2 = i.Arg1.Name
	}
	if i.Arg2 != nil {
		i3 = i.Arg2.Name
	}
	if i.Target != nil {
		i4 = i.Target.Label
	}
	if i.Fun != nil {
		i5 = i.Fun.Name
	}
	return []string{
		OpType[i.Op],
		i1,
		i2,
		i3,
		i.Label,
		i4,
		i5,
	}
}

/*
全局变量（全局符号）在汇编中的引用方式为符号名。 例如 'sum = 0;' 翻译为 'mov [sum], 0'
局部变量（局部符号）在汇编中的引用方式为base+offset。例如 'mov [ebp+v.offset], 0'
*/
func (i *InterInst) ToX86Asm() {
	if i.Label != "" {
		Emit(fmt.Sprintf("%s:", i.Label))
		return
	}
	switch i.Op {
	case OP_DEC:
		InitVar(i.Arg1)
	case OP_ENTRY:
		Emit("push ebp")
		Emit("mov ebp, esp")
		Emit(fmt.Sprintf("sub esp, %d", i.Fun.MaxDepth))
	case OP_EXIT:
		Emit("mov esp, ebp")
		Emit("pop ebp")
		Emit("ret")
	case OP_AS:
		LoadVar("eax", "al", i.Arg1)
		StoreVar("eax", "al", i.Result)
	case OP_ADD:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("add eax, ebx")
		StoreVar("eax", "al", i.Result)
	case OP_SUB:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("sub eax, ebx")
		StoreVar("eax", "al", i.Result)
	case OP_MUL:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("imul ebx")
		StoreVar("eax", "al", i.Result)
	case OP_DIV:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("idiv ebx")
		StoreVar("eax", "al", i.Result)
	case OP_MOD:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("idiv ebx")
		StoreVar("edx", "dl", i.Result)
	case OP_NEG:
		LoadVar("eax", "al", i.Arg1)
		Emit("neg eax")
		StoreVar("eax", "al", i.Result)
	case OP_GT:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("mov ecx, 0")
		Emit("cmp eax, ebx")
		Emit("setg cl")
		StoreVar("ecx", "cl", i.Result)
	case OP_GE:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("mov ecx, 0")
		Emit("cmp eax, ebx")
		Emit("setge cl")
		StoreVar("ecx", "cl", i.Result)
	case OP_LT:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("mov ecx, 0")
		Emit("cmp eax, ebx")
		Emit("setl cl")
		StoreVar("ecx", "cl", i.Result)
	case OP_LE:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("mov ecx, 0")
		Emit("cmp eax, ebx")
		Emit("setle cl")
		StoreVar("ecx", "cl", i.Result)
	case OP_EQU:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("mov ecx, 0")
		Emit("cmp eax, ebx")
		Emit("sete cl")
		StoreVar("ecx", "cl", i.Result)
	case OP_NEQU:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("mov ecx, 0")
		Emit("cmp eax, ebx")
		Emit("setne cl")
		StoreVar("ecx", "cl", i.Result)
	case OP_NOT:
		LoadVar("eax", "al", i.Arg1)
		Emit("mov ebx, 0")
		Emit("cmp eax, 0")
		Emit("sete bl")
		StoreVar("ebx", "bl", i.Result)
	case OP_AND:
		LoadVar("eax", "al", i.Arg1)
		Emit("cmp, eax, 0")
		Emit("setne cl")
		LoadVar("ebx", "bl", i.Arg2)
		Emit("cmp ebx, 0")
		Emit("setne bl")
		Emit("and al, bl")
		StoreVar("eax", "al", i.Result)
	case OP_OR:
		LoadVar("eax", "al", i.Arg1)
		Emit("cmp eax, 0")
		Emit("setne al")
		LoadVar("ebx", "bl", i.Arg2)
		Emit("cmp ebx, 0")
		Emit("setne bl")
		Emit("or al, bl")
		StoreVar("eax", "al", i.Result)
	case OP_JMP:
		Emit(fmt.Sprintf("jmp %s", i.Target.Label))
	case OP_JT:
		LoadVar("eax", "al", i.Arg1)
		Emit("cmp eax, 0")
		Emit(fmt.Sprintf("jne %s", i.Target.Label))
	case OP_JF:
		LoadVar("eax", "al", i.Arg1)
		Emit("cmp eax, 0")
		Emit(fmt.Sprintf("je %s", i.Target.Label))
	case OP_JNE:
		LoadVar("eax", "al", i.Arg1)
		LoadVar("ebx", "bl", i.Arg2)
		Emit("cmp eax, ebx")
		Emit(fmt.Sprintf("jne %s", i.Target.Label))
	case OP_ARG:
		LoadVar("eax", "al", i.Arg1)
		Emit("push eax")
	case OP_PROC:
		Emit(fmt.Sprintf("call %s", i.Fun.Name))
		Emit(fmt.Sprintf("add esp, %d", len(i.Fun.ParaVar)*4))
	case OP_CALL:
		Emit(fmt.Sprintf("call %s", i.Fun.Name))
		Emit(fmt.Sprintf("add esp, %d", len(i.Fun.ParaVar)*4))
		StoreVar("eax", "al", i.Result)
	case OP_RET:
		Emit(fmt.Sprintf("jmp %s", i.Target.Label))
	case OP_RETV:
		LoadVar("eax", "al", i.Arg1)
		Emit(fmt.Sprintf("jmp %s", i.Target.Label))
	case OP_LEA:
		LeaVar("eax", i.Arg1)
		StoreVar("eax", "al", i.Result) //TODO:??
	case OP_SET:
		LoadVar("eax", "al", i.Result)
		LoadVar("ebx", "bl", i.Arg1)
		Emit("mov [ebx], eax")
	case OP_GET:
		LoadVar("eax", "al", i.Arg1)
		Emit("mov eax, [eax]")
		StoreVar("eax", "al", i.Result)
	}
}

const (
	OP_NOP Operator = iota
	OP_DEC
	OP_ENTRY
	OP_EXIT
	OP_AS
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_MOD
	OP_NEG
	OP_GT
	OP_GE
	OP_LT
	OP_LE
	OP_EQU
	OP_NEQU
	OP_NOT
	OP_AND
	OP_OR
	OP_LEA
	OP_SET
	OP_GET
	OP_JMP
	OP_JT
	OP_JF
	OP_JNE
	OP_ARG
	OP_PROC
	OP_CALL
	OP_RET
	OP_RETV
)

var OpType = []string{
	"OP_NOP",
	"OP_DEC",
	"OP_ENTRY",
	"OP_EXIT",
	"OP_AS",
	"OP_ADD",
	"OP_SUB",
	"OP_MUL",
	"OP_DIV",
	"OP_MOD",
	"OP_NEG",
	"OP_GT",
	"OP_GE",
	"OP_LT",
	"OP_LE",
	"OP_EQU",
	"OP_NEQU",
	"OP_NOT",
	"OP_AND",
	"OP_OR",
	"OP_LEA",
	"OP_SET",
	"OP_GET",
	"OP_JMP",
	"OP_JT",
	"OP_JF",
	"OP_JNE",
	"OP_ARG",
	"OP_PROC",
	"OP_CALL",
	"OP_RET",
	"OP_RETV",
}
