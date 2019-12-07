sync.pool数据结构定义:

    type Pool struct {
        noCopy noCopy

        //指向poolLocal[cpu核数]数组
        local     unsafe.Pointer // local fixed-size per-P pool, actual type is [P]poolLocal
        //poolLocal[cpu核数]数组的大小 等于cpu的核心数
        localSize uintptr        // size of the local array

        //上次回收的local
        victim     unsafe.Pointer // local from previous cycle
        //上次回收的localSize
        victimSize uintptr        // size of victims array

        //创建新的对象方法
        New func() interface{}
    }

    // Local per-P Pool appendix.
    type poolLocalInternal struct {
        //当前P持有的一个缓存的值，用于快速访问
        private interface{} // Can be used only by the respective P.
        //缓存要保存的值
        shared  poolChain   // Local P can pushHead/popHead; any P can popTail.
    }

    type poolLocal struct {
        poolLocalInternal

        // Prevents false sharing on widespread platforms with
        // 128 mod (cache line size) = 0 .
        pad [128 - unsafe.Sizeof(poolLocalInternal{})%128]byte
    }

## Put

    // 保存x到缓存中
    // 优先保存在当前P的私有变量处，在保存在队列中  
    func (p *Pool) Put(x interface{}) {
        if x == nil {
            return
        }
        
        //获取当前p对应的poolLocal
        l, _ := p.pin()
        //当前P的私有变量还未被使用
        if l.private == nil {
            //保存一个要缓存的值
            l.private = x
            x = nil
        }
        //如果当前P的私有变量处无法保存则保存在队列中
        if x != nil {
            l.shared.pushHead(x)
        }
        runtime_procUnpin()
        if race.Enabled {
            race.Enable()
        }
    }

    func (p *Pool) pin() (*poolLocal, int) {
        //m.lock++,禁止抢占
        pid := runtime_procPin()
        // In pinSlow we store to local and then to localSize, here we load in opposite order.
        // Since we've disabled preemption, GC cannot happen in between.
        // Thus here we must observe local at least as large localSize.
        // We can observe a newer/larger local, it is fine (we must observe its zero-initialized-ness).
        //当前pool数组大小
        s := atomic.LoadUintptr(&p.localSize) // load-acquire
        //pool数组
        l := p.local                          // load-consume
        if uintptr(pid) < s {
            //pool数组中第pid个元素
            return indexLocal(l, pid), pid
        }
        //如果P的id大于poolLocal数组的大小，则重新分配数组
        return p.pinSlow()
    }

    //返回当前P关联的poolLocal
    func (p *Pool) pinSlow() (*poolLocal, int) {
        // Retry under the mutex.
        // Can not lock the mutex while pinned.
        runtime_procUnpin()
        allPoolsMu.Lock()
        defer allPoolsMu.Unlock()
        pid := runtime_procPin()
        // poolCleanup won't be called while we are pinned.
        s := p.localSize
        l := p.local
        if uintptr(pid) < s {
            return indexLocal(l, pid), pid
        }
        if p.local == nil {
            allPools = append(allPools, p)
        }
        // If GOMAXPROCS changes between GCs, we re-allocate the array and lose the old one.
        // cpu核数
        size := runtime.GOMAXPROCS(0)
        local := make([]poolLocal, size)
        atomic.StorePointer(&p.local, unsafe.Pointer(&local[0])) // store-release
        atomic.StoreUintptr(&p.localSize, uintptr(size))         // store-release
        return &local[pid], pid
    }

## Get

    func (p *Pool) Get() interface{} {
        if race.Enabled {
            race.Disable()
        }
        //当前P所属的poolLocal
        l, pid := p.pin()
        //当前P的私有值
        x := l.private
        //清空，用于下次保存私有值
        l.private = nil
        //当前P的私有值没有，则从队列中获取
        if x == nil {
            // 先尝试从head中pop,再从其它P中偷
            // Try to pop the head of the local shard. We prefer
            // the head over the tail for temporal locality of
            // reuse.
            x, _ = l.shared.popHead()
            if x == nil {
                x = p.getSlow(pid)
            }
        }
        //m.lock-- 可以被抢占
        runtime_procUnpin()
        if race.Enabled {
            race.Enable()
            if x != nil {
                race.Acquire(poolRaceAddr(x))
            }
        }
        if x == nil && p.New != nil {
            x = p.New()
        }
        return x
    }

    //从其它P的缓冲中偷
    func (p *Pool) getSlow(pid int) interface{} {
        // See the comment in pin regarding ordering of the loads.
        size := atomic.LoadUintptr(&p.localSize) // load-acquire
        locals := p.local                        // load-consume
        // Try to steal one element from other procs.
        // 尝试从其它P的缓冲中偷
        for i := 0; i < int(size); i++ {
            l := indexLocal(locals, (pid+i+1)%int(size))
            if x, _ := l.shared.popTail(); x != nil {
                return x
            }
        }

        // 尝试从上次gc后的缓存中获取
        // Try the victim cache. We do this after attempting to steal
        // from all primary caches because we want objects in the
        // victim cache to age out if at all possible.
        size = atomic.LoadUintptr(&p.victimSize)
        if uintptr(pid) >= size {
            return nil
        }
        locals = p.victim
        l := indexLocal(locals, pid)
        if x := l.private; x != nil {
            l.private = nil
            return x
        }
        for i := 0; i < int(size); i++ {
            l := indexLocal(locals, (pid+i)%int(size))
            if x, _ := l.shared.popTail(); x != nil {
                return x
            }
        }

        // Mark the victim cache as empty for future gets don't bother
        // with it.
        atomic.StoreUintptr(&p.victimSize, 0)

        return nil
    }

## init

gc时触发poolCleanup

    func init() {
        //gc cleanup
        runtime_registerPoolCleanup(poolCleanup)
    }

    //stw时调用
    //gc时将allPool移动到oldPools,并清空当前缓存，下次gc时清空oldPools，在重新赋值未allPool,
    //所以sync.Pool不适合缓存socket连接
    func poolCleanup() {
        // This function is called with the world stopped, at the beginning of a garbage collection.
        // It must not allocate and probably should not call any runtime functions.

        // Because the world is stopped, no pool user can be in a
        // pinned section (in effect, this has all Ps pinned).

        // Drop victim caches from all pools.
        for _, p := range oldPools {
            p.victim = nil
            p.victimSize = 0
        }

        // Move primary cache to victim cache.
        for _, p := range allPools {
            p.victim = p.local
            p.victimSize = p.localSize
            p.local = nil
            p.localSize = 0
        }

        // The pools with non-empty primary caches now have non-empty
        // victim caches and no pools have primary caches.
        oldPools, allPools = allPools, nil
    }
