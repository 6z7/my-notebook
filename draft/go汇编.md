## go汇编

go使用的是类似Plan9风格的汇编，是一种半抽象的汇编，需要经过编译器翻译成不同平台上的指令


```go
$ cat x.go
package main

func main() {
	println(3)
}
```
对应的go汇编代码
```go
$ GOOS=linux GOARCH=amd64 go tool compile -S x.go        # or: go build -gcflags -S x.go
"".main STEXT size=74 args=0x0 locals=0x10
	0x0000 00000 (x.go:3)	TEXT	"".main(SB), $16-0
	0x0000 00000 (x.go:3)	MOVQ	(TLS), CX
	0x0009 00009 (x.go:3)	CMPQ	SP, 16(CX)
	0x000d 00013 (x.go:3)	JLS	67
	0x000f 00015 (x.go:3)	SUBQ	$16, SP
	0x0013 00019 (x.go:3)	MOVQ	BP, 8(SP)
	0x0018 00024 (x.go:3)	LEAQ	8(SP), BP
	0x001d 00029 (x.go:3)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x001d 00029 (x.go:3)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x001d 00029 (x.go:3)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x001d 00029 (x.go:4)	PCDATA	$0, $0
	0x001d 00029 (x.go:4)	PCDATA	$1, $0
	0x001d 00029 (x.go:4)	CALL	runtime.printlock(SB)
	0x0022 00034 (x.go:4)	MOVQ	$3, (SP)
	0x002a 00042 (x.go:4)	CALL	runtime.printint(SB)
	0x002f 00047 (x.go:4)	CALL	runtime.printnl(SB)
	0x0034 00052 (x.go:4)	CALL	runtime.printunlock(SB)
	0x0039 00057 (x.go:5)	MOVQ	8(SP), BP
	0x003e 00062 (x.go:5)	ADDQ	$16, SP
	0x0042 00066 (x.go:5)	RET
	0x0043 00067 (x.go:5)	NOP
	0x0043 00067 (x.go:3)	PCDATA	$1, $-1
	0x0043 00067 (x.go:3)	PCDATA	$0, $-1
	0x0043 00067 (x.go:3)	CALL	runtime.morestack_noctxt(SB)
	0x0048 00072 (x.go:3)	JMP	0
...
```

`FUNCDATA`和`PCDATA`指令包含gc时需要的信息，是由编译器生成的。

反汇编生成后的可执行文件得到的汇编代码:
```go
$ go build -o x.exe x.go
$ go tool objdump -s main.main x.exe
TEXT main.main(SB) /tmp/x.go
  x.go:3		0x10501c0		65488b0c2530000000	MOVQ GS:0x30, CX
  x.go:3		0x10501c9		483b6110		CMPQ 0x10(CX), SP
  x.go:3		0x10501cd		7634			JBE 0x1050203
  x.go:3		0x10501cf		4883ec10		SUBQ $0x10, SP
  x.go:3		0x10501d3		48896c2408		MOVQ BP, 0x8(SP)
  x.go:3		0x10501d8		488d6c2408		LEAQ 0x8(SP), BP
  x.go:4		0x10501dd		e86e45fdff		CALL runtime.printlock(SB)
  x.go:4		0x10501e2		48c7042403000000	MOVQ $0x3, 0(SP)
  x.go:4		0x10501ea		e8e14cfdff		CALL runtime.printint(SB)
  x.go:4		0x10501ef		e8ec47fdff		CALL runtime.printnl(SB)
  x.go:4		0x10501f4		e8d745fdff		CALL runtime.printunlock(SB)
  x.go:5		0x10501f9		488b6c2408		MOVQ 0x8(SP), BP
  x.go:5		0x10501fe		4883c410		ADDQ $0x10, SP
  x.go:5		0x1050202		c3			RET
  x.go:3		0x1050203		e83882ffff		CALL runtime.morestack_noctxt(SB)
  x.go:3		0x1050208		ebb6			JMP main.main(SB)
```

go汇编中定义了4个伪寄存器:

* FP(Frame pointer): arguments and locals
* PC(Program counter): jumps and branches
* SB(Static base pointer):global symbols
* SP(Stack pointer): top of stack

所有用户定义的符号都被定义为相对于FP和SB的偏移。

SB伪寄存器可以认为是内存的起始地址，因此`foo(SB)`是`foo`在内存中的地址。这种方式用于命名全局函数和数据。在名字后边天加`<>`,如`foo<>(SB)`表示该符号仅在当前文件中可用。在名字后添加偏移，如`foo+4(SB)`表示foo的位置在相对于起始位置4个字节后的位置。

FP伪寄存器是一个虚拟帧指针，指向函数参数的位置。编译器维护一个虚拟帧指针，并将堆栈上的参数表示为该伪寄存器的偏移量。`0(FP)`表示函数的第一个参数， 8(FP)第二个参数(64位机器)。但是，当以这种方式引用函数参数时，必须在开头放置一个名称，如`first_arg+0(FP)`和`second_arg+8(FP)`。汇编程序强制执行此约定。名称没有特殊要求，通俗易懂即可。FP是一个伪寄存器，不是指的硬件寄存器FP。

在32位系统上，64位值的低32位和高32位是通过在名称后添加一个_lo或_hi后缀来区分的，如arg_lo+0(FP)或arg_hi+4(FP)。

SP伪寄存器是一个虚拟栈的指针，用于局部变量和为函数调用准备的参数。它指向本地栈帧的顶部，因此引用应该在范围内使用负偏移量[-framesize，0)：x-8(SP)，y-4(SP)，依此类推。

需要在SP添加一个名称前缀来区分是伪寄存器SP还是硬件寄存器SP。x-8(SP)和-8(SP)是不同的内存地址：前一个是相对于虚拟栈的位寄存器指针，后者则是相对于硬件寄存器SP的。

SP和PC是物理寄存器的别名，在Go汇编中使用SP和PC，需要带一个符号，像上边的FP一样。如果要访问硬件寄存器需要使用真实的R开头的名字。在ARM架构下，硬件SP和PC可以使用R13和R15访问。

分支和直接跳转地址使用的是相对于PC的偏移或跳转到指定label:

```go
label:
	MOVW $0, R1
	JMP label
```

每个label仅仅在所定义的函数内部可见，因此同一个文件中的不同函数中使用相同的lable是允许的。直接跳转和调用指令可以使用name(SB)，但不能带偏移的方式，如name+4(SB)。

https://golang.org/doc/asm















