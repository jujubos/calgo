package table

type Operator int

type InterInst struct {
	Label  string
	Op     Operator
	Result *Var
	Arg1   *Var
	Arg2   *Var
	Fun    *Fun
	Target *InterInst
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
		Op:     OP_RET,
		Arg1:   retv,
		Target: retp,
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
