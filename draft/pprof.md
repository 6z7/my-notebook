profiling是指对应用程序的画像，画像就是应用程序使用 CPU 和内存等资源的情况

路径: /debug/pprof/

绘制图形时，方框越大代表消耗资源越多

如果没有使用默认的DefaultServeMux路由器，则需要注册handler到当前使用的mux上

```
import _ "net/http/pprof"

go func() {
	log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

查看堆的信息：`go tool pprof http://localhost:6060/debug/pprof/heap`

查看30秒内的cpu采样信息：`go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30`


go tool pprof http://localhost:6060/debug/pprof/block

查看5秒的执行堆栈信息 wget  http://localhost:6060/debug/pprof/trace?seconds=5

go tool pprof http://localhost:6060/debug/pprof/mutex

浏览器中查看所有可用的profile http://localhost:6060/debug/pprof/



go tool pprof -top http://localhost:6060/debug/pprof/heap

go tool pprof -png http://localhost:6060/debug/pprof/heap > out.png


pprof 采样数据主要有三种获取方式:

* runtime/pprof: 手动调用runtime.StartCPUProfile或者runtime.StopCPUProfile等 API来生成和写入采样文件，灵活性高
* net/http/pprof: 通过 http 服务获取Profile采样文件，简单易用，适用于对应用程序的整体监控。通过 runtime/pprof 实现
* go test: 通过 go test -bench . -cpuprofile prof.cpu生成采样文件 适用对函数进行针对性测试


通过go test生成profile数据:  
```
// go test -bench=. -benchmem -cpuprofile profile.out      
// go test -bench=. -benchmem -memprofile memprofile.out -cpuprofile profile.out
```

通过http的方式支持的profile种类有:

* allocs: A sampling of all past memory allocations

* block: Stack traces that led to blocking on synchronization primitives

	协程阻塞的情况，可以用来分析和查找死锁等性能瓶颈，默认不开启，需要调用`runtime.SetBlockProfileRate`开启。

* cmdline: The command line invocation of the current program

	获取程序的命令行启动参数

* goroutine: Stack traces of all current goroutines

	协程相关信息，哪些协程在运行

* heap: A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.

* mutex: Stack traces of holders of contended mutexes

	查看互斥的争用情况，默认不开启，需要在调用`runtime.SetMutexProfileFraction`开启。

* profile: CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.

	 获取指定时间内(从请求时开始)的cpuprof，倒计时结束后自动返回。参数: seconds, 默认值为30。cpuprofile 每秒钟采样100次，收集当前运行的 goroutine 堆栈信息

* threadcreate: Stack traces that led to the creation of new OS threads

* trace: A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.


pprof数据分析
```
// 启动shell交互窗口进行分析profile文件
go tool pprof memprofile.out
// 启动一个web分析profile文件
go tool pprof -http=:18080 memprofile.out
// 抓取profile信息保存到文件并启动shell交互窗口进行分析
go tool pprof http://localhost:6060/debug/pprof/profile
```

shell交互窗口中常用命令:

* topN
* list xx ：查看代码
* web [xx]：生成调用图，支持过滤


## cpu profile

`go tool pprof cpuprofile.out` 输出
```
Type: cpu  //profile类型
Time: Mar 20, 2020 at 3:30pm (CST)   // 采样开始时间
Duration: 30s, Total samples = 20ms (0.067%)  //运行时间  采样时间  (采样时间占比)
Entering interactive mode (type "help" for commands, "o" for options)
```

### topN
```
>top 10
Showing nodes accounting for 157.69s, 99.34% of 158.74s total
Dropped 46 nodes (cum <= 0.79s)
      flat  flat%   sum%        cum   cum%
    77.19s 48.63% 48.63%    125.58s 79.11%  runtime.selectnbrecv
    48.23s 30.38% 79.01%     48.25s 30.40%  runtime.chanrecv
    32.16s 20.26% 99.27%    157.74s 99.37%  main.logicCode
     0.07s 0.044% 99.31%      0.89s  0.56%  runtime.newstack
     0.04s 0.025% 99.34%      0.80s   0.5%  runtime.morestack
```

* flat：当前函数占用CPU的耗时(实际采样时间)，不包含调用的子函数的时间
* flat：:当前函数占用CPU的耗时百分比
* sun%：函数占用CPU的耗时累计百分比
* cum：当前函数加上调用的子函数耗时占用CPU的总耗时
* cum%：当前函数加上调用的子函数耗时占用CPU的总耗时百分比
* 最后一列：函数名称


### list
```
(pprof) list main.logicCode
Total: 2.65mins
ROUTINE ======================== main.logicCode in C:\senki\study\go-first\gc\localprofile.go
    32.16s   2.63mins (flat, cum) 99.37% of Total
         .          .     11:// 一段有问题的代码
         .          .     12:func logicCode() {
         .          .     13:   var c chan int
         .          .     14:   for {
         .          .     15:           select {
    32.16s   2.63mins     16:           case v := <-c:
         .          .     17:                   fmt.Printf("recv from chan, value:%v\n", v)
         .          .     18:           default:
         .          .     19:
         .          .     20:           }
         .          .     21:   }

```


## memory profile

```
go tool pprof  memprofile.out`：查看使用的内存的大小
go tool pprof -alloc_space  memprofile.out`：查看使用的内存的大小
go tool pprof -alloc_objects  memprofile.out`：查看使用的内存的大小
```


