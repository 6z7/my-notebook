## 创建G

我们知道g在go代表一个协程，用执行指定的函数。

```
func main() {
  go func() {
     fmt.Println("hi,goroutine ")
  }()
}
```