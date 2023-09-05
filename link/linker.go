package link

import (
	"log"
	"unsafe"
)

type Linker struct {
	exe        *ELF
	elfs       []*ELF
	segnames   []string
	symdefs    []*SymLink
	symlinks   []*SymLink
	startowner *ELF
	seglists   map[string]*SegList
}

type Block struct {
	data   []byte
	offset uint32 //相对段基址的偏移
	size   uint32
}

type SegList struct {
	baseaddr  uint32   //基地址
	begin     uint32   //对齐前偏移
	offset    uint32   //对齐后偏移
	size      uint32   //总大小
	ownerlist []*ELF   //所有者文件
	blocks    []*Block //数据块
}

type SymLink struct {
	name string
	recv *ELF
	prov *ELF
}

func NewLinker() *Linker {
	l := &Linker{
		seglists: map[string]*SegList{},
		exe:      NewELF(),
	}
	l.segnames = append(l.segnames, ".data")
	l.segnames = append(l.segnames, ".text")
	return l
}

func (l *Linker) AllocAddr() {
	curAddr := uint32(BaseAddr)
	curoff := uint32(52 + int(unsafe.Sizeof(Elf32_Phdr{}))*len(l.segnames)) //offset
	for _, n := range l.segnames {
		l.seglists[n].AllocAddr(n, &curAddr, &curoff)
	}
}

func (l *Linker) AddELF(f string) {
	elf := NewELF()
	elf.ReadElf(f)
	l.elfs = append(l.elfs, elf)
}

// 分配虚拟地址, 目前虚拟地址已分配到base, 可执行文件已分配到off。根据name确定对齐大小
func (s *SegList) AllocAddr(name string, base, off *uint32) {
	s.begin = *off //对齐前偏移
	align := uint32(DiscAlign)
	if name == ".text" {
		align = 16
	}
	*off += (align - (*off)%align) % align //对齐后的段文件偏移
	*base += (MemAlign-(*base)%MemAlign)%MemAlign + (*off)%MemAlign

	s.baseaddr = *base
	s.offset = *off
	s.size = 0 //s.size最终是seglist的所有段的大小之和。在计算的过程中，s.size是当前段在seglist内的偏移量。

	for _, elf := range s.ownerlist {
		shdr := elf.ShdrTab[name]
		shalign := shdr.sh_addralign
		s.size += (shalign - s.size%shalign) % shalign

		sb := make([]byte, shdr.sh_size)
		elf.GetData(sb, shdr.sh_offset)
		s.blocks = append(s.blocks, &Block{sb, s.size, shdr.sh_size})
		shdr.sh_addr = *base + s.size
		s.size += shdr.sh_size
	}
	*base += s.size
	*off += s.size
}

func (l *Linker) ColletInfo() {
	for _, elf := range l.elfs {
		for _, seg := range l.segnames {
			if _, ok := elf.ShdrTab[seg]; ok {
				l.seglists[seg].ownerlist = append(l.seglists[seg].ownerlist, elf)
			}
		}
		for name, sym := range elf.SymTab {
			if (sym.ST_Info >> 4) == STB_GLOBAL {
				symlink := &SymLink{}
				symlink.name = name
				if sym.ST_Shndx == STN_UNDEF { //导入符号
					symlink.recv = elf
					l.symlinks = append(l.symlinks, symlink)
				} else {
					symlink.prov = elf
					l.symdefs = append(l.symdefs, symlink)
				}
			}
		}
	}
}

func (l *Linker) SymValid() bool {
	for i := 0; i < len(l.symdefs); i++ {
		if l.symdefs[i].name == "@start" {
			l.startowner = l.symdefs[i].prov
		}
		for j := i + 1; j < len(l.symdefs); j++ {
			if l.symdefs[i].name == l.symdefs[j].name {
				log.Fatal("SymValid:符号重定义")
			}
		}
	}
	if l.startowner == nil {
		log.Fatal("SymValid:找不到入口点")
	}
	for _, syml := range l.symdefs {
		for _, symd := range l.symlinks {
			if syml.name == symd.name {
				syml.prov = symd.prov
			}
		}
		if syml.prov == nil {
			log.Fatal("SymValid:符号为未定义")
		}
	}
	return true
}

/*
1. 段加载的基址已经确定，将基址加上符号相对段的偏移得到符号的虚拟地址
2. 对于每个符号引用，设置符号引用的prov。
*/
func (l *Linker) SymParse() {
	//1.
	for _, elf := range l.elfs {
		for _, sym := range elf.SymTab {
			segname := elf.ShdrNames[sym.ST_Shndx]
			sym.ST_Value += elf.ShdrTab[segname].sh_offset
		}
	}
	//2.
	for _, sl := range l.symlinks {
		name := sl.name
		sl.recv.SymTab[name].ST_Value = sl.prov.SymTab[name].ST_Value
	}
}

func (l *Linker) Relocate() {
	for _, elf := range l.elfs {
		reltab := elf.RelTab
		for _, rel := range reltab {
			//符号
			symname := rel.Name
			sym := elf.SymTab[symname]
			//段
			segname := rel.Segname
			shdr := elf.ShdrTab[segname]
			//位置
			addr := shdr.sh_addr + rel.Rel.r_offset //addr是符号的虚拟地址(绝对)
			typ := rel.Rel.r_info & 0xff

			l.seglists[segname].RelocAddr(addr, typ, sym.ST_Value)
		}
	}
}

func (s *SegList) RelocAddr(reladdr, typ, symaddr uint32) {
	reloff := reladdr - s.baseaddr
	block := (*Block)(nil)
	for _, b := range s.blocks {
		st := b.offset
		ed := b.size
		if reloff >= st && reloff <= ed {
			block = b
			break
		}
	}
	paddr := reloff - block.offset //从这个block的第paddr个字节开始的四个字节将被修改
	ab := *(*[4]byte)(unsafe.Pointer(&symaddr))

	if typ == R_386_32 {
		for i := uint32(0); i < 4; i++ {
			block.data[paddr+i] = ab[i]
		}
	} else if typ == R_386_PC32 {
		v := uint32(0)
		for i := uint32(0); i < 4; i++ {
			v = v<<8 + uint32(block.data[paddr+i])
		}
		symaddr = symaddr - reladdr + v
		ab = *(*[4]byte)(unsafe.Pointer(&symaddr))
		for i := uint32(0); i < 4; i++ {
			block.data[paddr+i] = ab[i]
		}
	}
}

func (l *Linker) Link() {
	l.ColletInfo()
	l.SymValid()
	l.AllocAddr()
	l.SymParse()
	l.Relocate()
	l.exe.WriteElf(l)
}

const BaseAddr = 0x08048000
const MemAlign = 4096
const DiscAlign = 4
const TextAlign = 16

const (
	STB_GLOBAL = 1
	STN_UNDEF  = 0
)

const (
	R_386_NONE = 0
	R_386_32   = 1
	R_386_PC32 = 2
	SHN_UNDEF  = 0
)

const Start = "@start"
