package table

import (
	"fmt"
	"os"
)

/*
变量的类型包括：基本类型(int, char), 指针类型(指针, 数组)
兼容规则：
- 基本类型之间兼容
- 如果是指针类型，基类型相同则兼容，否则不兼容
- 其他都不兼容
*/
func TypeCheck(p1, p2 *Var) bool {
	if p1.IsBase() && p2.IsBase() {
		return true
	}
	if !p1.IsBase() && !p2.IsBase() {
		return p1.Typ == p2.Typ
	}
	return false
}

func Warning(info string) {
	fmt.Printf("警告:%s\n", info)
	os.Exit(0)
}

func Error(info string) {
	fmt.Println(fmt.Sprintf("%s", info))
	os.Exit(0)
}
