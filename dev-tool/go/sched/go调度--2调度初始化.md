# 调度初始化

调度初始化([runtime·schedinit](https://github.com/6z7/go/blob/release-branch.go1.13-study/src/runtime/asm_amd64.s#L231))是在go的启动流程中触发的，具体在`runtime·rt0_go`函数的中:
```
CALL	runtime·args(SB)   
CALL	runtime·osinit(SB)   
CALL	runtime·schedinit(SB) //proc.go
```

`schedinit`进行了许多初始化操作:

* 内存分配器初始化
* gc初始化
* 创建指定数量的P
*

```go
func schedinit() {	
	_g_ := getg()
	if raceenabled {
		_g_.racectx, raceprocctx0 = raceinit()
	}

	sched.maxmcount = 10000

	tracebackinit()
	moduledataverify()
	stackinit()
	mallocinit()
	mcommoninit(_g_.m)
	cpuinit()       // must run before alginit
	alginit()       // maps must not be used before this call
	modulesinit()   // provides activeModules
	typelinksinit() // uses maps, activeModules
	itabsinit()     // uses activeModules

	msigsave(_g_.m)
	initSigmask = _g_.m.sigmask

	goargs()
	goenvs()
	parsedebugvars()
	gcinit()

	sched.lastpoll = uint64(nanotime())
	procs := ncpu
	if n, ok := atoi32(gogetenv("GOMAXPROCS")); ok && n > 0 {
		procs = n
	}
	if procresize(procs) != nil {
		throw("unknown runnable goroutine during bootstrap")
	} 
	if debug.cgocheck > 1 {
		writeBarrier.cgo = true
		writeBarrier.enabled = true
		for _, p := range allp {
			p.wbBuf.reset()
		}
	}

	if buildVersion == "" {	 
		buildVersion = "unknown"
	}
	if len(modinfo) == 1 {	 
		modinfo = ""
	}
}
``` 