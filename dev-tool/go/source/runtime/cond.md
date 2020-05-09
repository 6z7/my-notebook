
Cond的实现依赖runtime/sema.go


    //wait先释放锁在将当前g放入等待队列,被唤醒后先获取锁
    type Cond struct {
        // noCopy可以嵌入到结构中，在第一次使用后不可复制,使用go vet作为检测使用
        noCopy noCopy

        // 根据需求初始化不同的锁，如*Mutex 和 *RWMutex       
        L Locker

        // 通知列表,调用Wait()方法的goroutine会被放入list中,每次唤醒,从这里取出
        notify notifyList
        // 复制检查,检查cond实例是否被复制
        //被复制后指向原始对象的指针,如果没有复制执行一个0的指针
        checker copyChecker
    }

## Wait

    func (c *Cond) Wait() {
        // 检查c是否是被复制的，如果是就panic
        c.checker.check()
        //获取当前g等待唤醒的编号
        t := runtime_notifyListAdd(&c.notify)
        //释放锁
        c.L.Unlock()
        //挂起当前g等待唤醒
        runtime_notifyListWait(&c.notify, t)
        c.L.Lock()
    }

## Signal

    func (c *Cond) Signal() {
        c.checker.check()
        //唤醒一个g
        runtime_notifyListNotifyOne(&c.notify)
    }

## Broadcast

    func (c *Cond) Broadcast() {
        c.checker.check()
        runtime_notifyListNotifyAll(&c.notify)
    }



## 检查对象是否被复制

    type copyChecker uintptr

    //如果没有被复制 c指向的是0的指针，否则指向被复制对象
    func (c *copyChecker) check() {
        if uintptr(*c) != uintptr(unsafe.Pointer(c)) &&
            !atomic.CompareAndSwapUintptr((*uintptr)(c), 0, uintptr(unsafe.Pointer(c))) &&
            uintptr(*c) != uintptr(unsafe.Pointer(c)) {
            panic("sync.Cond is copied")
        }
    }
