# go调度--sysmon

sysmon用来监视P的运行情况。主要进行以下操作:

* netpoll
* 抢占
* 强制GC


sysmon是在runtime.main中触发执行的，通过创建一个新的M执行监控，因为此M仅执行sysmon不需要参与调度所以无需P。

```go
systemstack(func() {		 
			newm(sysmon, nil)
  })
```


通过sysmon的实现可以看到该函数是在进行一个死循环，每次进行循环之前都要sleep一段时间(20us、21us、22us...10ms)，最长睡眠10ms。接下来我们来具体看下每个操作。
```go
func sysmon() {
	lock(&sched.lock)
	sched.nmsys++
	checkdead()
	unlock(&sched.lock)

	lasttrace := int64(0)
	idle := 0 // how many cycles in succession we had not wokeup somebody
	delay := uint32(0)
	for {
		if idle == 0 { // start with 20us sleep...
			delay = 20
		} else if idle > 50 { // start doubling the sleep after 1ms...
			delay *= 2
		}
		if delay > 10*1000 { // up to 10ms
			delay = 10 * 1000
		}
		usleep(delay)
        ......
		// poll network if not polled for more than 10ms
		lastpoll := int64(atomic.Load64(&sched.lastpoll))
		now := nanotime()
		if netpollinited() && lastpoll != 0 && lastpoll+10*1000*1000 < now {
			atomic.Cas64(&sched.lastpoll, uint64(lastpoll), uint64(now))
			list := netpoll(false) // non-blocking - returns list of goroutines
			if !list.empty() {	
				incidlelocked(-1)
				injectglist(&list)
				incidlelocked(1)
			}
		}
	
		if retake(now) != 0 {
			idle = 0
		} else {
			idle++
		}
		// check if we need to force a GC
		if t := (gcTrigger{kind: gcTriggerTime, now: now}); t.test() && atomic.Load(&forcegc.idle) != 0 {
			lock(&forcegc.lock)
			forcegc.idle = 0
			var list gList
			list.push(forcegc.g)
			injectglist(&list)
			unlock(&forcegc.lock)
		}
		if debug.schedtrace > 0 && lasttrace+int64(debug.schedtrace)*1000000 <= now {
			lasttrace = now
			schedtrace(debug.scheddetail > 0)
		}
	}
}

```

## 抢占

sysmon会遍历所有的P，对`_Prunning`和`_Psyscall`状态的P发起抢占，分为以下两种情况:

1. P上的g持续运行时间超过10毫秒，则将P上持续运行的g标记为可抢占，在进入函数时如果检测到抢占标记则会尝试发起抢占。

2. P上的g进行系统调用被阻塞，如果满足以下条件之一，则将P交给一个新的M来执行P上的g
	-  P上有需要运行的g
	- 没有空闲的P和M(自旋m的数量+空闲P的数量=0)
	- 系统调用阻塞超过10毫秒



```go
func retake(now int64) uint32 {
	// 将当前P交出的次数
	n := 0
	// Prevent allp slice changes. This lock will be completely
	// uncontended unless we're already stopping the world.
	lock(&allpLock)
	// We can't use a range loop over allp because we may
	// temporarily drop the allpLock. Hence, we need to re-fetch
	// allp each time around the loop.
	// 遍历所有的P
	for i := 0; i < len(allp); i++ {
		_p_ := allp[i]
		if _p_ == nil {
			// This can happen if procresize has grown
			// allp but not yet created new Ps.
			// procresize中扩容p p还没初始化
			continue
		}
		pd := &_p_.sysmontick
		s := _p_.status
		sysretake := false
		if s == _Prunning || s == _Psyscall {
			// Preempt G if it's running for too long.
			// p被调度次数
			t := int64(_p_.schedtick)
			if int64(pd.schedtick) != t {
				pd.schedtick = uint32(t)
				pd.schedwhen = now
			} else if pd.schedwhen+forcePreemptNS <= now { //g一直在运行 超过了10ms
				// g运行时间太长了 将当前g标记为可抢占
				preemptone(_p_)
				// In case of syscall, preemptone() doesn't
				// work, because there is no M wired to P.
				// 系统调用时 M和P没有关联?  所以无法使用preemptone方法，需要手动设置
				sysretake = true
			}
		}
		// 当前P处于系统调用状态
		if s == _Psyscall {
			// Retake P from syscall if it's there for more than 1 sysmon tick (at least 20us).
			t := int64(_p_.syscalltick)
			if !sysretake && int64(pd.syscalltick) != t {
				// sysmon监控当前运行时 系统调用次数已经发生变换 更新
				pd.syscalltick = uint32(t)
				pd.syscallwhen = now
				continue
			}
			 
			// 时间条件满足的前提下，满足于以下任一条件才去抢占当前P
			// 1.  P上有需要运行的g
			// 2. 自旋m的数量+空闲P的数量==0，
			// 3. 系统调用阻塞超过10毫秒
			if runqempty(_p_) && atomic.Load(&sched.nmspinning)+atomic.Load(&sched.npidle) > 0 && pd.syscallwhen+10*1000*1000 > now {
				continue
			}		 
			unlock(&allpLock)		 
			incidlelocked(-1)
			if atomic.Cas(&_p_.status, s, _Pidle) {
				if trace.enabled {
					traceGoSysBlock(_p_)
					traceProcStop(_p_)
				}
				n++
				_p_.syscalltick++
				// 将当前P交给其它M执行
				handoffp(_p_)
			}
			incidlelocked(1)
			lock(&allpLock)
		}
	}
	unlock(&allpLock)
	return uint32(n)
}
```

