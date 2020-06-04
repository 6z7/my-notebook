# defer源码分析

defer是go中用于声明在函数结束时必然执行的回调函数。

defer的源码在runtime/panic.go中(通过dlv调试,下断点b runtime.deferprocStack就可以找到)

对于defer，编译时逃逸分析后决定是分配在堆上还是栈上。不管是分配在哪里，都会使用到`_defer`数据结构用于在运行时保存定义的defer函数: 

 ```go  
type _defer struct {
    //函数的参数总大小包括返回值
    siz     int32 // includes both arguments and results
    //是否已经执行
    started bool
    //defer分配在堆上还是栈上
    heap    bool
    // 存储调用 defer 函数的函数的 sp 寄存器值
    sp      uintptr // sp at time of defer
    //存储 call deferproc 的下一条汇编指令的指令地址
    pc      uintptr
    //指向需要执行的匿名函数
    fn      *funcval
    //panic信息
    _panic  *_panic // panic that is running defer
    //下一个需要执行的defer
    link    *_defer
}
 ```

下面分别看下defer在栈或堆上的分配的情况

## defer分配在栈上

由于defer在栈上分配，所以在编译时已经确定了defer在栈上的分布情况。下面这段代码就是编译时构造的defer结构的[方法](https://github.com/6z7/go/blob/03250054f8512d35b10f17d3c886dbc4b1ad43c6/src/cmd/compile/internal/gc/ssa.go#L3836)，可以看到在_defer后边还有一个参数数组:

```go
func deferstruct(stksize int64) *types.Type {
     ......
    argtype := types.NewArray(types.Types[TUINT8], stksize)
	argtype.Width = stksize
	argtype.Align = 1
	// These fields must match the ones in runtime/runtime2.go:_defer and
	// cmd/compile/internal/gc/ssa.go:(*state).call.
	fields := []*types.Field{
		makefield("siz", types.Types[TUINT32]),
		makefield("started", types.Types[TBOOL]),
		makefield("heap", types.Types[TBOOL]),
		makefield("sp", types.Types[TUINTPTR]),
		makefield("pc", types.Types[TUINTPTR]),
		// Note: the types here don't really matter. Defer structures
		// are always scanned explicitly during stack copying and GC,
		// so we make them uintptr type even though they are real pointers.
		makefield("fn", types.Types[TUINTPTR]),
		makefield("_panic", types.Types[TUINTPTR]),
		makefield("link", types.Types[TUINTPTR]),
		makefield("args", argtype),
	}
......
}
```

举一个分配在栈的例子，通过反汇编看下defer在栈上是如何分配的。

```go
func main() {
	a, b := 1, 2
	defer func(a, b int) {
		fmt.Println(a, b)
	}(a, b)
}
```
通过`go tool compile -N -S -l demo.go`获取对应的汇编
```
 "".main STEXT size=149 args=0x0 locals=0x68
	0x0000 00000 (demo.go:7)	TEXT	"".main(SB), ABIInternal, $104-0
	0x0000 00000 (demo.go:7)	MOVQ	(TLS), CX
	0x0009 00009 (demo.go:7)	CMPQ	SP, 16(CX)
	0x000d 00013 (demo.go:7)	JLS	139
	0x000f 00015 (demo.go:7)	SUBQ	$104, SP
	0x0013 00019 (demo.go:7)	MOVQ	BP, 96(SP)
	0x0018 00024 (demo.go:7)	LEAQ	96(SP), BP	 
	0x001d 00029 (demo.go:8)	MOVQ	$1, "".a+24(SP)
	0x0026 00038 (demo.go:8)	MOVQ	$2, "".b+16(SP)
	0x002f 00047 (demo.go:9)	MOVL	$16, ""..autotmp_3+32(SP)	 
	0x0037 00055 (demo.go:9)	LEAQ	"".main.func1·f(SB), AX	 
	0x003e 00062 (demo.go:9)	MOVQ	AX, ""..autotmp_3+56(SP)
	0x0043 00067 (demo.go:9)	MOVQ	"".a+24(SP), AX
	0x0048 00072 (demo.go:9)	MOVQ	AX, ""..autotmp_3+80(SP)
	0x004d 00077 (demo.go:9)	MOVQ	"".b+16(SP), AX
	0x0052 00082 (demo.go:9)	MOVQ	AX, ""..autotmp_3+88(SP)	 
	0x0057 00087 (demo.go:9)	LEAQ	""..autotmp_3+32(SP), AX	 
	0x005c 00092 (demo.go:9)	MOVQ	AX, (SP)
	0x0060 00096 (demo.go:9)	CALL	runtime.deferprocStack(SB)
	0x0065 00101 (demo.go:9)	TESTL	AX, AX
	0x0067 00103 (demo.go:9)	JNE	123
	0x0069 00105 (demo.go:9)	JMP	107
	0x006b 00107 (demo.go:12)	XCHGL	AX, AX
	0x006c 00108 (demo.go:12)	CALL	runtime.deferreturn(SB)
	0x0071 00113 (demo.go:12)	MOVQ	96(SP), BP
	0x0076 00118 (demo.go:12)	ADDQ	$104, SP
	0x007a 00122 (demo.go:12)	RET
	0x007b 00123 (demo.go:9)	XCHGL	AX, AX
	0x007c 00124 (demo.go:9)	CALL	runtime.deferreturn(SB)
	0x0081 00129 (demo.go:9)	MOVQ	96(SP), BP
	0x0086 00134 (demo.go:9)	ADDQ	$104, SP
	0x008a 00138 (demo.go:9)	RET
	0x008b 00139 (demo.go:9)	NOP	 
	0x008b 00139 (demo.go:7)	CALL	runtime.morestack_noctxt(SB)
	0x0090 00144 (demo.go:7)	JMP	0
	 
"".main.func1 STEXT size=94 args=0x10 locals=0x10
	0x0000 00000 (demo.go:9)	TEXT	"".main.func1(SB), ABIInternal, $16-16
	0x0000 00000 (demo.go:9)	MOVQ	(TLS), CX
	0x0009 00009 (demo.go:9)	CMPQ	SP, 16(CX)
	0x000d 00013 (demo.go:9)	JLS	87
	0x000f 00015 (demo.go:9)	SUBQ	$16, SP
	0x0013 00019 (demo.go:9)	MOVQ	BP, 8(SP)
	0x0018 00024 (demo.go:9)	LEAQ	8(SP), BP	 
	0x001d 00029 (demo.go:10)	CALL	runtime.printlock(SB)
	0x0022 00034 (demo.go:10)	MOVQ	"".a+24(SP), AX
	0x0027 00039 (demo.go:10)	MOVQ	AX, (SP)
	0x002b 00043 (demo.go:10)	CALL	runtime.printint(SB)
	0x0030 00048 (demo.go:10)	CALL	runtime.printsp(SB)
	0x0035 00053 (demo.go:10)	MOVQ	"".b+32(SP), AX
	0x003a 00058 (demo.go:10)	MOVQ	AX, (SP)
	0x003e 00062 (demo.go:10)	CALL	runtime.printint(SB)
	0x0043 00067 (demo.go:10)	CALL	runtime.printnl(SB)
	0x0048 00072 (demo.go:10)	CALL	runtime.printunlock(SB)
	0x004d 00077 (demo.go:11)	MOVQ	8(SP), BP
	0x0052 00082 (demo.go:11)	ADDQ	$16, SP
	0x0056 00086 (demo.go:11)	RET
	0x0057 00087 (demo.go:11)	NOP	 
	0x0057 00087 (demo.go:9)	CALL	runtime.morestack_noctxt(SB)
	0x005c 00092 (demo.go:9)	JMP	0 

```
  
通过上边的汇编，可以看出在栈上分配了一段空间来表示defer,大致画图main方法上栈的布局：

![](../image/defer2.png)


**runtime.deferprocStack**

用于将栈上新分配的defer保存到当前g._defer上，所有的defer构成一个FIFO队列。

![](../image/defer1.png)

```go    
// 将新的defer加入LIFO队列   
// 第一个参数指向栈上的defer地址 
//go:nosplit
func deferprocStack(d *_defer) {
    gp := getg()
    // g0上不能执行defer
    if gp.m.curg != gp {
        // go code on the system stack can't defer
        throw("defer on system stack")
    }
   
    d.started = false
    d.heap = false
    d.sp = getcallersp()
    d.pc = getcallerpc()

    //构建链表，当前g上保存defer
    *(*uintptr)(unsafe.Pointer(&d._panic)) = 0
    *(*uintptr)(unsafe.Pointer(&d.link)) = uintptr(unsafe.Pointer(gp._defer))
    *(*uintptr)(unsafe.Pointer(&gp._defer)) = uintptr(unsafe.Pointer(d))    
    //隐式返回0  编译器生产的代码会插入判断，当程序发生 panic 之后会返回非0
    return0()
    
} 
```
***return0的汇编实现***
```
TEXT runtime·return0(SB), NOSPLIT, $0
	MOVL	$0, AX
	RET
```

**runtime.deferreturn**

循环执行当前g上所有的defer。既然是循环调用，就要有退出条件，满足以下条件退出循环:

* d==nil 即g上的defer执行完了

* d.sp!=sp  非同一个SP下的直接返回，这种情况发生在函数嵌套调用时,如 f1()->f2()->f3()，每个函数中都有defer,每次函数退出时只执行当前函数的defer

根据参数的大小确定如何复制参数：

1. 如果参数或返回值大小为0，不做处理
2. 如果参数或返回值大小为一个指针大小直接复制
3. 反之则通过内存复制 

```go
// arg0:第一个参数地址，即调用者SP
func deferreturn(arg0 uintptr) {
    gp := getg()
    d := gp._defer
    if d == nil {
        return
    }
    sp := getcallersp()
    if d.sp != sp {
        return
    }

    switch d.siz {
    case 0:
        // Do nothing.
    case sys.PtrSize:
        *(*uintptr)(unsafe.Pointer(&arg0)) = *(*uintptr)(deferArgs(d))
    default:
        memmove(unsafe.Pointer(&arg0), deferArgs(d), uintptr(d.siz))
    }
    fn := d.fn
    d.fn = nil
    gp._defer = d.link
    freedefer(d)
    // 通过汇编实现调用fn并循环调用deferreturn直至结束
    // &arg0的地址就是defer的地址
    jmpdefer(fn, uintptr(unsafe.Pointer(&arg0)))
}

// 计算defer参数的位置
//go:nosplit
func deferArgs(d *_defer) unsafe.Pointer {
	if d.siz == 0 {
		// Avoid pointer past the defer allocation.
		return nil
	}
	return add(unsafe.Pointer(d), unsafe.Sizeof(*d))
}
```

***jmpdefer的汇编实现***

该汇编会回复SP、BP为调用者对应的值，并将调用者调用deferreturn时要入栈的下一条指令替换`call runtime.deferreturn`指令，这样就可以实现循环调用deferreturn了。
```
TEXT runtime·jmpdefer(SB), NOSPLIT, $0-16
    // defer的函数的地址
    MOVQ	fv+0(FP), DX	
    // 参数argp的地址，即调用者的SP地址
    MOVQ	argp+8(FP), BX
    // SP-8,即 调用者执行call runtime.deferreturn时压入栈的下一条指令的地址 
    LEAQ	-8(BX), SP	// caller sp after CALL
    // 恢复调用者的BP
    MOVQ	-8(SP), BP	
    // SP-5 指令恢复到call runtime.deferreturn位置
    SUBQ	$5, (SP)	// return to CALL again
    MOVQ	0(DX), BX
    // 执行函数fn
    // fn执行完成后，返回到deferreturn,再返回到调用者，由于调用者的返回地址被修改为call runtime.deferreturn的地址，则循环执行call runtime.deferreturn
    JMP	BX	 
```
runtime.deferreturn的上下文，CALL	runtime.deferreturn与下一条指令的地址相差5(00129-00124)
```
0x007c 00124 (demo.go:7)	CALL	runtime.deferreturn(SB)
0x0081 00129 (demo.go:7)	MOVQ	96(SP), BP
0x0086 00134 (demo.go:7)	ADDQ	$104, SP
0x008a 00138 (demo.go:7)	RET
```

调用者与deferreturn栈的关系:

![](../image/defer3.png)

## defer分配在堆上

先看一个分配在堆上的例子:
```go
func main() {
    for i := 0; i < 5; i++ {
        defer func(a,b int)  {
            println(a,b)
        }(1,2)
    }
}

```
对应的汇编：
```
"".main STEXT size=161 args=0x0 locals=0x30
	0x0000 00000 (demo.go:7)	TEXT	"".main(SB), ABIInternal, $48-0
	0x0000 00000 (demo.go:7)	MOVQ	(TLS), CX
	0x0009 00009 (demo.go:7)	CMPQ	SP, 16(CX)
	0x000d 00013 (demo.go:7)	JLS	151
	0x0013 00019 (demo.go:7)	SUBQ	$48, SP
	0x0017 00023 (demo.go:7)	MOVQ	BP, 40(SP)
	0x001c 00028 (demo.go:7)	LEAQ	40(SP), BP	
	0x0021 00033 (demo.go:8)	MOVQ	$0, "".i+32(SP)
	0x002a 00042 (demo.go:8)	JMP	44
	0x002c 00044 (demo.go:8)	CMPQ	"".i+32(SP), $5
	0x0032 00050 (demo.go:8)	JLT	54
	0x0034 00052 (demo.go:8)	JMP	135
    // 参数大小  16字节
	0x0036 00054 (demo.go:9)	MOVL	$16, (SP)	
	0x003d 00061 (demo.go:9)	LEAQ	"".main.func1·f(SB), AX
    // 函数地址	
	0x0044 00068 (demo.go:9)	MOVQ	AX, 8(SP)
    // 函数参数
	0x0049 00073 (demo.go:9)	MOVQ	$1, 16(SP)
    // 函数参数
	0x0052 00082 (demo.go:9)	MOVQ	$2, 24(SP)
	0x005b 00091 (demo.go:9)	CALL	runtime.deferproc(SB)
	0x0060 00096 (demo.go:11)	TESTL	AX, AX
	0x0062 00098 (demo.go:11)	JNE	119
	0x0064 00100 (demo.go:11)	JMP	102	 
	0x0066 00102 (demo.go:8)	JMP	104
	0x0068 00104 (demo.go:8)	PCDATA	$0, $0
	0x0068 00104 (demo.go:8)	PCDATA	$1, $0
	0x0068 00104 (demo.go:8)	MOVQ	"".i+32(SP), AX
	0x006d 00109 (demo.go:8)	INCQ	AX
	0x0070 00112 (demo.go:8)	MOVQ	AX, "".i+32(SP)
	0x0075 00117 (demo.go:8)	JMP	44
	0x0077 00119 (demo.go:9)	XCHGL	AX, AX
	0x0078 00120 (demo.go:9)	CALL	runtime.deferreturn(SB)
	0x007d 00125 (demo.go:11)	MOVQ	40(SP), BP
	0x0082 00130 (demo.go:11)	ADDQ	$48, SP
	0x0086 00134 (demo.go:11)	RET
	0x0087 00135 (demo.go:13)	XCHGL	AX, AX
	0x0088 00136 (demo.go:13)	CALL	runtime.deferreturn(SB)
	0x008d 00141 (demo.go:13)	MOVQ	40(SP), BP
	0x0092 00146 (demo.go:13)	ADDQ	$48, SP
	0x0096 00150 (demo.go:13)	RET
	0x0097 00151 (demo.go:13)	NOP
	0x0097 00151 (demo.go:7)	PCDATA	$1, $-1
	0x0097 00151 (demo.go:7)	PCDATA	$0, $-1
	0x0097 00151 (demo.go:7)	CALL	runtime.morestack_noctxt(SB)
	0x009c 00156 (demo.go:7)	JMP	0
	 
"".main.func1 STEXT size=94 args=0x10 locals=0x10
	0x0000 00000 (demo.go:9)	TEXT	"".main.func1(SB), ABIInternal, $16-16
	0x0000 00000 (demo.go:9)	MOVQ	(TLS), CX
	0x0009 00009 (demo.go:9)	CMPQ	SP, 16(CX)
	0x000d 00013 (demo.go:9)	JLS	87
	0x000f 00015 (demo.go:9)	SUBQ	$16, SP
	0x0013 00019 (demo.go:9)	MOVQ	BP, 8(SP)
	0x0018 00024 (demo.go:9)	LEAQ	8(SP), BP	 
	0x001d 00029 (demo.go:10)	CALL	runtime.printlock(SB)
	0x0022 00034 (demo.go:10)	MOVQ	"".a+24(SP), AX
	0x0027 00039 (demo.go:10)	MOVQ	AX, (SP)
	0x002b 00043 (demo.go:10)	CALL	runtime.printint(SB)
	0x0030 00048 (demo.go:10)	CALL	runtime.printsp(SB)
	0x0035 00053 (demo.go:10)	MOVQ	"".b+32(SP), AX
	0x003a 00058 (demo.go:10)	MOVQ	AX, (SP)
	0x003e 00062 (demo.go:10)	CALL	runtime.printint(SB)
	0x0043 00067 (demo.go:10)	CALL	runtime.printnl(SB)
	0x0048 00072 (demo.go:10)	CALL	runtime.printunlock(SB)
	0x004d 00077 (demo.go:11)	MOVQ	8(SP), BP
	0x0052 00082 (demo.go:11)	ADDQ	$16, SP
	0x0056 00086 (demo.go:11)	RET
	0x0057 00087 (demo.go:11)	NOP
	0x0057 00087 (demo.go:9)	PCDATA	$1, $-1
	0x0057 00087 (demo.go:9)	PCDATA	$0, $-1
	0x0057 00087 (demo.go:9)	CALL	runtime.morestack_noctxt(SB)
	0x005c 00092 (demo.go:9)	JMP	0
```

**runtime.deferproc**

1. 在堆上创建defer对象
2. 将栈上的函数参数复制到堆上defer的后边

![](../image/defer4.png)

```go
//go:nosplit
func deferproc(siz int32, fn *funcval) { 
    if getg().m.curg != getg() {
        // go code on the system stack can't defer
        throw("defer on system stack")
    }     
    sp := getcallersp()
    // 参数地址
    argp := uintptr(unsafe.Pointer(&fn)) + unsafe.Sizeof(fn)
    callerpc := getcallerpc()

    //创建defer
    d := newdefer(siz)
    if d._panic != nil {
        throw("deferproc: d.panic != nil after newdefer")
    }
    d.fn = fn
    d.pc = callerpc
    d.sp = sp
    switch siz {
    case 0:
        // Do nothing.
    case sys.PtrSize:
        *(*uintptr)(deferArgs(d)) = *(*uintptr)(unsafe.Pointer(argp))
    default:
        memmove(deferArgs(d), unsafe.Pointer(argp), uintptr(siz))
    }
    
   
    return0()   
}
```

***newdefer***

在堆上创建defer时，如果函数参数大小小于5字节则优先从缓冲中获取，反之则直接从堆上创建。

从缓冲中获取defer时，优先从当前P上的deferpool中获取，如果没有则从全局复制一半(最多)到本地缓冲中，然后再从本地缓冲中获取，如果仍然还没有则直接在堆上创建。

```go
//创建一个defer
func newdefer(siz int32) *_defer {
var d *_defer
// 参数大小
sc := deferclass(uintptr(siz))
gp := getg()
//deferpool    [5][]*_defer
//如果参数大小<5使用缓存
if sc < uintptr(len(p{}.deferpool)) {
    pp := gp.m.p.ptr()
    //如果P上没有deferpool缓存，则从全局sched.deferpool转移一部分到P上
    if len(pp.deferpool[sc]) == 0 && sched.deferpool[sc] != nil {
        // Take the slow path on the system stack so
        // we don't grow newdefer's stack.
        systemstack(func() {
            lock(&sched.deferlock)
            //最多转移一半
            for len(pp.deferpool[sc]) < cap(pp.deferpool[sc])/2 && sched.deferpool[sc] != nil {
                d := sched.deferpool[sc]
                sched.deferpool[sc] = d.link
                d.link = nil
                pp.deferpool[sc] = append(pp.deferpool[sc], d)
            }
            unlock(&sched.deferlock)
        })
    }
    if n := len(pp.deferpool[sc]); n > 0 {
        d = pp.deferpool[sc][n-1]
        pp.deferpool[sc][n-1] = nil
        pp.deferpool[sc] = pp.deferpool[sc][:n-1]
    }
}
//参数大小大于>=5直接分配内存
if d == nil {
    // Allocate new defer+args.
    systemstack(func() {
        total := roundupsize(totaldefersize(uintptr(siz)))
        d = (*_defer)(mallocgc(total, deferType, true))
    })    
}
d.siz = siz
//分配在堆上
d.heap = true
d.link = gp._defer
//defer保存到g上
gp._defer = d
return d
}
```

**freedefer**

defer函数执行完成后，空闲的defer会保存到缓冲中。优先先放入本地缓冲中，如果本地缓冲已满则先将本地缓冲移动一半到全局defer缓冲中在保存到本地。
```go
//go:nosplit
func freedefer(d *_defer) {
	if d._panic != nil {
		freedeferpanic()
	}
	if d.fn != nil {
		freedeferfn()
	}
	if !d.heap {
		return
	}
	sc := deferclass(uintptr(d.siz))
	if sc >= uintptr(len(p{}.deferpool)) {
		return
	}
	pp := getg().m.p.ptr()
	if len(pp.deferpool[sc]) == cap(pp.deferpool[sc]) {	 
		systemstack(func() {
			var first, last *_defer
			for len(pp.deferpool[sc]) > cap(pp.deferpool[sc])/2 {
				n := len(pp.deferpool[sc])
				d := pp.deferpool[sc][n-1]
				pp.deferpool[sc][n-1] = nil
				pp.deferpool[sc] = pp.deferpool[sc][:n-1]
				if first == nil {
					first = d
				} else {
					last.link = d
				}
				last = d
			}
			lock(&sched.deferlock)
			last.link = sched.deferpool[sc]
			sched.deferpool[sc] = first
			unlock(&sched.deferlock)
		})
	}

	// These lines used to be simply `*d = _defer{}` but that
	// started causing a nosplit stack overflow via typedmemmove.
	d.siz = 0
	d.started = false
	d.sp = 0
	d.pc = 0
	// d._panic and d.fn must be nil already.
	// If not, we would have called freedeferpanic or freedeferfn above,
	// both of which throw.
	d.link = nil

	pp.deferpool[sc] = append(pp.deferpool[sc], d)
}
```
defer缓冲

![](../image/defer5.png)