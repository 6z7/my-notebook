# map源码分析


```go
    // make(map[string]int, 3/*hint*/)
	var m = map[string]int{
		"banana1": 1,
		"banana2": 2,
		"banana3": 3,
	 
	}
	var m2 = make(map[string]int, 9/*hint*/)
```
编译器根据不同的条件使用不同的方式创建hmap:

* makemap_small

* makemap64

* makemap


```go
// hint不大于8时
func makemap_small() *hmap {
	h := new(hmap)
	h.hash0 = fastrand()
	return h
}
 
func makemap64(t *maptype, hint int64, h *hmap) *hmap {
	if int64(int(hint)) != hint {
		hint = 0
	}
	return makemap(t, int(hint), h)
}

// bucket数组是否创建根据情况而定
func makemap(t *maptype, hint int, h *hmap) *hmap {
    // hint*size(bucket)使用的字节 在指针返回内？？
	mem, overflow := math.MulUintptr(uintptr(hint), t.bucket.size)
	if overflow || mem > maxAlloc {
		hint = 0
	}

	// 未在栈上创建hmap结构
	if h == nil {
		h = new(hmap)
	}
	h.hash0 = fastrand()
 
	B := uint8(0)
	// 找到满足扩容条件的最小B值
	// 如果hit<=8||hit<=6.5*2^B 不需要扩容
	for overLoadFactor(hint, B) {
		B++
	}
	// B可能为0
	h.B = B
	
	// 如果B=0则在使用时在创建bucket数组
	if h.B != 0 {
		var nextOverflow *bmap
		h.buckets, nextOverflow = makeBucketArray(t, h.B, nil)
		if nextOverflow != nil {
			h.extra = new(mapextra)
			h.extra.nextOverflow = nextOverflow
		}
	}

	return h
}
```

通过以上方式编译器创建了map结构`hmap`，我们先看下它的结构：
```go
type hmap struct {	
	// map中的key数量
	count     int  
	//标记
	flags     uint8
	// hashtable中bucket数量2^B
	// 负载因子计算 loadFactor * 2^B
	B         uint8 
	// 溢出bucket大概数量
	// B<16时 每次分配溢出bucket都统计
	// B>=16时 不一定统计
	noverflow uint16  
	//hash种子 一个随机数
	hash0     uint32  
    // 指向bucket数组的指针
	buckets    unsafe.Pointer  
	// growing 时保存原buckets的指针
	oldbuckets unsafe.Pointer  
    // growing 时已迁移的个数 迁移进度
    nevacuate  uintptr
    // 保存一些扩展信息        
	extra *mapextra 
}
```
bucket的结构：
```go
type bmap struct {

	// top hash包含该bucket中每个键的hash值的高八位。
	// 如果tophash[0]小于mintophash，则tophash[0]为桶疏散状态
	// hash(key)的前8位
    tophash [bucketCnt]uint8
    
    // 后面是8个key和value 存储格式：k1k2...k8v1v2...v8, 避免了因为cpu要求固定长度读取，字节对齐，造成的空间浪费
    
	// 后边是一个overflow指针,指向当前bucket的溢出bucket	
	
}

```
mapextra
```go
type mapextra struct {	
	// 保存创建的溢出bucket
    overflow    *[]*bmap
    // 扩容时旧的溢出bucket
	oldoverflow *[]*bmap	 
    // 预分配的bucket的指针，
    // 指向创建bucket数组时预先多创建的部分
	nextOverflow *bmap
}
```
分配bucket时，根据情况的不同，可能会预创建一些bucket备用，减少bucket不足造成频繁创建的问题
```go
func makeBucketArray(t *maptype, b uint8, dirtyalloc unsafe.Pointer) (buckets unsafe.Pointer, nextOverflow *bmap) {
	//2^b
	// 期望的bucket数量
	base := bucketShift(b)
	// 实际创建的bucket数量
	nbuckets := base
	//对于小b，不太可能出现溢出桶。
	//避免计算的开销。
	if b >= 4 {		
		//加上估计的溢出桶数
		//插入元素的中位数
		//与此值b一起使用。
		nbuckets += bucketShift(b - 4)
		sz := t.bucket.size * nbuckets
		up := roundupsize(sz)
		if up != sz {
			nbuckets = up / t.bucket.size
		}
	}

	if dirtyalloc == nil {
		buckets = newarray(t.bucket, int(nbuckets))
	} else {		
		buckets = dirtyalloc
		size := t.bucket.size * nbuckets
		if t.bucket.ptrdata != 0 {
			memclrHasPointers(buckets, size)
		} else {
			memclrNoHeapPointers(buckets, size)
		}
	}

	// 预分配bucket
	if base != nbuckets {
		
		// 预分配bucket的起始位置
		nextOverflow = (*bmap)(add(buckets, base*uintptr(t.bucketsize)))
		// 预分配bucket的结束位置
		last := (*bmap)(add(buckets, (nbuckets-1)*uintptr(t.bucketsize)))
        //  预分配bucket的最后一个bucket的溢出指针指向当前bucket数组，
        // 即预分配的的最后一个bucket的溢出指针不等于null，后边会用到？？？
    
		last.setoverflow(t, (*bmap)(buckets))
	}
	return buckets, nextOverflow
}
```

