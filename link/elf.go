package link

import (
	"io"
	"log"
	"os"
	"unsafe"
)

type ELF struct {
	Ehdr      Elf32_Ehdr             //文件头
	PhdrTab   []*Elf32_Phdr          //程序头表
	ShdrTab   map[string]*Elf32_Shdr //段表
	SymTab    map[string]*Elf32_Sym  //符号表
	RelTab    []*RelItem             //重定位表
	ShdrNames []string               //段名顺序
	SymNames  []string               //符号名
	LocSym    []*Elf32_Sym           //全局符号
	GlbSym    []*Elf32_Sym           //局部符号
	StrTab    string                 //字符串表
	ShStrTab  string                 //段表字符串表
	//RelTextTab []*Elf32_Rel
	//RelDataTab []*Elf32_Rel
	file *os.File
}

func NewELF() *ELF {
	elf := &ELF{
		ShdrTab: map[string]*Elf32_Shdr{},
		SymTab:  map[string]*Elf32_Sym{},
	}
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

type Elf32_Rel struct {
	r_offset uint32 //relocatable object file: 重定位位置偏移量，相对于段基址。executable object file: 重定位位置的线性地址，与动态链接相关
	r_info   uint32 //低8:重定位类型, 高24:重定位符号索引
}

type RelItem struct {
	Segname string
	Rel     *Elf32_Rel
	Name    string //重定位符号名
}

func (e *ELF) AddSym(name string, s *Elf32_Sym) {
	sym := &Elf32_Sym{}
	e.SymTab[name] = sym
	if name != "" {
		sym.ST_Name = 0
		sym.ST_Value = s.ST_Value
		sym.ST_Size = s.ST_Size
		sym.ST_Info = s.ST_Info
		sym.ST_Other = s.ST_Other
		sym.ST_Shndx = s.ST_Shndx
	}
	e.SymNames = append(e.SymNames, name)
}

func (e *ELF) addShdr(name string,
	sh_type, sh_flags, sh_addr, sh_offset, sh_size, sh_link, sh_info, sh_addralign, sh_entsize uint32) {
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

/*
文件头、.data、.text、.shstrtab、.shdrtab、.symtab、.strtab、.rel.text、.rel.data
*/
func (e *ELF) ReadElf(filename string) {
	fil, err := os.Open(filename)
	defer fil.Close()
	if err != nil {
		panic("ReadElf err! ")
	}
	e.file = fil
	sb := make([]byte, 52)
	_, err = fil.Read(sb)
	if err != nil {
		panic("ReadElf read ehdr err! ")
	}
	//ehdr
	e.Ehdr = *(*Elf32_Ehdr)(*(*unsafe.Pointer)(unsafe.Pointer(&sb)))
	//phdr
	sb = make([]byte, unsafe.Sizeof(Elf32_Phdr{}))
	for i := uint16(0); i < e.Ehdr.E_Phnum; i++ {
		_, err = fil.Read(sb)
		if err != nil {
			panic("ReadElf read phnum err! ")
		}
		e.PhdrTab = append(e.PhdrTab, (*Elf32_Phdr)(*(*unsafe.Pointer)(unsafe.Pointer(&sb))))
	}
	//.shstrtab
	fil.Seek(int64(e.Ehdr.E_Shoff)+int64(e.Ehdr.E_Shstrndx)*int64(unsafe.Sizeof(Elf32_Shdr{})), 0)
	sb = make([]byte, unsafe.Sizeof(Elf32_Shdr{}))
	_, err = fil.Read(sb)
	if err != nil {
		panic("Readelf read .shstrhdr err! ")
	}
	shstrhdr := (*Elf32_Shdr)(*(*unsafe.Pointer)(unsafe.Pointer(&sb)))
	fil.Seek(int64(shstrhdr.sh_offset), 0)
	sb = make([]byte, shstrhdr.sh_size)
	_, err = fil.Read(sb)
	if err != nil {
		panic("Readelf read .shstrtab err! ")
	}
	e.ShStrTab = string(sb)
	//.shdrtab
	fil.Seek(int64(e.Ehdr.E_Shoff), 0)
	sb = make([]byte, e.Ehdr.E_Shentsize)
	for i := uint16(0); i < e.Ehdr.E_Shnum; i++ {
		_, err = fil.Read(sb)
		if err != nil {
			panic("Readelf read .shdrtab err! ")
		}
		sh := *(*Elf32_Shdr)(*(*unsafe.Pointer)(unsafe.Pointer(&sb)))
		name := e.GetSegName(int(sh.sh_name))
		e.ShdrTab[name] = &sh
		e.ShdrNames = append(e.ShdrNames, name)
	}
	//.strtab
	strshdr := e.ShdrTab[".strtab"]
	fil.Seek(int64(strshdr.sh_offset), 0)
	sb = make([]byte, strshdr.sh_size)
	_, err = fil.Read(sb)
	if err != nil {
		panic("Readelf read .strtab err! ")
	}
	e.StrTab = string(sb)
	//.symtab
	symshdr := e.ShdrTab[".symtab"]
	fil.Seek(int64(symshdr.sh_offset), 0)
	symnum := symshdr.sh_size / symshdr.sh_entsize
	sb = make([]byte, symshdr.sh_entsize)
	for i := uint32(0); i < symnum; i++ {
		_, err = fil.Read(sb)
		if err != nil {
			panic("Readelf read .symtab err! ")
		}
		sym := *(*Elf32_Sym)(*(*unsafe.Pointer)(unsafe.Pointer(&sb)))
		name := e.GetSymName(int(sym.ST_Name))
		e.SymTab[name] = &sym
		e.SymNames = append(e.SymNames, name)
	}
	//.rel.data .rel.text
	for i := 0; i < len(e.ShdrNames); i++ {
		name := e.ShdrNames[i]
		shdr := e.ShdrTab[name]
		if shdr.sh_type == SHT_REL {
			fil.Seek(int64(shdr.sh_offset), 0)
			relnum := shdr.sh_size / shdr.sh_entsize
			sb = make([]byte, shdr.sh_entsize)
			for j := uint32(0); j < relnum; j++ {
				_, err = fil.Read(sb)
				if err != nil {
					panic("Readelf read .rel err! ")
				}
				rel := (*Elf32_Rel)(*(*unsafe.Pointer)(unsafe.Pointer(&sb)))
				segname := e.ShdrNames[shdr.sh_info]
				symname := e.SymNames[rel.r_info>>8]
				relitem := &RelItem{
					Segname: segname,
					Rel:     &Elf32_Rel{r_offset: rel.r_offset, r_info: rel.r_info},
					Name:    symname,
				}
				e.RelTab = append(e.RelTab, relitem)
			}
		}
	}
}

func (e *ELF) GetSegName(idx int) string {
	i := idx
	for ; i < len(e.ShStrTab); i++ {
		if e.ShStrTab[i] == 0 {
			break
		}
	}
	return e.ShStrTab[idx:i]
}

func (e *ELF) GetSymName(idx int) string {
	i := idx
	for ; i < len(e.StrTab); i++ {
		if e.StrTab[i] == 0 {
			break
		}
	}
	return e.StrTab[idx:i]
}

func (e *ELF) GetData(sb []byte, off uint32) {
	e.file.Seek(int64(off), 0)
	_, err := e.file.Read(sb)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *ELF) AssemObj(linker *Linker) {
	allsegnames := []string{""}
	for _, n := range linker.segnames {
		allsegnames = append(allsegnames, n)
	}
	allsegnames = append(allsegnames, ".shstrtab", ".symtab", ".strtab")
	shidx := map[string]int{}
	shstridx := map[string]int{}
	for i, n := range allsegnames {
		shidx[n] = i
		shstridx[n] = len(e.ShStrTab)
		e.ShStrTab += n
		e.ShStrTab += string(0)
	}
	//.symtab
	e.AddSym("", nil)
	symdef := linker.symdefs
	for _, sl := range symdef {
		name := sl.name
		prov := sl.prov
		sym := prov.SymTab[name]
		segname := prov.ShdrNames[sym.ST_Shndx]
		sym.ST_Shndx = uint16(shidx[segname])
		e.AddSym(name, sym)
	}
	symidx := map[string]uint32{}
	stridx := map[string]uint32{}
	for i, n := range e.SymNames {
		symidx[n] = uint32(i)
		stridx[n] = uint32(len(e.StrTab))
		e.StrTab += n
		e.StrTab += string(0)
	}
	for name, sym := range e.SymTab {
		sym.ST_Name = stridx[name]
	}

	magic := [16]byte{
		0x7f, 0x45, 0x4c, 0x46,
		0x01, 0x01, 0x01, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
	copy(e.Ehdr.E_Ident[:], magic[:])
	e.Ehdr.E_Type = ET_EXEC
	e.Ehdr.E_Machine = EM_386
	e.Ehdr.E_Version = EV_CURRENT
	e.Ehdr.E_Entry = e.SymTab[Start].ST_Name
	e.Ehdr.E_Phoff = 0
	e.Ehdr.E_Shoff = 0
	e.Ehdr.E_Flags = 0
	e.Ehdr.E_Ehsize = uint16(unsafe.Sizeof(Elf32_Ehdr{}))
	e.Ehdr.E_Phentsize = uint16(unsafe.Sizeof(Elf32_Phdr{}))
	e.Ehdr.E_Phnum = uint16(len(linker.segnames))
	e.Ehdr.E_Shentsize = 40
	e.Ehdr.E_Shnum = uint16(len(allsegnames))
	e.Ehdr.E_Shstrndx = uint16(shidx[".shstrtab"])
	//ehdr
	curoff := uint32(unsafe.Sizeof(e.Ehdr))
	e.Ehdr.E_Phoff = curoff
	//phdr
	for _, n := range linker.segnames {
		flags := uint32(PF_W | PF_R)
		if n == ".text" {
			flags = PF_X | PF_R
		}
		e.AddPhdr(PT_LOAD, linker.seglists[n].offset, linker.seglists[n].baseaddr, linker.seglists[n].size,
			linker.seglists[n].size, flags, MemAlign)
	}
	curoff += uint32(e.Ehdr.E_Phentsize * e.Ehdr.E_Phnum)
	//空段表项
	e.addShdr("", 0, 0, 0, 0, 0, 0, 0, 0, 0)
	for _, n := range linker.segnames {
		shflags := uint32(SHF_ALLOC | SHF_WRITE)
		shalign := uint32(DiscAlign)
		if n == ".text" {
			shflags = SHF_ALLOC | SHF_EXECINSTR
			shalign = TextAlign
		}
		e.addShdr(n, SHT_PROGBITS, shflags, linker.seglists[n].baseaddr, linker.seglists[n].offset, linker.seglists[n].size,
			0, 0, shalign, 0)
		curoff = linker.seglists[n].offset + linker.seglists[n].size
	}
	curoff += (4 - curoff%4) % 4
	//.shstrtab
	e.addShdr(".shstrtab", SHT_STRTAB, 0, 0, curoff, uint32(len(e.ShStrTab)),
		SHN_UNDEF, 0, 1, 0)
	curoff += uint32(len(e.ShStrTab))
	curoff += (4 - curoff%4) % 4
	//.shdrtab
	e.Ehdr.E_Shoff = curoff
	curoff += uint32(e.Ehdr.E_Shnum * e.Ehdr.E_Shentsize)
	//.symtab
	e.addShdr(".symtab", SHT_SYMTAB, 0, 0, curoff, uint32(len(e.SymNames))*20,
		uint32(shidx[".strtab"]), 0, 1, 20)
	curoff += uint32(len(e.SymNames)) * 20
	curoff += (4 - curoff%4) % 4
	//.strtab
	e.addShdr(".strtab", SHT_STRTAB, 0, 0, curoff, uint32(len(e.StrTab)),
		SHN_UNDEF, 0, 1, 0)
	curoff += uint32(len(e.StrTab))
	curoff += (4 - curoff%4) % 4
	//更新段表名
	for _, n := range allsegnames {
		e.ShdrTab[n].sh_name = uint32(shstridx[n])
	}
}

func (e *ELF) AddPhdr(typ, off, vaddr, filesz, memsz, flags, align uint32) {
	ph := &Elf32_Phdr{
		P_Type:   typ,
		P_Offset: off,
		P_Vaddr:  vaddr,
		P_FileSZ: filesz,
		P_MemSZ:  memsz,
		P_Flags:  flags,
		P_Align:  align,
	}
	e.PhdrTab = append(e.PhdrTab, ph)
}

func (e *ELF) WriteElf(linker *Linker) {
	e.AssemObj(linker)
	padnum := uint32(0)
	pad := byte(0)
	//文件头
	FWrite(unsafe.Pointer(&e.Ehdr), uint32(unsafe.Sizeof(e.Ehdr)), EXEFILE)
	//程序头表
	for _, ph := range e.PhdrTab {
		FWrite(unsafe.Pointer(ph), uint32(e.Ehdr.E_Phentsize), EXEFILE)
	}
	//.text .data
	for _, n := range linker.segnames {
		segs := linker.seglists[n]
		padnum := segs.offset - segs.begin
		FWrite(unsafe.Pointer(&pad), padnum, EXEFILE)
		last := (*Block)(nil)
		for _, bk := range segs.blocks {
			if last != nil {
				lasted := last.offset - last.size
				padnum = bk.offset - lasted
				FWrite(unsafe.Pointer(&pad), padnum, EXEFILE)
			}
			p := *(*uintptr)(unsafe.Pointer(&bk.data))
			FWrite(unsafe.Pointer(p), bk.size, EXEFILE)
		}
	}
	//.shstrtab
	padnum = e.ShdrTab[".shstrtab"].sh_offset - e.ShdrTab[".text"].sh_offset - e.ShdrTab[".text"].sh_size
	FWrite(unsafe.Pointer(&pad), padnum, EXEFILE)
	EXEFILE.Write([]byte(e.ShStrTab))
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
	//.shdrtab
	FWrite(unsafe.Pointer(&pad), padnum, EXEFILE)
	for _, n := range e.ShdrNames {
		s := e.ShdrTab[n]
		FWrite(unsafe.Pointer(s), 40, EXEFILE)
	}
	//.symtab
	for _, n := range e.SymNames {
		sym := e.SymTab[n]
		FWrite(unsafe.Pointer(sym), 20, EXEFILE)
	}
	//.strtab
	EXEFILE.Write([]byte(e.StrTab))
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

type SliceH struct {
	arr unsafe.Pointer
	l   int
	cap int
}

const (
	SHT_PROGBITS uint32 = 1 /* program defined information */
	SHT_SYMTAB   uint32 = 2 /* symbol table section */
	SHT_STRTAB   uint32 = 3 /* string table section */
	SHT_REL      uint32 = 9 /* relocation section - no addends */
)

const (
	ET_NONE = 0 /* Unknown type. */
	ET_REL  = 1 /* Relocatable. */
	ET_EXEC = 2 /* Executable. */
)

const (
	PF_X        = 0x1        /* Executable. */
	PF_W        = 0x2        /* Writable. */
	PF_R        = 0x4        /* Readable. */
	PF_MASKOS   = 0x0ff00000 /* Operating system-specific. */
	PF_MASKPROC = 0xf0000000 /* Processor-specific. */
)

const (
	PT_NULL   = 0 /* Unused entry. */
	PT_LOAD   = 1 /* Loadable segment. */
	PT_DYNAMI = 2 /* Dynamic linking information segment. */
	PT_INTERP = 3 /* Pathname of interpreter. */
	PT_NOTE   = 4 /* Auxiliary information. */
	PT_SHLIB  = 5 /* Reserved (not used). */
	PT_PHDR   = 6 /* Location of program header itself. */
	PT_TLS    = 7 /* Thread local storage segment */
)

const (
	SHF_WRITE            = 0x1        /* Section contains writable data. */
	SHF_ALLOC            = 0x2        /* Section occupies memory. */
	SHF_EXECINSTR        = 0x4        /* Section contains instructions. */
	SHF_MERGE            = 0x10       /* Section may be merged. */
	SHF_STRINGS          = 0x20       /* Section contains strings. */
	SHF_INFO_LINK        = 0x40       /* sh_info holds section index. */
	SHF_LINK_ORDER       = 0x80       /* Special ordering requirements. */
	SHF_OS_NONCONFORMING = 0x100      /* OS-specific processing required. */
	SHF_GROUP            = 0x200      /* Member of section group. */
	SHF_TLS              = 0x400      /* Section contains TLS data. */
	SHF_COMPRESSED       = 0x800      /* Section is compressed. */
	SHF_MASKOS           = 0x0ff00000 /* OS-specific semantics. */
	SHF_MASKPROC         = 0xf0000000 /* Processor-specific semantics. */
)

const EM_386 = 3
const EV_CURRENT = 1

var ElfExe = NewELF()
var EXEFILE *os.File

func init() {
	var err error
	EXEFILE, err = os.OpenFile("./out/exe.out", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
}
