package table

import (
	"fmt"
)

func Emit(s string) {
	_, err := AsmFile.WriteString(s + "\n")
	if err != nil {
		fmt.Println(err)
		panic("Emit error!")
	}
}

func LoadVar(reg32, reg8 string, v *Var) {
	reg := ""
	if v.IsChar() {
		reg = reg8
	} else {
		reg = reg32
	}
	name := v.Name
	if !v.Literal {
		off := v.Offset
		if off == 0 { //全局变量
			if !v.IsArray { //非数组
				Emit(fmt.Sprintf("mov %s, [%s]", reg, name))
			} else {
				Emit(fmt.Sprintf("mov %s, %s", reg, name))
			}
		} else { //局部变量
			if !v.IsArray {
				Emit(fmt.Sprintf("mov %s, [ebp%+d]", reg, off))
			} else {
				Emit(fmt.Sprintf("lea %s, [ebp%+d]", reg, off))
			}
		}
	} else {
		if v.IsBase() { //数字，字符
			var val int64
			if v.IsChar() {
				val = int64(v.CharVal)
			} else {
				val = v.IntVal
			}
			Emit(fmt.Sprintf("mov %s, %d", reg, val))
		} else { //字符串
			Emit(fmt.Sprintf("mov %s, %s", reg, name))
		}
	}
}

func LeaVar(reg32 string, v *Var) {
	name := v.Name
	if v.Offset == 0 {
		Emit(fmt.Sprintf("mov %s, %s", reg32, name))
	} else {
		Emit(fmt.Sprintf("mov %s, [ebp%+d]", reg32, v.Offset))
	}
}

func StoreVar(reg32, reg8 string, v *Var) {
	var reg string
	if v.IsChar() {
		reg = reg8
	} else {
		reg = reg32
	}
	name := v.Name
	if v.Offset == 0 {
		Emit(fmt.Sprintf("mov [%s], %s", name, reg))
	} else {
		Emit(fmt.Sprintf("mov [ebp%+d], %s", v.Offset, reg))
	}
}

func InitVar(v *Var) {
	if v.inited {
		var val int64
		if v.IsChar() {
			val = int64(v.CharVal)
		} else {
			val = v.IntVal
		}
		if v.IsBase() { //int, char
			Emit(fmt.Sprintf("mov eax, %d", val))
		} else { //int*, char*, int arr[],
			Emit(fmt.Sprintf("mov eax, %s", v.PtrVal)) //TODO:整数指针不考虑了???
		}
		StoreVar("eax", "ax", v)
	}
}
