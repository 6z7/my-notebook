先看下要用到的数据结构与常变定义
    
    //互斥锁
    //g通过竞争获取mutex状态的修改所有权，所有g阻塞在信号量的队列上，
    //当unlock后排在队列头部的g会与新的g竞争所有权(即唤醒的g不一定能够获取到mutex的所有权)，
    //当g被唤醒后如果等待时间大于1ms,则mutex状态会被标记为饥饿状态,
    //如果当前获取锁的g处于饥饿状态,则新的g不会自旋,并将当前g放到队列的首位，下次唤醒直接执行该g
    //
    type Mutex struct {
        //互斥锁状态
        //向移动3位置表示等待获取锁的goroutine数量
        state int32
        //信号量
        sema uint32
    }


    const (
	//已获取锁1
	mutexLocked = 1 << iota // mutex is locked
	//已释放锁2
	mutexWoken
	//饥饿模式(排在前边的go协程一直未获取到锁)2^2
	mutexStarving
	//3 表示mutex.state右移3位后即为等待的goroutine的数量
	mutexWaiterShift = iota

	//互斥锁2种模式：正常模式，饥饿模式
	//正常模式下waiter按照FIFO顺序排队，但是唤醒时会与新的goroutine竞争mutex,
	//新的goroutine应为已经在CPU上运行会比新唤醒的goroutine更有优势获取到mutex,
	//在这种情况下，如果waiter等待获取mutex超过1ms，则将该waiter放到队列的前面，同时锁状态切换到饥饿模式。
	//
	// 饥饿模式下，mutex的所有权直接从unlock goruntine交到队列头部的waiter。
    // 新的goroutine直接排到队列的尾部，不会尝试获mutex。
	//
	// 如果waiter获取到mutex的后满足以下情况，则恢复到正常模式：
	// 1.队列中最后一个waiter
	// 2.获取Mutex的时间小于1ms
	 
	//切换到饥饿模式的阀值1ms
	starvationThresholdNs = 1e6
    )

    //锁的接口定义
    type Locker interface {
	    Lock()
	    Unlock()
    }

## Lock

获取锁

    func (m *Mutex) Lock() {
        // Fast path: grab unlocked mutex.
        if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
            if race.Enabled {
                race.Acquire(unsafe.Pointer(m))
            }
            return
        }
        // Slow path (outlined so that the fast path can be inlined)
        m.lockSlow()
    }

    func (m *Mutex) lockSlow() {
        //开始等待的时间
        var waitStartTime int64
        //是否进入了饥饿模式
        starving := false
        //是否唤醒了当前的goroutine
        awoke := false
        //自旋次数
        iter := 0
        //当前状态
        old := m.state
        for {
            // 如果是饥饿情况，无需自旋
            // 如果其它g获取到了锁，则当前g尝试自旋获取锁            
            if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) { 
                //old>>mutexWaiterShift 锁上等待的goroutine数量
                //锁的状态设置为唤醒，这样当Unlock的时候就不会去唤醒其它被阻塞的goroutine了
                if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
                    atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
                    awoke = true
                }
                //自旋转30次数
                runtime_doSpin()
                //统计当前goroutine自旋次数
                iter++
                //更新锁的状态(有可能在自旋的这段时间之内锁的状态已经被其它goroutine改变)
                old = m.state
                continue
            }
            //自选完了还未获取到锁，则开始竞争锁

            //复制一份最新锁状态，用来存放期望的锁状态
            new := old

            // Don't try to acquire starving mutex, new arriving goroutines must queue.
            if old&mutexStarving == 0 {
                //非饥饿模式下，可以抢锁
                new |= mutexLocked
            }

            if old&(mutexLocked|mutexStarving) != 0 {
                //其它g已获取到锁或处于饥饿模式下，则会阻塞当前g，等待g的数量+1
                new += 1 << mutexWaiterShift
            }
            
            //当前goroutine的starving=true是饥饿状态，并且锁被其它goroutine获取了，
            // 那么将期望的锁的状态设置为饥饿状态           
            if starving && old&mutexLocked != 0 {
                new |= mutexStarving
            }
            //当前g被唤醒
            if awoke {               
                if new&mutexWoken == 0 {
                    throw("sync: inconsistent mutex state")
                }
                //new设置为非唤醒状态
                new &^= mutexWoken
            }
            // 通过CAS来尝试设置锁的状态
            if atomic.CompareAndSwapInt32(&m.state, old, new) {
                // 如果说old状态不是饥饿状态也不是被获取状态
                // 那么代表当前goroutine已经通过自旋成功获取了锁
                if old&(mutexLocked|mutexStarving) == 0 {
                    break // locked the mutex with CAS
                }               
                //是否放到队列的最前面，如果之前已经等待过直接放到队列最前面
                queueLifo := waitStartTime != 0
                //如果说之前没有等待过，就初始化设置现在的等待时间
                if waitStartTime == 0 {
                    //获取当前时间(单位ns)
                    waitStartTime = runtime_nanotime()
                }
                // 通过信号量来排队获取锁
                // 如果是新来的goroutine，就放到队列尾部
                // 如果是被唤醒的等待锁的goroutine，就放到队列头部
                runtime_SemacquireMutex(&m.sema, queueLifo, 1)  // runtime/sema.go

                //获取到信号量进入下面的步骤(唤醒了当前的waiter)

                //等待时间超过阀值进入饥饿模式
                starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs
                //获取锁的最新状态
                old = m.state
                // 如果说锁现在是饥饿状态，就代表现在锁是被释放的状态(unlock释放队列最前面的goroutine)，
                // 当前goroutine是被信号量所唤醒的
                // 也就是说，锁被直接交给了当前goroutine
                if old&mutexStarving != 0 {                   
                    //饥饿状态下不会有其它G获取到了锁或被唤醒
                    if old&(mutexLocked|mutexWoken) != 0 || old>>mutexWaiterShift == 0 {
                        throw("sync: inconsistent mutex state")
                    }
                    //当前goroutine已获取锁，则等待获取锁的goroutine数量-1
                    //最终状态atomic.AddInt32(&m.state, delta)
                    delta := int32(mutexLocked - 1<<mutexWaiterShift)
                    // 如果当前goroutine非饥饿状态，或者说当前goroutine是队列中最后一个goroutine
                    // 那么就退出饥饿模式，把状态设置为正常
                    if !starving || old>>mutexWaiterShift == 1 {                       
                        delta -= mutexStarving
                    }
                    //原子性地加上改动的状态
                    atomic.AddInt32(&m.state, delta)
                    break
                }
                // 如果锁不是饥饿模式，就把当前的goroutine设为被唤醒，由unlock操作导致的唤醒
                // 并且重置自旋计数器
                awoke = true
                iter = 0
            } else {
                //mutex状态已经被修改，刷新一遍重新计算
                old = m.state
            }
        }
    }

