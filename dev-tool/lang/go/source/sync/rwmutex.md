读写锁实现了Lock接口，数据结构定义如下:

    type RWMutex struct {
        //互斥锁
        w Mutex // held if there are pending writers
        //写锁信号量
        writerSem uint32 // semaphore for writers to wait for completing readers
        //读锁信号量
        readerSem uint32 // semaphore for readers to wait for completing writers
        //读锁计数器，写时会-rwmutexMaxReaders变成负数，告诉reader不能获取读锁
        readerCount int32 // number of pending readers
        //获取写锁时需要等待的读锁释放数量  如果>0会挂起当前g等待读锁释放后唤醒
        readerWait int32 // number of departing readers
    }

## Lock

获取写锁

    func (rw *RWMutex) Lock() {        
        //获取写锁,rwlock只能有一个g获取到wlock,所以在lock过程中不会释放锁，
        //cond允许多个g进行wait,所以在挂起g后释放了锁
        rw.w.Lock()

        // 将当前的readerCount置为负数，告诉RUnLock当前存在写锁等待
        // 先原子性减去在加回来  返回的是rlock数量
        r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
        // 等待读锁释放	
        if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
            //等待rlock释放才能写
            runtime_SemacquireMutex(&rw.writerSem, false, 0)
        }	
    }


## Unlock

释放写锁

    func (rw *RWMutex) Unlock() {  
        // 读锁计数器恢复，可以进行获取读锁了
        r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)
        // 没执行Lock调用Unlock，抛出异常
        if r >= rwmutexMaxReaders {           
            throw("sync: Unlock of unlocked RWMutex")
        }       
        // 唤醒挂起在readerSem上的g
        for i := 0; i < int(r); i++ {
            runtime_Semrelease(&rw.readerSem, false, 0)
        }       
        // 释放互斥锁允许其它g进行获取wlock
        rw.w.Unlock()      
    }

## RLock

获取读锁

    func (rw *RWMutex) RLock() {       
        //读锁计数器<0说明其它g获取到了wlock
        if atomic.AddInt32(&rw.readerCount, 1) < 0 {
            // A writer is pending, wait for it.
            //挂起当前g
            runtime_SemacquireMutex(&rw.readerSem, false, 0)
        }       
    }

## RUnlock

释放读锁

    func (rw *RWMutex) RUnlock() {       
        //读锁数量<0说明有g获取到了wlock
        if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {          
            // rlock释放还有wlock的g，则去尝试唤醒
            rw.rUnlockSlow(r)
        }
        if race.Enabled {
            race.Enable()
        }
    }

    func (rw *RWMutex) rUnlockSlow(r int32) {
        // r + 1 == 0表示直接执行RUnlock()
        // r + 1 == -rwmutexMaxReaders表示执行Lock()再执行RUnlock()
        if r+1 == 0 || r+1 == -rwmutexMaxReaders {           
            throw("sync: RUnlock of unlocked RWMutex")
        }      
        //读锁释放，wlock等待的rlock-1
        if atomic.AddInt32(&rw.readerWait, -1) == 0 {
            // The last reader unblocks the writer.
            // 所有rlock都释放，唤醒挂起的执行wlock的g
            runtime_Semrelease(&rw.writerSem, false, 1)
        }
    }