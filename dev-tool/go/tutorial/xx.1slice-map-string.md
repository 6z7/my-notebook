切片源码参见:

字符串源码参见:

map源码参见:

### 字符串与字节切片之间的高效转换

字符串内存结构:
```go
type stringStruct struct {
    str unsafe.Pointer
    len int
}
```
字节切片内存结构:
```go
type slice struct {
    array unsafe.Pointer
    len   int
    cap   int
}
```

字符串与字节切片底层数据都是通过字节数组保存的且初始化切片或字符串的参数都可以拿到，所以通过`unsafe`包可以直接操作内存进行无需复制的相互转换。

```go
func main() {
	str := byte2Str(str2byte("北京"))
	fmt.Println(str)
}

func str2byte(str string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&str))
	s := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&s))
}

func byte2Str(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}
```