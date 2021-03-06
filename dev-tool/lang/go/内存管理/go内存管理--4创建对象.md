# go内存管理--创建对象

[`newobject`](https://github.com/6z7/go/blob/15b78db9cb8b111d93835e4710adb70e4b437c11/src/runtime/malloc.go#L1272)用于创建对象。根据对象的不同大小，会采用不同的策略分配进行创建对象所需的内存。

下面我们先看下创建对象的整体流程。

1. 根据待分配对象的大小是否大于32KB，决定走不同的流程

    - 不大于32KB分配流程:  mcache-->mcentral-->mheap-->os
    - 大于32kb分配流程: mheap-->os


```go
// 创建对象
func newobject(typ *_type) unsafe.Pointer {
	return mallocgc(typ.size, typ, true)
}
```

```go
// size:对象大小
// typ:对象类型
// needzero:是否需要清零
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
    // ......

	mp := acquirem()
	if mp.mallocing != 0 {
		throw("malloc deadlock")
	}
	if mp.gsignal == getg() {
		throw("malloc during signal")
	}
	// 标记正在分配内存
	mp.mallocing = 1

	shouldhelpgc := false
	dataSize := size
	c := gomcache()
	var x unsafe.Pointer
	// noscan为true代表对象不包含指针
	noscan := typ == nil || typ.ptrdata == 0
	// 对象小于32kb
	if size <= maxSmallSize {
		if noscan && size < maxTinySize {		 
			off := c.tinyoffset			 
			if size&7 == 0 {
				off = round(off, 8)
			} else if size&3 == 0 {
				off = round(off, 4)
			} else if size&1 == 0 {
				off = round(off, 2)
			}
			if off+size <= maxTinySize && c.tiny != 0 {
				// The object fits into existing tiny block.
				x = unsafe.Pointer(c.tiny + off)
				c.tinyoffset = off + size
				c.local_tinyallocs++
				mp.mallocing = 0
				releasem(mp)
				return x
			}
			// Allocate a new maxTinySize block.
			span := c.alloc[tinySpanClass]
			v := nextFreeFast(span)
			if v == 0 {
				v, _, shouldhelpgc = c.nextFree(tinySpanClass)
			}
			x = unsafe.Pointer(v)
			(*[2]uint64)(x)[0] = 0
            (*[2]uint64)(x)[1] = 0
            		
			if size < c.tinyoffset || c.tiny == 0 {
				c.tiny = uintptr(x)
				c.tinyoffset = size
			}
			size = maxTinySize
		} else {
			// 计算处对象的大小所属于的size class, 此处是szie calss数组的索引
			var sizeclass uint8
			// 小于1kb的对象
			if size <= smallSizeMax-8 {
				sizeclass = size_to_class8[(size+smallSizeDiv-1)/smallSizeDiv]
			} else {
				// 1kb~32kb的对象
				sizeclass = size_to_class128[(size-smallSizeMax+largeSizeDiv-1)/largeSizeDiv]
			}
			// 对象应该分配的大小
			size = uintptr(class_to_size[sizeclass])
			// 计算sizeclss所属的spaclass
			spc := makeSpanClass(sizeclass, noscan)
			span := c.alloc[spc]
			v := nextFreeFast(span)
			if v == 0 {
				v, span, shouldhelpgc = c.nextFree(spc)
			}
			x = unsafe.Pointer(v)
			if needzero && span.needzero != 0 {
				memclrNoHeapPointers(unsafe.Pointer(v), size)
			}
		}
	} else {
		var s *mspan
		shouldhelpgc = true
		systemstack(func() {
			s = largeAlloc(size, needzero, noscan)
		})
		s.freeindex = 1
		s.allocCount = 1
		x = unsafe.Pointer(s.base())
		size = s.elemsize
	}

	var scanSize uintptr
	if !noscan {	 
		if typ == deferType {
			dataSize = unsafe.Sizeof(_defer{})
		}
		heapBitsSetType(uintptr(x), size, dataSize, typ)
		if dataSize > typ.size {
			// Array allocation. If there are any
			// pointers, GC has to scan to the last
			// element.
			if typ.ptrdata != 0 {
				scanSize = dataSize - typ.size + typ.ptrdata
			}
		} else {
			scanSize = typ.ptrdata
		}
		c.local_scan += scanSize
	}
	
    publicationBarrier()
    
	if gcphase != _GCoff {
		gcmarknewobject(uintptr(x), size, scanSize)
	}

	if raceenabled {
		racemalloc(x, size)
	}

	if msanenabled {
		msanmalloc(x, size)
	}

	mp.mallocing = 0
	releasem(mp)

	if debug.allocfreetrace != 0 {
		tracealloc(x, size, typ)
	}

	if rate := MemProfileRate; rate > 0 {
		if rate != 1 && size < c.next_sample {
			c.next_sample -= size
		} else {
			mp := acquirem()
			profilealloc(mp, x, size)
			releasem(mp)
		}
	}

	if assistG != nil {
		// Account for internal fragmentation in the assist
		// debt now that we know it.
		assistG.gcAssistBytes -= int64(size - dataSize)
	}

	if shouldhelpgc {
		if t := (gcTrigger{kind: gcTriggerHeap}); t.test() {
			gcStart(t)
		}
	}

	return x
}
```

## 大对象分配流程

1. 根据对象大小计算需要分配的内存页数，并进行对齐
2. 将对象对应的size class(大对象是0)转为span class
3. 从堆上为span分配内存
4. 初始化span和bitmap

```go
func largeAlloc(size uintptr, needzero bool, noscan bool) *mspan {
	// print("largeAlloc size=", size, "\n")

	if size+_PageSize < size {
		throw("out of memory")
	}
	// 计算需要的内存页数量
	npages := size >> _PageShift
	// 对象大小不是页的整数倍
	if size&_PageMask != 0 {
		npages++
    }
     
	deductSweepCredit(npages*_PageSize, npages)
    // sizeClass:0代表大对象 直接从os分配内存
	s := mheap_.alloc(npages, makeSpanClass(0, noscan), true, needzero)
	if s == nil {
		throw("out of memory")
	}
	s.limit = s.base() + size
	heapBitsForAddr(s.base()).initSpan(s)
	return s
}
```

**将size class转为对应的span class**

span class根据是否包含指针数据分为两种，所以span class的种类是size class的2倍(2*67)。

大对象的span class索引等于0
```go
// 根据sizeclass和noscan计算所属于的spanClass
// noscan为true代表对象不包含指针
func makeSpanClass(sizeclass uint8, noscan bool) spanClass {
	return spanClass(sizeclass<<1) | spanClass(bool2int(noscan))
}
```

## 小对象分配流程


## 微小对象分配流程