## 赋值部分

`mapassign_faststr`
`mapassign_fast64`
`mapassign_fast64ptr`
`mapassign_fast32`
`mapassign_fast32ptr`
`mapassign`

主要包括：

* 是否并发写
* 计算key所属的bucket
* 如果当前bucket正在进行扩容，则先先将当前bucket迁移到新bucket
* 遍历bucket中的8个cell,比价tophash和key
* 如果找到key，则覆盖旧值
* 如果未找到，则找一个cell存储新的kv
* 根据条件判断是否需要扩容
* 如果bucket的cell已满，则先创建一个溢出bucket，在进行保存kv

以`mapassign_faststr`为例，具体看下赋值过程：

先计算hash(key)，找到key所属于的bucket,如果并发写直接panic。如果当前bucket正在扩容,则先完成bucket迁移在进行后续操作，则如下

```go
unc mapassign_faststr(t *maptype, h *hmap, s string) unsafe.Pointer {
	if h == nil {
		panic(plainError("assignment to entry in nil map"))
	}
	if raceenabled {
		callerpc := getcallerpc()
		racewritepc(unsafe.Pointer(h), callerpc, funcPC(mapassign_faststr))
	}
	if h.flags&hashWriting != 0 {
		throw("concurrent map writes")
	}
	key := stringStructOf(&s)
	// hash(key)
	hash := t.hasher(noescape(unsafe.Pointer(&s)), uintptr(h.hash0))
	
	h.flags ^= hashWriting

	if h.buckets == nil {
		h.buckets = newobject(t.bucket) // newarray(t.bucket, 1)
    }
again:
	// key所属的bucket
	bucket := hash & bucketMask(h.B)
	if h.growing() {
		growWork_faststr(t, h, bucket) //进行迁移
	}
	// key所属bucket的指针
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.bucketsize)))
	top := tophash(hash)

	//保存key的bucket
	var insertb *bmap
	//key保存buccket中的哪个位置
	var inserti uintptr
	var insertk unsafe.Pointer
```
遍历bucket中的8个cell,先使用key的前8个字节进行快速比较，最终在当前bucket或溢出bucket中找到相等key的cell或找到一个空的cell,如下
```go
bucketloop:
	for {
		// 遍历bucket中的8个kv
		for i := uintptr(0); i < bucketCnt; i++ {
			// 比较高8位是否相等，用于快速判断
			if b.tophash[i] != top {
                // 找到一个空的cell备用(初始化或被标记为删除都认为是空),需要遍历所有的cell才能知道是否存在相等的key
				// 如果找不到相等的key，则使用这个cell
				if isEmpty(b.tophash[i]) && insertb == nil {
					insertb = b
					inserti = i
				}

                // emptyRest当前和后边cell中的key已经被删除，无需在遍历
				if b.tophash[i] == emptyRest {
					break bucketloop
				}
				continue
			}
			// 到这里说明在当前bucket找到了高8字节相同的条目了

			// dataOffset=8 hmap中kv偏移位置
			// 2*sys.PtrSize string类型占用的字节
			// bucket中对应项的key位置
			k := (*stringStruct)(add(unsafe.Pointer(b), dataOffset+i*2*sys.PtrSize))
			//高8字节相同的情况下，还要比较是否hash key是否完全一致
			if k.len != key.len {
				continue
			}
			if k.str != key.str && !memequal(k.str, key.str, uintptr(key.len)) {
				continue
			}		
			// 找到对应的key
			inserti = i
			insertb = b
			goto done
		}
		// 下一个溢出bucket位置
		ovf := b.overflow(t)
		if ovf == nil {
			break
		}
		b = ovf
	}
 ```   
 
bucket和其溢出bucket遍历完成后，如果加上当前要新增的key满足扩容条件，则先进行扩容，等扩容完成后在重新寻找key的位置，如下

```go
	// 达到最大负载或bucket溢出超过阀值 但是还没开始扩容 则尝试扩容
	if !h.growing() && (overLoadFactor(h.count+1, h.B) || tooManyOverflowBuckets(h.noverflow, h.B)) {
		hashGrow(t, h)  // 扩容
		goto again
    }
```
如果遍历完bucket和bucket的溢出bucket，仍然没找到可以存储key的位置，则说明bucket已满，需要创建一个新的溢出bucket进行存储kv,当前bucket的溢出指针指向新创建的溢出bucket。计算kv存储的位置，map的数量+1，如下

