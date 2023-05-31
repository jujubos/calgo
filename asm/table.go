package asm

type SymTable struct {
	Lb_Map map[string]*Lb_Record
	DefLbs []*Lb_Record
}

func NewSymTable() *SymTable {
	return &SymTable{
		Lb_Map: make(map[string]*Lb_Record),
	}
}

func (s *SymTable) AddLb(nlb *Lb_Record) { //nlb:new label
	if ScanNum > 1 {
		return
	}
	if olb, ok := s.Lb_Map[nlb.Name]; ok && olb.Global { //olb:old label
		nlb.Global = true
	}
	s.Lb_Map[nlb.Name] = nlb
	if nlb.Times != 0 && nlb.SegName == ".data" {
		s.DefLbs = append(s.DefLbs, nlb)
	}
}

func (s *SymTable) GetLb(name string) *Lb_Record {
	if l, ok := s.Lb_Map[name]; ok {
		return l
	}
	l := NewLabel(name, true)
	s.Lb_Map[name] = l
	return l
}

func (s *SymTable) ExportSyms() {
	for _, lb := range s.Lb_Map {
		if !lb.IsEqu {
			ELFOBJ.AddSym(lb)
		}
	}
}

func (s *SymTable) Write() {
	for _, l := range s.DefLbs {
		l.Write()
	}
}

var datalen = 0 //上一个段的结束位置

func SwitchSeg(name string) {
	if ScanNum == 1 {
		datalen += (4 - datalen%4) % 4
		ELFOBJ.AddShdr(CurSeg, CurAddr)
		datalen += CurAddr
	}
	CurSeg = name
	CurAddr = 0
}

var ScanNum = 0
var Symtab *SymTable = NewSymTable()
var RelLb *Lb_Record
