package asm

import "unsafe"

type Lb_Record struct {
	SegName  string
	Name     string
	IsEqu    bool
	Externed bool
	Global   bool
	Addr     int //如果是宏，表示宏的值
	Times    int
	Len      int
	Cont     []int
}

// 对于可重定位文件是0，链接后开始于0x08048100
var CurAddr = 0x00000000
var CurSeg string

// 标签或外部符号
func NewLabel(name string, ex bool) *Lb_Record {
	lb := &Lb_Record{
		Name:     name,
		Externed: ex,
		Addr:     CurAddr,
		SegName:  CurSeg,
	}
	if ex {
		lb.Addr = 0x0
		lb.SegName = ""
	}
	return lb
}

// 宏
func NewEquLb(name string, v int) *Lb_Record {
	return &Lb_Record{
		Name:    name,
		Addr:    v,
		SegName: CurSeg,
		IsEqu:   true,
	}
}

// 数据
func NewDataLb(name string, t int, l int, v []int) *Lb_Record {
	lb := &Lb_Record{
		Name:    name,
		Times:   t,
		Len:     l,
		Cont:    v,
		SegName: CurSeg,
		Addr:    CurAddr,
	}
	CurAddr += t * l * len(v)
	return lb
}

func (l *Lb_Record) Write() {
	for i := 0; i < l.Times; i++ {
		for j := 0; j < len(l.Cont); j++ {
			FWrite(unsafe.Pointer(&l.Cont[i]), uint32(l.Len), Elfout)
			//WriteBytes(l.Cont[i], l.Len)
		}
	}
}
