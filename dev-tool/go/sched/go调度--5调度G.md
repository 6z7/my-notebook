## go调度--调度G

通过[go调度--4创建g](./go调度--4创建g.md)一文的分析，我们知道了G是如何创建出来的，现在我们来看下创建出来的G是如何被调度执行的。

在`runtime·rt0_go`函数的汇编代码中有段`CALL	runtime·mstart(SB)`指令，这段代码就是用于启动调度流程来执行G的入口。

## mstart

mstat是go启动时执行g的入口，可以看到mstart中实际是去调用mstart1方法
```go
func mstart() {
	_g_ := getg()

    //g0已经设置了lo 此处osStack=false
	osStack := _g_.stack.lo == 0
	if osStack {
		// Initialize stack bounds from system stack.
		// Cgo may have left stack size in stack.hi.
		// minit may update the stack bounds.
		size := _g_.stack.hi
		if size == 0 {
			size = 8192 * sys.StackGuardMultiplier
		}
		_g_.stack.hi = uintptr(noescape(unsafe.Pointer(&size)))
		_g_.stack.lo = _g_.stack.hi - size + 1024
	}
	// Initialize stack guard so that we can start calling regular
	// Go code.
	_g_.stackguard0 = _g_.stack.lo + _StackGuard
	// This is the g0, so we can also call go:systemstack
	// functions, which check stackguard1.
	_g_.stackguard1 = _g_.stackguard0
	mstart1()

	// Exit this thread.
	switch GOOS {
	case "windows", "solaris", "illumos", "plan9", "darwin", "aix":
		// Windows, Solaris, illumos, Darwin, AIX and Plan 9 always system-allocate
		// the stack, but put it in _g_.stack before mstart,
		// so the logic above hasn't set osStack yet.
		osStack = true
	}
	mexit(osStack)
}
```

## mstart1

* mstart1保存pc和sp到当前g的gobuf中(保存当前的寄存器等上下文信息的结构)，恢复到这里
* 执行当前g关联的M上设置的mstartfn函数
* 如果P和M还没绑定则先绑定
* 执行调度方法schedule
```go
func mstart1() {
	_g_ := getg()

	if _g_ != _g_.m.g0 {
		throw("bad runtime·mstart")
	}

	// Record the caller for use as the top of stack in mcall and
	// for terminating the thread.
	// We're never coming back to mstart1 after we call schedule,
	// so other calls can reuse the current frame.
	save(getcallerpc(), getcallersp())
	asminit()
	minit()

	// Install signal handlers; after minit so that minit can
	// prepare the thread to be able to handle the signals.
	if _g_.m == &m0 {
		mstartm0()
	}

	if fn := _g_.m.mstartfn; fn != nil {
		fn()
	}

	if _g_.m != &m0 {
		acquirep(_g_.m.nextp.ptr())
		_g_.m.nextp = 0
	}
	schedule()
}
```

## schedule

schedule中我们先仅关注与启动相关的流程，看看是如何获取待运行的G，然后进行调度执行的。

* 每个P每调度61次就从g全局队列上获取g来执行
* 先从P上本地队列获取g,如果没有则从全局队列获取g，还不行就去从其它P上偷，如果还没有g，则将当前使用的m放入全局空闲m链表中，然后挂起当前m等待唤醒后在继续执行
* 调用execute执行获取到的g
```go
func schedule() {
	_g_ := getg()

	if _g_.m.locks != 0 {
		throw("schedule: holding locks")
	}

	if _g_.m.lockedg != 0 {
		stoplockedm()
		execute(_g_.m.lockedg.ptr(), false) // Never returns.
	}

	// We should not schedule away from a g that is executing a cgo call,
	// since the cgo call is using the m's g0 stack.
	if _g_.m.incgo {
		throw("schedule: in cgo")
	}

top:
	if sched.gcwaiting != 0 {
		gcstopm()
		goto top
	}
	if _g_.m.p.ptr().runSafePointFn != 0 {
		runSafePointFn()
	}

	var gp *g
	var inheritTime bool

	// Normal goroutines will check for need to wakeP in ready,
	// but GCworkers and tracereaders will not, so the check must
	// be done here instead.
	tryWakeP := false
	if trace.enabled || trace.shutdown {
		gp = traceReader()
		if gp != nil {
			casgstatus(gp, _Gwaiting, _Grunnable)
			traceGoUnpark(gp, 0)
			tryWakeP = true
		}
	}
	if gp == nil && gcBlackenEnabled != 0 {
		gp = gcController.findRunnableGCWorker(_g_.m.p.ptr())
		tryWakeP = tryWakeP || gp != nil
	}
	if gp == nil {
		// Check the global runnable queue once in a while to ensure fairness.
		// Otherwise two goroutines can completely occupy the local runqueue
		// by constantly respawning each other.
		if _g_.m.p.ptr().schedtick%61 == 0 && sched.runqsize > 0 {
			lock(&sched.lock)
			gp = globrunqget(_g_.m.p.ptr(), 1)
			unlock(&sched.lock)
		}
	}
	if gp == nil {
		gp, inheritTime = runqget(_g_.m.p.ptr())
		if gp != nil && _g_.m.spinning {
			throw("schedule: spinning with local work")
		}
	}
	if gp == nil {
		gp, inheritTime = findrunnable() // blocks until work is available
	}

	// This thread is going to run a goroutine and is not spinning anymore,
	// so if it was marked as spinning we need to reset it now and potentially
	// start a new spinning M.
	if _g_.m.spinning {
		resetspinning()
	}

	if sched.disable.user && !schedEnabled(gp) {
		// Scheduling of this goroutine is disabled. Put it on
		// the list of pending runnable goroutines for when we
		// re-enable user scheduling and look again.
		lock(&sched.lock)
		if schedEnabled(gp) {
			// Something re-enabled scheduling while we
			// were acquiring the lock.
			unlock(&sched.lock)
		} else {
			sched.disable.runnable.pushBack(gp)
			sched.disable.n++
			unlock(&sched.lock)
			goto top
		}
	}

	// If about to schedule a not-normal goroutine (a GCworker or tracereader),
	// wake a P if there is one.
	if tryWakeP {
		if atomic.Load(&sched.npidle) != 0 && atomic.Load(&sched.nmspinning) == 0 {
			wakep()
		}
	}
	if gp.lockedm != 0 {
		// Hands off own p to the locked m,
		// then blocks waiting for a new p.
		startlockedm(gp)
		goto top
	}

	execute(gp, inheritTime)
}
```

