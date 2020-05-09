## 切片源码分析

我们先看一个slice的例子

```go
func main() {
	aa := make([]int, 3)
	aa = append(aa, 1,2)
	fmt.Println(aa)
}
```
由于make与append都是内部函数，所以我们需要先通过反编译来看下，运行时实际调用的是哪些函数

使用`go tool compile -N -S -l demo.go`获取未优化过的汇编代码，关键信息如下，其中最关键的就是`runtime.makeslice`和`runtime.growslice`，分别是用来创建一个切片类型和对切片进行扩容。`slice`源码在`runtime/slice.go`文件中
```go
	0x0041 00065 (demo.go:6)	MOVQ	$3, 8(SP)
	0x004a 00074 (demo.go:6)	MOVQ	$3, 16(SP)
	0x0053 00083 (demo.go:6)	CALL	runtime.makeslice(SB)
    ....
	0x0086 00134 (demo.go:7)	MOVQ	$3, 16(SP)
	0x008f 00143 (demo.go:7)	MOVQ	$3, 24(SP)
	0x0098 00152 (demo.go:7)	MOVQ	$5, 32(SP)
	0x00a1 00161 (demo.go:7)	CALL	runtime.growslice(SB)
	0x00a6 00166 (demo.go:7)	PCDATA	$0, $1
	0x00a6 00166 (demo.go:7)	MOVQ	40(SP), AX
```

切片在运行时的内存结构:
```go
type slice struct {
    //指向存储数据的数组
    array unsafe.Pointer
    //切片长度
    len   int
    //切片容量
	cap   int
}
```

## make

使用`make`创建切片时，运行时实际调用代码如下:
```go
//make创建切片
// et:封装了切片类型信息  len:切片长度 cap:切片容量
func makeslice(et *_type, len, cap int) unsafe.Pointer {
	//容量*每个类型占用的字节   判断是否内存溢出
	mem, overflow := math.MulUintptr(et.size, uintptr(cap))
	if overflow || mem > maxAlloc || len < 0 || len > cap {		 
		mem, overflow := math.MulUintptr(et.size, uintptr(len))
		if overflow || mem > maxAlloc || len < 0 {
			panicmakeslicelen()
		}
		panicmakeslicecap()
	}
    //分配对象内存 返回指针地址
	return mallocgc(mem, et, true)
}
```

## append

`append`在运行时会计算新的切片的容量并将保存数据的底层数组复制一份并返回新的切片。这样修改新的切片中的数据就不会影响到老的切片。

```go
//et:封装了切片类型信息 old:就的切片  cap:新切片的容量(编译器分析出来的)
func growslice(et *_type, old slice, cap int) slice {
	if raceenabled {
		callerpc := getcallerpc()
		racereadrangepc(old.array, uintptr(old.len*int(et.size)), callerpc, funcPC(growslice))
	}
	if msanenabled {
		msanread(old.array, uintptr(old.len*int(et.size)))
	}

	if cap < old.cap {
		panic(errorString("growslice: cap out of range"))
	}

	if et.size == 0 {		 
		return slice{unsafe.Pointer(&zerobase), old.len, cap}
	}
    //计算切片的容量
	newcap := old.cap
	doublecap := newcap + newcap
	if cap > doublecap {
		newcap = cap
	} else {
		if old.len < 1024 {
			newcap = doublecap
		} else {		 
			for 0 < newcap && newcap < cap {
				newcap += newcap / 4
			}		 
			if newcap <= 0 {
				newcap = cap
			}
		}
	}

	var overflow bool
	var lenmem, newlenmem, capmem uintptr	 
	switch {
	case et.size == 1:
		lenmem = uintptr(old.len)
		newlenmem = uintptr(cap)
		capmem = roundupsize(uintptr(newcap))
		overflow = uintptr(newcap) > maxAlloc
		newcap = int(capmem)
	case et.size == sys.PtrSize:
		lenmem = uintptr(old.len) * sys.PtrSize
		newlenmem = uintptr(cap) * sys.PtrSize
		capmem = roundupsize(uintptr(newcap) * sys.PtrSize)
		overflow = uintptr(newcap) > maxAlloc/sys.PtrSize
		newcap = int(capmem / sys.PtrSize)
	case isPowerOfTwo(et.size):
		var shift uintptr
		if sys.PtrSize == 8 {
			// Mask shift for better code generation.
			shift = uintptr(sys.Ctz64(uint64(et.size))) & 63
		} else {
			shift = uintptr(sys.Ctz32(uint32(et.size))) & 31
		}
		lenmem = uintptr(old.len) << shift
		newlenmem = uintptr(cap) << shift
		capmem = roundupsize(uintptr(newcap) << shift)
		overflow = uintptr(newcap) > (maxAlloc >> shift)
		newcap = int(capmem >> shift)
	default:
		lenmem = uintptr(old.len) * et.size
		newlenmem = uintptr(cap) * et.size
		capmem, overflow = math.MulUintptr(et.size, uintptr(newcap))
		capmem = roundupsize(capmem)
		newcap = int(capmem / et.size)
	}

	 
	if overflow || capmem > maxAlloc {
		panic(errorString("growslice: cap out of range"))
	}

	var p unsafe.Pointer
	if et.ptrdata == 0 {
		p = mallocgc(capmem, nil, false)		 
		memclrNoHeapPointers(add(p, newlenmem), capmem-newlenmem)
	} else {		 
		p = mallocgc(capmem, et, true)
		if lenmem > 0 && writeBarrier.enabled {	 
			bulkBarrierPreWriteSrcOnly(uintptr(p), uintptr(old.array), lenmem)
		}
    }
    //copy数据到新的数组
	memmove(p, old.array, lenmem)
	return slice{p, old.len, newcap}
}
```

