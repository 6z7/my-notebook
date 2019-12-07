    type Once struct {
        // 是否已执行       
        done uint32
        //互斥锁
        m Mutex
    }

    func (o *Once) Do(f func()) {	 
        if atomic.LoadUint32(&o.done) == 0 {           
            o.doSlow(f)
        }
    }

    func (o *Once) doSlow(f func()) {
        o.m.Lock()
        defer o.m.Unlock()
        if o.done == 0 {
            defer atomic.StoreUint32(&o.done, 1)
            f()
        }
    }
