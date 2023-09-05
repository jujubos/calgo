package asm

import (
	"io"
	"log"
	"os"
	"unsafe"
)

type ELF struct {
	Ehdr       Elf32_Ehdr             //文件头
	PhdrTab    []*Elf32_Phdr          //程序头表
	ShdrTab    map[string]*Elf32_Shdr //段表
	SymTab     map[string]*Elf32_Sym  //符号表
	RelTab     []*RelItem             //重定位表
	ShdrNames  []string               //段名顺序
	SymNames   []string               //符号名
	LocSym     []*Elf32_Sym           //全局符号
	GlbSym     []*Elf32_Sym           //局部符号
	StrTab     string                 //字符串表
	ShStrTab   string                 //段表字符串表
	RelTextTab []*Elf32_Rel
	RelDataTab []*Elf32_Rel
}

func NewELF() *ELF {
	elf := &ELF{
		ShdrTab: map[string]*Elf32_Shdr{},
		SymTab:  map[string]*Elf32_Sym{},
	}
	//添加空段表项和空符号表项
	elf.addShdr("", 0, 0, 0, 0, 0, 0, 0, 0, 0)
	elf.AddSym(NewLabel("", false))

	return elf
}

type Elf32_Ehdr struct {
	E_Ident     [16]byte
	E_Type      uint16
	E_Machine   uint16
	E_Version   uint32
	E_Entry     uint32
	E_Phoff     uint32
	E_Shoff     uint32
	E_Flags     uint32
	E_Ehsize    uint16
	E_Phentsize uint16
	E_Phnum     uint16
	E_Shentsize uint16
	E_Shnum     uint16
	E_Shstrndx  uint16
}

type Elf32_Phdr struct {
	P_Type   uint32
	P_Offset uint32
	P_Vaddr  uint32
	P_Paddr  uint32
	P_FileSZ uint32
	P_MemSZ  uint32
	P_Flags  uint32
	P_Align  uint32
}

type Elf32_Shdr struct {
	sh_name      uint32 //段名偏移量：相对于字符串表
	sh_type      uint32 //1:SHT_PROGBITS 2:SHT_SYMTAB 3:SHT_STRTAB 8:SHT_NOBITS 9:SHT_REL
	sh_flags     uint32 //1:SHF_WRITE 2:SHF_ALLOC 4:SHF_EXECINSTR
	sh_addr      uint32 //relocatable object file: 0, executable object file: 线性地址
	sh_offset    uint32 //段偏移量: 相对于ELF文件开始
	sh_size      uint32 //段的大小
	sh_link      uint32 //段的链接信息。对于符号表，link是串表的索引；info是第一个全局符号的的索引
	sh_info      uint32 //段的链接信息。对于重定位表，link是符号表的索引；info是代码段或数据段的索引
	sh_addralign uint32 //sh_offset%sh_addralign == 0
	sh_entsize   uint32 //表类型的段的一行的大小
}

type Elf32_Sym struct {
	ST_Name  uint32
	ST_Value uint32 //relocatable object file: 符号偏移量：相对段基址；executable object file: 符号线性地址
	ST_Size  uint32 //符号大小: 以字节为单位
	ST_Info  uint8  /* 低四位为符号类型。0:STT_NOTYPE 1:STT_OBJECT 2:STT_FUNC 3:STT:SECTION 4:STT_FILE。 高四位为符号绑定信息。0:STB_LOCAL 1:STB_GLOBAL 2:STB_WEAK */
	ST_Other uint8  //无用
	ST_Shndx uint16 //符号所在段。0:SHN_UNDEF 0xfff1:SHN_ABS 0xfff2:SHN_COMMON
}

type RelItem struct {
	Segname string
	Rel     *Elf32_Rel
	Name    string //重定位符号名
}

type Elf32_Rel struct {
	r_offset uint32 //relocatable object file: 重定位位置偏移量，相对于段基址。executable object file: 重定位位置的线性地址，与动态链接相关
	r_info   uint32 //低8:重定位类型, 高24:重定位符号索引
}

type ModRM struct {
	Mod int
	Reg int
	RM  int
}

type SIB struct {
	Scale int
	Index int
	Base  int
}

type Inst struct {
	Opcode  int
	Disp    int
	Imm32   int
	Displen int
}

func (i *Inst) SetDisp(d, dl int) {
	i.Disp = d
	i.Displen = dl
}

var ELFOBJ = NewELF()

func ProcessRel(typ int) bool {
	if ScanNum == 1 {
		return false
	}
	if RelLb == nil {
		return false
	}
	flg := false
	if typ == R_386_32 {
		ELFOBJ.AddRel(CurSeg, CurAddr, RelLb.Name, typ)
		flg = true
	} else if typ == R_386_PC32 {
		if RelLb.Externed {
			ELFOBJ.AddRel(CurSeg, CurAddr, RelLb.Name, typ)
			flg = true
		}
	}
	RelLb = nil
	return flg
}

