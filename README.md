# calgo
So far, the purpose of this project is to facilitate my own learning.
# Environment
- Machine:    x86-64
- OS:         ubuntu 20.04
- Go version: 1.18
# Install Step
1. `git clone`
2. `cd calgo`
3. `go build`
4. `./calgo`

# An example
The file **'intercode.demo'** in directory **'demo'** showcases what this language can achieve.
If you just run ./calgo, the default output diretory is '**out'**
and the default output files include 'code.asm' which is the output of compiler
and **'elf_reloc.o'** which is the output of assembler.
The file **'tmp_code_seg_out'** is the data segment of 'elf_reloc.o'.
You can view detailed information for 'elf_reloc.o' using 'readelf' and 'objdump'.

And you can view function's intercode generated by compiler by using command line argument
**'--print_intercode=func_name1,func_name2'**. The following picture illustrates an example.

![image](https://github.com/jujubos/imgrepo/blob/master/calgo_print_intercode.png)

# About language
## Type
> Basic Type
- void
- char
- int

> Derived Type
- pointer
- array

## Declaration and Definition
> **Global Scope**
```
int var;
int var = 3;
int var1, var2;
int *ptr, arr[7];
void func();
char* func(int x, char *s) { ... }
int func() { ... }

/* Not currently supported */
int arr[] = {1, 2, 3}; //array initilization is not supported
```

> **Local Scope**

Same as **"Global Scope"**

## Expression
- =
- ||, &&
- \>, <, >=, <=, ==, !=
- +, -, *, /, %
- !,-,&,*,
- ++,--: *prefix increment/decrement*
- ++,--: *postfix increment/decrement*
- (): *Bracket expression*
- []: *array index expression*
- (): *function call expression*

## Statement
- expression statement
- while statement
- if statement
- break statement
- continue statement
- return statement