**第一种情况下的抢占流程**

对P上持续执行时间超过10ms的g会设置抢占标记:

```go
// 设置抢占标记
func preemptone(_p_ *p) bool {
	mp := _p_.m.ptr()
	// p与m已经解绑或g已经管理它m 直接返回
	if mp == nil || mp == getg().m {
		return false
	}
	gp := mp.curg
	// m与g以解绑或当前g时g0直接返回
	if gp == nil || gp == mp.g0 {
		return false
	}
    //可以抢占标记
	gp.preempt = true

	// Every call in a go routine checks for stack overflow by
	// comparing the current stack pointer to gp->stackguard0.
	// Setting gp->stackguard0 to StackPreempt folds
	// preemption into the normal stack overflow check.
	// 每次函数调用时都会通过g.stackguard0判断栈是否溢出
	// stackPreempt是一个很大的标记值，通过该标记可以知道当前g可以被抢占
	gp.stackguard0 = stackPreempt
	return true
}
```

此处设置了抢占标记，那什么时候使用会使用到呢？具体可以参见：[G抢占.md](./go调度--10G抢占.md)


**第二种情况下抢占流程**

将阻塞的P交给其它M时，如果P的本地队列和全局队列上都没有需要执行的g，则就不需要启动新的M去执行g了，空闲的P放入全局空闲列表。

>优先使用调度器sched上空闲的m，没有再创建新的m执行g

```go
// 将P交给其它M执行
func handoffp(_p_ *p) {	
	// 如果本地队列或全局队列上还有g,则直接运行
	if !runqempty(_p_) || sched.runqsize != 0 {	
		startm(_p_, false)
		return
	}
	// if it has GC work, start it straight away
	if gcBlackenEnabled != 0 && gcMarkWorkAvailable(_p_) {
		startm(_p_, false)
		return
	}
	// no local work, check that there are no spinning/idle M's,
	// otherwise our help is not required
	// 没有空闲的P M也都在忙没有在自旋 那么新起的M只能先自旋了
	if atomic.Load(&sched.nmspinning)+atomic.Load(&sched.npidle) == 0 && atomic.Cas(&sched.nmspinning, 0, 1) { // TODO: fast atomic
		startm(_p_, true)
		return
	}
	lock(&sched.lock)
	if sched.gcwaiting != 0 {
		_p_.status = _Pgcstop
		sched.stopwait--
		if sched.stopwait == 0 {
			notewakeup(&sched.stopnote)
		}
		unlock(&sched.lock)
		return
	}
	if _p_.runSafePointFn != 0 && atomic.Cas(&_p_.runSafePointFn, 1, 0) {
		sched.safePointFn(_p_)
		sched.safePointWait--
		if sched.safePointWait == 0 {
			notewakeup(&sched.safePointNote)
		}
	}
	// 全局队列上还有g
	if sched.runqsize != 0 {
		unlock(&sched.lock)
		startm(_p_, false)
		return
	}
	// If this is the last running P and nobody is polling network,
	// need to wakeup another M to poll network.
	if sched.npidle == uint32(gomaxprocs-1) && atomic.Load64(&sched.lastpoll) != 0 {
		unlock(&sched.lock)
		startm(_p_, false)
		return
	}
	// P放入全局空闲列表
	pidleput(_p_)
	unlock(&sched.lock)
} 

// 启动M运行P
func startm(_p_ *p, spinning bool) {
	lock(&sched.lock)
	//没有指定p的话需要从p的空闲队列中获取一个p
	if _p_ == nil {
		_p_ = pidleget()
		if _p_ == nil {
			unlock(&sched.lock)
			if spinning {
				// The caller incremented nmspinning, but there are no idle Ps,
				// so it's okay to just undo the increment and give up.
				if int32(atomic.Xadd(&sched.nmspinning, -1)) < 0 {
					throw("startm: negative nmspinning")
				}
			}
			return
		}
	}
	
	// 空闲的m
	mp := mget()
	unlock(&sched.lock)
	if mp == nil { 
		var fn func()
		if spinning {			
			// 新线程启动时 执行此函数
			fn = mspinning
		}
		//创建新的工作线程
		newm(fn, _p_)
		return
	}
	if mp.spinning {
		throw("startm: m is spinning")
	}
	if mp.nextp != 0 {
		throw("startm: m has p")
	}
	if spinning && !runqempty(_p_) {
		throw("startm: p has runnable gs")
	}	 
	mp.spinning = spinning
	mp.nextp.set(_p_)
	//唤醒处于休眠状态的工作线程
	// 当在schedule过程中没有获取到需要执行的g时会挂起M,此处唤醒M去执行新的P上的任务
	notewakeup(&mp.park)
}
```


