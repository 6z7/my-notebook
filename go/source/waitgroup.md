WaitGroup数据结构定义如下:

    type WaitGroup struct {
        //go vet验证是否copy
        noCopy noCopy

        //64位机器是内存地址是8字节对齐，32位机器是4字节对齐，32位的编译器不能保证8字节对齐
        //count|waiter|sema
        state1 [3]uint32
    }

## Add

    func (wg *WaitGroup) Add(delta int) {
        statep, semap := wg.state()
        if race.Enabled {
            _ = *statep // trigger nil deref early
            if delta < 0 {
                // Synchronize decrements with Wait.
                race.ReleaseMerge(unsafe.Pointer(wg))
            }
            race.Disable()
            defer race.Enable()
        }
        //count(8字节)|waiter(8字节)
        state := atomic.AddUint64(statep, uint64(delta)<<32)
        v := int32(state >> 32) //count的数量
        w := uint32(state)   //waiter的数量	
        //count<0异常
        if v < 0 {
            panic("sync: negative WaitGroup counter")
        }
        //waiter!=0  wait操作已经执行
        // delta > 0 说明是Add操作不是Done
        if w != 0 && delta > 0 && v == int32(delta) {
            //Add与Wait同时执行
            panic("sync: WaitGroup misuse: Add called concurrently with Wait")
        }
        //waiter=0说明没有g被挂起
        //count>0说明还可以挂起g
        if v > 0 || w == 0 {
            return
        }
        if *statep != state {
            panic("sync: WaitGroup misuse: Add called concurrently with Wait")
        }
        // Reset waiters count to 0.
        //waiter与count重置为0
        *statep = 0
        //唤醒挂起的g
        for ; w != 0; w-- {
            runtime_Semrelease(semap, false, 0)
        }
    }


    // 返回指向计数器的指针(8字节)与指向信号量的指针(4字节)
    //  count|waiter|sema
    // state returns pointers to the state and sema fields stored within wg.state1.
    func (wg *WaitGroup) state() (statep *uint64, semap *uint32) {
        // state1 [3]uint32  总共12个字节
        // 根据state1的起始地址分析,若是8字节对齐的,则直接用前8个字节作为*uint64类型
        // 若不是,说明是4字节对齐,则后移4个字节后,这样必为8字节对齐,然后取后面8个字节作为*uint64类型
        if uintptr(unsafe.Pointer(&wg.state1))%8 == 0 {
            return (*uint64)(unsafe.Pointer(&wg.state1)), &wg.state1[2]
        } else {
            return (*uint64)(unsafe.Pointer(&wg.state1[1])), &wg.state1[0]
        }
    }

## Done

    func (wg *WaitGroup) Done() {
        wg.Add(-1)
    }

## Wait

    func (wg *WaitGroup) Wait() {
        statep, semap := wg.state()
        for {
            //count|waiter
            state := atomic.LoadUint64(statep)
            v := int32(state >> 32) //count的数量
            w := uint32(state) //waiter的数量
            //count=0说明
            if v == 0 {			
                return
            }		
            //waiter数量+1
            if atomic.CompareAndSwapUint64(statep, state, state+1) {			
                //挂起当前g
                runtime_Semacquire(semap)
                if *statep != 0 {
                    panic("sync: WaitGroup is reused before previous Wait has returned")
                }			
                return
            }
        }
    }