## Unlock

释放锁

    func (m *Mutex) Unlock() {
       
        // 这里获取到锁的状态，然后将状态减去被获取的状态(也就是解锁)，称为new(期望)状态
        // Fast path: drop lock bit.
        new := atomic.AddInt32(&m.state, -mutexLocked)
        //有其它g需要唤醒
        if new != 0 {           
            m.unlockSlow(new)
        }
    }

    func (m *Mutex) unlockSlow(new int32) {
        //unlock调用多次触发panic
        if (new+mutexLocked)&mutexLocked == 0 {
            throw("sync: unlock of unlocked mutex")
        }
        //非饥饿状态
        if new&mutexStarving == 0 {
            old := new
            for {
                // 如果说锁没有等待拿锁的goroutine
                // 或者锁被获取了(在循环的过程中被其它goroutine获取了)
                // 或者锁是被唤醒状态(表示有goroutine被唤醒，不需要再去尝试唤醒其它goroutine)
                // 或者锁是饥饿模式(会直接转交给队列头的goroutine)
                // 那么就直接返回             
                if old>>mutexWaiterShift == 0 || old&(mutexLocked|mutexWoken|mutexStarving) != 0 {
                    return
                }
                // 走到这一步的时候，说明锁目前还是空闲状态，并且没有goroutine被唤醒且队列中有goroutine等待拿锁
                // 那么我们就要把锁的状态设置为被唤醒，等待队列-1              
                new = (old - 1<<mutexWaiterShift) | mutexWoken
                if atomic.CompareAndSwapInt32(&m.state, old, new) {
                    //通过信号量去唤醒goroutine
                    runtime_Semrelease(&m.sema, false, 1)
                    return
                }
                old = m.state
            }
        } else {
            // 如果是饥饿状态下，那么我们就直接把锁的所有权通过信号量移交给队列头的goroutine就好了
            // handoff = true表示直接把锁交给队列头部的goroutine
            // 注意：在这个时候，锁被获取的状态没有被设置，会由被唤醒的goroutine在唤醒后设置
            // 但是当锁处于饥饿状态的时候，我们也认为锁是被获取的(因为我们手动指定了获取的goroutine)
            // 所以说新来的goroutine不会尝试去获取锁(在Lock中有体现)           
            runtime_Semrelease(&m.sema, true, 1)
        }
    }