func (e *ELF) AddRel(seg string, addr int, lb string, typ int) *RelItem {
	relitem := &RelItem{
		Segname: seg,
		Rel:     &Elf32_Rel{r_offset: uint32(addr), r_info: uint32(typ) & 0xff},
		Name:    lb,
	}
	e.RelTab = append(e.RelTab, relitem)
	return relitem
}

func (e *ELF) AddShdr(name string, sz int) {
	off := 52 + datalen
	if name == ".text" {
		e.addShdr(name, SHT_PROGBITS, SHF_EXECINSTR|SHF_ALLOC, 0, off, sz, 0, 0, 4, 0)
	} else if name == ".data" {
		e.addShdr(name, SHT_PROGBITS, SHF_ALLOC|SHF_WRITE, 0, off, sz, 0, 0, 4, 0)
	}
}

func (e *ELF) addShdr(name string,
	sh_type, sh_flags, sh_addr, sh_offset, sh_size, sh_link, sh_info, sh_addralign, sh_entsize int) {
	sh := &Elf32_Shdr{
		sh_name:      0,
		sh_type:      uint32(sh_type),
		sh_flags:     uint32(sh_flags),
		sh_addr:      uint32(sh_addr),
		sh_offset:    uint32(sh_offset),
		sh_size:      uint32(sh_size),
		sh_link:      uint32(sh_link),
		sh_info:      uint32(sh_info),
		sh_addralign: uint32(sh_addralign),
		sh_entsize:   uint32(sh_entsize),
	}
	e.ShdrTab[name] = sh
	e.ShdrNames = append(e.ShdrNames, name)
}

func (e *ELF) AddSym(lb *Lb_Record) {
	if lb.Name == "" { //空符号表项
		e.SymTab[lb.Name] = &Elf32_Sym{}
		e.SymNames = append(e.SymNames, lb.Name)
		return
	}
	s := &Elf32_Sym{
		ST_Name:  0,
		ST_Value: uint32(lb.Addr),
		ST_Size:  uint32(lb.Times * lb.Len * len(lb.Cont)),
	}
	if lb.Global {
		s.ST_Info = uint8(1 << 4)
	}
	if lb.Externed {
		s.ST_Shndx = SHN_UNDEF
	} else {
		if y := e.GetSegIndex(lb.SegName); y == -1 {
			panic("AddSym panic: y == -1")
		} else {
			s.ST_Shndx = uint16(y)
		}
	}
	e.SymTab[lb.Name] = s
	e.SymNames = append(e.SymNames, lb.Name)
}

func (e *ELF) GetSegIndex(name string) int {
	for i, s := range e.ShdrNames {
		if s == name {
			return i
		}
	}
	return -1
}