## execute

* G转台修改为`_Grunning`
* 调用gogo切换到指定的g来执行
```go
func execute(gp *g, inheritTime bool) {
	_g_ := getg()

	casgstatus(gp, _Grunnable, _Grunning)
	gp.waitsince = 0
	gp.preempt = false
	gp.stackguard0 = gp.stack.lo + _StackGuard
	if !inheritTime {
		_g_.m.p.ptr().schedtick++
	}
	_g_.m.curg = gp
	gp.m = _g_.m

	// Check whether the profiler needs to be turned on or off.
	hz := sched.profilehz
	if _g_.m.profilehz != hz {
		setThreadCPUProfiler(hz)
	}

	if trace.enabled {
		// GoSysExit has to happen when we have a P, but before GoStart.
		// So we emit it here.
		if gp.syscallsp != 0 && gp.sysblocktraced {
			traceGoSysExit(gp.sysexitticks)
		}
		traceGoStart()
	}

	gogo(&gp.sched)
}
```

## gogo

* 将当前g的寄存器保存到当前g.gobuf中，恢复目标g.gobuf中的信息到寄存器中，主要是pc、sp寄存器
* 跳转到pc寄存器指定的位置继续执行，启动时就是main方法

```
TEXT runtime·gogo(SB), NOSPLIT, $16-8
    //buf = &gp.sched
	MOVQ	buf+0(FP), BX		// gobuf 传入的g的寄存器数据
	MOVQ	gobuf_g(BX), DX    //DX=gp.sched.g
	MOVQ	0(DX), CX		// make sure g != nil
	get_tls(CX)
	//把要运行的g的指针放入线程本地存储，这样后面的代码就可以通过线程本地存储
    //获取到当前正在执行的goroutine的g结构体对象，从而找到与之关联的m和p
	MOVQ	DX, g(CX)
	// 把CPU的SP寄存器设置为sched.sp，完成了栈的切换
	MOVQ	gobuf_sp(BX), SP	// restore SP
	//下面三条同样是恢复调度上下文到CPU相关寄存器
	MOVQ	gobuf_ret(BX), AX
	MOVQ	gobuf_ctxt(BX), DX
	MOVQ	gobuf_bp(BX), BP
	//清空sched的值，因为我们已把相关值放入CPU对应的寄存器了，不再需要
	MOVQ	$0, gobuf_sp(BX)	// clear to help garbage collector
	MOVQ	$0, gobuf_ret(BX)
	MOVQ	$0, gobuf_ctxt(BX)
	MOVQ	$0, gobuf_bp(BX)
	//把sched.pc值放入BX寄存器
	MOVQ	gobuf_pc(BX), BX
	// JMP把BX寄存器的包含的地址值放入CPU的pc寄存器
    // CPU跳转到该地址继续执行指令
	// 第一次执行实际跳转到runtime.main
	JMP	BX
```

## 总结

通过上述分析，我们可以看到程序启动时G的调度流程：

runtime·rt0_go-->mstart-->mstart1-->schedule-->execute-->gogo-->runtime.main



