# mutex

muxte是互斥锁的实现，用于不同协程间的同步。

g通过竞争获取mutex状态的修改所有权，所有其它g阻塞在信号量的队列上，除了唤醒的情况外。(如果当释放锁时，正好有g通过自旋使mutext的状态被标记了唤醒，此时先获取到锁的g会直接返回而不是阻塞到信号量队列上)。

当unlock后，排在队列头部的g会与新的g竞争所有权(即唤醒的g不一定能够获取到mutex的所有权)，当g被唤醒后如果等待时间大于1ms,则mutex状态会被标记为饥饿状态。

如果当前获取锁的g处于饥饿状态,则新的g不会自旋而是直接放入信号队列进行挂起,并将当前g放到队列的首位，下次唤醒直接执行该g。

互斥锁2种模式：正常模式，饥饿模式

正常模式下，g会按照FIFO顺序排队，唤醒时会与新的goroutine竞争mutex,新的goroutine因为已经在CPU上运行会比新唤醒的goroutine更有优势获取到mutex。如果g等待获取mutex超过1ms，则将该g将被放到队列的前面，同时锁状态切换到饥饿模式。

饥饿模式下，mutex的所有权直接从unlock goruntine交到队列头部的g。新的goroutine直接排到队列的尾部，不会尝试获mutex。如果g获取到mutex的后满足以下情况，则恢复到正常模式：  
	1.队列中最后一个g  
	2.获取Mutex的时间小于1ms

下要用到的数据结构与状态标识

```go    
    //互斥锁    
    type Mutex struct {
        //互斥锁状态
        //向右移动3位置表示等待获取锁的goroutine数量
        state int32
        //信号量
        sema uint32
    }

    const (
	//1：已获取到锁
	mutexLocked = 1 << iota // mutex is locked
	//2：已释放获取的锁
	mutexWoken
	//4：饥饿模式(排在前边的go协程一直未获取到锁)
	mutexStarving
	//3 表示mutex.state右移3位后即为等待的goroutine的数量
	mutexWaiterShift = iota	
	 
	//切换到饥饿模式的阀值1ms
	starvationThresholdNs = 1e6
    )

    //锁的接口定义
    type Locker interface {
	    Lock()
	    Unlock()
    }
 ```   

## Lock

获取锁:

1. 通过CAS尝试快速获取锁

2. 如果锁已经被其它g获取，则尝试自旋4次(饥饿状态就不用自旋了)，尝试将当前锁的状态标记为唤醒。

3. 根据当前锁的状态计算当前g获取锁时应该持有的状态
    - 如果目前锁处于非饥饿模式，则当前g可以去抢锁
    - 锁已经被获取，当前g只能等待，等待的g数量+1
    - 如果当前g挂起时间超过了阀值，则添加mutexStarving标记，这样下次可以优先被唤醒
    - 移除唤醒标记mutexWoken

4. CAS修改锁的状态为新计算出来的状态

5. 如果修改失败，从第2步重新开始

6. 当前g获取锁时state没有锁标记也没饥饿标记，那么通过CAS获取到锁后直接返回即可，

7. 根据当前g是否被挂起过，确定阻塞信号量队列的头还是尾

8. g被唤醒后，判断阻塞时间是否达到饥饿阀值。

9. 如果已经进入了饥饿模式，则判断这次被阻塞的时间是否小于饥饿阀值或是最后一个被阻塞的g，满足其中一个，移除饥饿标记，修改state，获取锁成功。否则从第2步重新竞争锁。


>唤醒状态是与锁的释放配合使用，可以使正在获取锁的某个g直接获取到锁。
>
>g通过自旋使当前锁的状态获取到了唤醒标记，如果此时进行锁的释放，看到了唤醒标记，直接返回即可。因为唤醒标记只在通过CAS获取到锁标记之前存在，释放锁时看到唤醒标记，说明获取锁的过程还在进行，第一个修改锁状态成功g，就直接获取到了锁，其它的则阻塞在信号量队列上。
>
>在通过CAS修改mutex的状态时，需要先移除唤醒标记(如果有)，否则在释放锁时，看到了此标记会直接返回，导致无法去唤醒阻塞在信号量队列上的g。唤醒标记只在获取到锁之前存在，使竞争获取锁标记的某个g直接获取到锁而不用阻塞在信号量队列上。

```go
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
                //自旋转30个时钟周期
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
```

## Unlock

释放锁:

1. 修改锁的状态移除mutexLocked标记，然后看看是否有其它g需要唤醒

2. 如果处于饥饿状态，则直接唤醒信号队列头部的g，此时state上的锁标记还在，被唤醒的g直接使用释放的g获取到锁。

3. 非饥饿模式下，如果state上有mutexLocked或mutexWoken或mutexStarving标记直接返回，否则通过CAS添加唤醒标记并将挂起g的数量-1，唤醒信号量队列上挂起的g。

>多次调用Unlock操作会panic

```go
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
```