func (e *ELF) AssemObj() {
	allsegnames := e.ShdrNames
	allsegnames = append(allsegnames, ".shstrtab", ".symtab", ".strtab", ".rel.text", ".rel.data")
	shidx := map[string]int{}    //段名 -> 段表索引
	shstridx := map[string]int{} //段名 -> 串表索引
	for i, n := range allsegnames {
		shidx[n] = i
		shstridx[n] = len(e.ShStrTab)
		e.ShStrTab += n
		e.ShStrTab += string(byte(0))
	}
	symidx := map[string]int{}     //
	stridx := map[string]int{}     //所有的符号的串表索引
	allsymnames := []string{}      //除空符号外的所有符号名都在这了
	for _, n := range e.SymNames { //先局部符号
		if n == "" {
			continue
		}
		sym := e.SymTab[n]
		if (sym.ST_Info >> 4) == 0 {
			allsymnames = append(allsymnames, n)
			e.LocSym = append(e.LocSym, sym)
		}
	}
	for _, n := range e.SymNames { //后全局符号
		if n == "" {
			continue
		}
		sym := e.SymTab[n]
		if (sym.ST_Info >> 4) == 1 {
			allsymnames = append(allsymnames, n)
			e.GlbSym = append(e.GlbSym, sym)
		}
	}
	//所有的符号都准备就绪，可以生成符号字符串表了
	e.StrTab += string(0) //第一个永远是null
	for i, n := range allsymnames {
		symidx[n] = i + 1
		stridx[n] = len(e.StrTab)
		e.StrTab += n
		e.StrTab += string(0)
	}
	//更新符号名索引
	for _, n := range allsymnames {
		e.SymTab[n].ST_Name = uint32(stridx[n])
	}
	//处理重定位表
	for _, r := range e.RelTab {
		rel := &Elf32_Rel{}
		rel.r_offset = r.Rel.r_offset
		rel.r_info = uint32(symidx[r.Name])<<8 + r.Rel.r_info
		if r.Segname == ".text" {
			e.RelTextTab = append(e.RelTextTab, rel)
		} else if r.Segname == ".data" {
			e.RelTextTab = append(e.RelTextTab, rel)
		}
	}
	magic := [16]byte{
		0x7f, 0x45, 0x4c, 0x46,
		0x01, 0x01, 0x01, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
	copy(e.Ehdr.E_Ident[:], magic[:])
	e.Ehdr.E_Type = ET_REL
	e.Ehdr.E_Machine = EM_386
	e.Ehdr.E_Version = EV_CURRENT
	e.Ehdr.E_Entry = 0
	e.Ehdr.E_Phoff = 0
	e.Ehdr.E_Shoff = 0
	e.Ehdr.E_Flags = 0
	e.Ehdr.E_Ehsize = uint16(unsafe.Sizeof(Elf32_Ehdr{}))
	e.Ehdr.E_Phentsize = 0
	e.Ehdr.E_Phnum = 0
	e.Ehdr.E_Shentsize = 40
	e.Ehdr.E_Shnum = uint16(len(allsegnames))
	e.Ehdr.E_Shstrndx = uint16(shidx[".shstrtab"])

	curoff := int(unsafe.Sizeof(e.Ehdr)) //curoff = 52
	curoff += datalen                    //curoff = 52 + size of ".data" and ".text"
	//.shstrtab
	e.addShdr(".shstrtab", SHT_STRTAB,
		0, 0, curoff, len(e.ShStrTab), SHN_UNDEF, 0, 1, 0)
	curoff += len(e.ShStrTab)
	curoff += (4 - curoff%4) % 4
	//.shdrtab
	e.Ehdr.E_Shoff = uint32(curoff)
	curoff += int(e.Ehdr.E_Shnum * e.Ehdr.E_Shentsize)
	//.symtab
	//符号表的描述项中, info字段需要设置下。设置为第一个global符号的索引
	e.addShdr(".symtab", SHT_SYMTAB, 0, 0, curoff, int(len(e.SymNames))*int(unsafe.Sizeof(Elf32_Sym{})), shidx[".strtab"], len(e.LocSym)+1, 1, int(unsafe.Sizeof(Elf32_Sym{})))
	curoff += len(e.SymNames) * int(unsafe.Sizeof(Elf32_Sym{})) // 已对齐
	//.strtab
	e.addShdr(".strtab", SHT_STRTAB, 0, 0, curoff, len(e.StrTab), SHN_UNDEF, 0, 1, 0)
	curoff += len(e.StrTab)
	curoff += (4 - curoff%4) % 4
	//.rel.text
	e.addShdr(".rel.text", SHT_REL, 0, 0, curoff, len(e.RelTextTab)*int(unsafe.Sizeof(Elf32_Rel{})), shidx[".symtab"], shidx[".text"], 1, int(unsafe.Sizeof(Elf32_Rel{})))
	curoff += len(e.RelTextTab) * int(unsafe.Sizeof(Elf32_Rel{}))
	//.rel.data
	e.addShdr(".rel.data", SHT_REL, 0, 0, curoff, len(e.RelDataTab)*int(unsafe.Sizeof(Elf32_Rel{})), shidx[".symtab"], shidx[".data"], 1, int(unsafe.Sizeof(Elf32_Rel{})))
	curoff += len(e.RelDataTab) * int(unsafe.Sizeof(Elf32_Rel{}))

	for _, n := range e.ShdrNames {
		e.ShdrTab[n].sh_name = uint32(shstridx[n])
	}
}

func (e *ELF) WriteElf() {
	e.AssemObj()
	padnum := uint32(0)
	pad := byte(0)
	//文件头
	FWrite(unsafe.Pointer(&e.Ehdr), uint32(unsafe.Sizeof(e.Ehdr)), EXEFILE)
	//数据段
	Symtab.Write()
	//padding
	padnum = e.ShdrTab[".text"].sh_offset - e.ShdrTab[".data"].sh_offset - e.ShdrTab[".data"].sh_size
	FWrite(unsafe.Pointer(&pad), padnum, EXEFILE)
	//代码段
	codeseg, err := os.OpenFile("./out/tmp_code_seg.out", os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal(err)
		panic("WriteElf err, open tmp_code_seg.out failed!")
	}
	txt, err := io.ReadAll(codeseg)
	if err != nil {
		log.Fatal(err)
		panic("WriteElf err, io.ReadAll failed!")
	}
	_, err = EXEFILE.Write(txt)
	if err != nil {
		log.Fatal(err)
		panic("WriteElf err, EXEFILE.Write failed!")
	}
	//padding
	padnum = e.ShdrTab[".shstrtab"].sh_offset - e.ShdrTab[".text"].sh_offset - e.ShdrTab[".text"].sh_size
	FWrite(unsafe.Pointer(&pad), padnum, EXEFILE)
	//.shstrtab
	EXEFILE.Write([]byte(e.ShStrTab))
	//padding
	padnum = e.Ehdr.E_Shoff - e.ShdrTab[".shstrtab"].sh_offset - e.ShdrTab[".shstrtab"].sh_size
	FWrite(unsafe.Pointer(&pad), padnum, EXEFILE)
	//.shdrtab
	for _, name := range e.ShdrNames {
		sh := e.ShdrTab[name]
		FWrite(unsafe.Pointer(sh), uint32(unsafe.Sizeof(Elf32_Shdr{})), EXEFILE)
	}
	//.symtab
	nullsym := e.SymTab[""]
	FWrite(unsafe.Pointer(nullsym), uint32(unsafe.Sizeof(Elf32_Sym{})), EXEFILE)
	for _, sym := range e.LocSym {
		FWrite(unsafe.Pointer(sym), uint32(unsafe.Sizeof(Elf32_Sym{})), EXEFILE)
	}
	for _, sym := range e.GlbSym {
		FWrite(unsafe.Pointer(sym), uint32(unsafe.Sizeof(Elf32_Sym{})), EXEFILE)
	}
	//.strtab
	EXEFILE.Write([]byte(e.StrTab))
	//padding
	padnum = e.ShdrTab[".rel.text"].sh_offset - e.ShdrTab[".strtab"].sh_offset - e.ShdrTab[".strtab"].sh_size
	FWrite(unsafe.Pointer(&pad), padnum, EXEFILE)
	//.rel.text
	for _, r := range e.RelTextTab {
		FWrite(unsafe.Pointer(r), uint32(unsafe.Sizeof(Elf32_Rel{})), EXEFILE)
	}
	//.rel.data
	for _, r := range e.RelDataTab {
		FWrite(unsafe.Pointer(r), uint32(unsafe.Sizeof(Elf32_Rel{})), EXEFILE)
	}
}

type SliceH struct {
	arr unsafe.Pointer
	l   int
	cap int
}

func FWrite(p unsafe.Pointer, l uint32, f *os.File) {
	s := &SliceH{
		arr: p,
		l:   int(l),
		cap: int(l),
	}
	sp := (*[]byte)(unsafe.Pointer(s))
	_, err := f.Write(*sp)
	if err != nil {
		log.Fatal("Fwrite err! ", err)
	}
	f.Sync()
}

const (
	SHT_PROGBITS int = 1 /* program defined information */
	SHT_SYMTAB   int = 2 /* symbol table section */
	SHT_STRTAB   int = 3 /* string table section */
	SHT_REL      int = 9 /* relocation section - no addends */
)

const (
	SHF_WRITE            int = 0x1        /* Section contains writable data. */
	SHF_ALLOC            int = 0x2        /* Section occupies memory. */
	SHF_EXECINSTR        int = 0x4        /* Section contains instructions. */
	SHF_MERGE            int = 0x10       /* Section may be merged. */
	SHF_STRINGS          int = 0x20       /* Section contains strings. */
	SHF_INFO_LINK        int = 0x40       /* sh_info holds section index. */
	SHF_LINK_ORDER       int = 0x80       /* Special ordering requirements. */
	SHF_OS_NONCONFORMING int = 0x100      /* OS-specific processing required. */
	SHF_GROUP            int = 0x200      /* Member of section group. */
	SHF_TLS              int = 0x400      /* Section contains TLS data. */
	SHF_COMPRESSED       int = 0x800      /* Section is compressed. */
	SHF_MASKOS           int = 0x0ff00000 /* OS-specific semantics. */
	SHF_MASKPROC         int = 0xf0000000 /* Processor-specific semantics. */
)

const (
	R_386_NONE = 0
	R_386_32   = 1
	R_386_PC32 = 2
	SHN_UNDEF  = 0
)

const (
	ET_NONE = 0 /* Unknown type. */
	ET_REL  = 1 /* Relocatable. */
	ET_EXEC = 2 /* Executable. */
)

const EM_386 = 3
const EV_CURRENT = 1

var MODRM *ModRM = &ModRM{}
var SIBP *SIB = &SIB{}
var Instr *Inst = &Inst{}

func InstrInit() {
	MODRM.Mod = -1
	SIBP.Scale = -1
}

var EXEFILE *os.File