```go

	// key对应的bucket和溢出bucket已满，则创建一个溢出bucet
	if insertb == nil {		
		// 创建一个溢出bucket
		insertb = h.newoverflow(t, b)
		inserti = 0
	}
	insertb.tophash[inserti&(bucketCnt-1)] = top

	insertk = add(unsafe.Pointer(insertb), dataOffset+inserti*2*sys.PtrSize)

	*((*stringStruct)(insertk)) = *key
	h.count++

done:
	// key对应的value保存位置
	// dataOffset:kv偏移位置
	// bucketCnt*2*sys.PtrSize:跳过k的位置
	// inserti*uintptr(t.elemsize):key对应的value位置
	elem := add(unsafe.Pointer(insertb), dataOffset+bucketCnt*2*sys.PtrSize+inserti*uintptr(t.elemsize))
	if h.flags&hashWriting == 0 {
		throw("concurrent map writes")
	}
	h.flags &^= hashWriting
	return elem
}
```

## 创建溢出bucket过程

如果存在预分配的bucket，则先从预分配的bucket中获取一个bucket,反之在创建一个bucket，统计溢出bucket数量，在extra上保存创建的bucket指针

```go
// 创建一个溢出bucket并关联到b上的溢出指针上
func (h *hmap) newoverflow(t *maptype, b *bmap) *bmap {
	var ovf *bmap
	if h.extra != nil && h.extra.nextOverflow != nil {		
		ovf = h.extra.nextOverflow
		// 预分配的bucket的末尾是不为nil的
		if ovf.overflow(t) == nil {		
            // 从预分配bucket中获取一个bucket
			h.extra.nextOverflow = (*bmap)(add(unsafe.Pointer(ovf), uintptr(t.bucketsize)))
		} else {	
            // 预分配的最后一个bucket，清空在创建bucket阶段设置的标记值		
			ovf.setoverflow(t, nil)
			h.extra.nextOverflow = nil
		}
	} else {
		ovf = (*bmap)(newobject(t.bucket))
	}
	// 统计溢出bucket数量
    h.incrnoverflow()
    // extra上记录溢出bucket位置
	if t.bucket.ptrdata == 0 {
		h.createOverflow()
		*h.extra.overflow = append(*h.extra.overflow, ovf)
	}
	b.setoverflow(t, ovf)
	return ovf
}
```

## 扩容过程

扩容条件
* (map中当前数量+1)>8且>6.5*bucket数量
* 溢出桶数量超过阀值2^15 
```go

func overLoadFactor(count int, B uint8) bool {
	return count > bucketCnt && uintptr(count) > loadFactorNum*(bucketShift(B)/loadFactorDen)
}


func tooManyOverflowBuckets(noverflow uint16, B uint8) bool {	 
	if B > 15 {
		B = 15
	}	 
	return noverflow >= uint16(1)<<(B&15)
}
```

判断是达到负载系数还是溢出bucket过多导致的扩容，如果是后者，则不会改变map的容量。
创建新的bucket数组和预分配的bucket用于扩容。这个过程并没有进行实际的kv的迁移，只进行了扩容所需bucket数组的准备操作。

```go
func hashGrow(t *maptype, h *hmap) {    
	bigger := uint8(1)
	if !overLoadFactor(h.count+1, h.B) {
		bigger = 0
		h.flags |= sameSizeGrow
	}
	oldbuckets := h.buckets
	// 创建新的bucket
	// nextOverflow：预分配的bucket
	newbuckets, nextOverflow := makeBucketArray(t, h.B+bigger, nil)  
	flags := h.flags &^ (iterator | oldIterator)
	if h.flags&iterator != 0 {
		flags |= oldIterator
	}	
	h.B += bigger
	h.flags = flags
	h.oldbuckets = oldbuckets
    h.buckets = newbuckets
    // 迁移的bucket数量
    h.nevacuate = 0
    // 溢出桶数量
	h.noverflow = 0

    // 保存旧的预分配bucket
	if h.extra != nil && h.extra.overflow != nil {	
		if h.extra.oldoverflow != nil {
			throw("oldoverflow is not nil")
		}
		h.extra.oldoverflow = h.extra.overflow
		h.extra.overflow = nil
    }
     // 保存新的预分配bucket
	if nextOverflow != nil {
		if h.extra == nil {
			h.extra = new(mapextra)
		}
		h.extra.nextOverflow = nextOverflow
	} 
}
```

##  bucket迁移过程

将属于当前bucket的kv从旧的bucket中迁移过来，如果扩容过程还在进行，则在多迁移一个bucket

```go
func growWork_faststr(t *maptype, h *hmap, bucket uintptr) {
	
	evacuate_faststr(t, h, bucket&h.oldbucketmask())
	 
	if h.growing() {
		evacuate_faststr(t, h, h.nevacuate)
	}
}
```