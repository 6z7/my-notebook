Map数据结构定义:

    type Map struct {
        mu Mutex

        // read部分内容，并发读无需锁
        // 如果key存在read中，则先尝试自旋更新value,如果失败在去获取mu	
        read atomic.Value // readOnly

        //ditry部分读写需要持有mu
        dirty map[interface{}]*entry

        // 统计 read中不存在需要访问ditry部分才能确定key是否存在的次数,
        // 一旦导致某一阈值(miss>=len(dirty))，dirty部分将会复制到read部分，并清空dirty部分
        misses int
    }

    type readOnly struct {
        m       map[interface{}]*entry
        //dirty部分包含read中不存在的key
        //每次dirty部分被复制到read部分时会重置为false
        amended bool
    }

## Load

    func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
        read, _ := m.read.Load().(readOnly)
        e, ok := read.m[key]
        //read部分不存在且dirty部分与read部分不相同
        if !ok && read.amended {
            m.mu.Lock()            
            //重新判断下，避免在获取mu时dirty部分复制到read部分
            read, _ = m.read.Load().(readOnly)
            e, ok = read.m[key]
            if !ok && read.amended {
                e, ok = m.dirty[key]               
                //无论dirty部分是否包含key，都统计miss,因为都穿透了read
                //如果miss>=len(m.dirty)则复制dirty部分到read
                m.missLocked()
            }
            m.mu.Unlock()
        }
        if !ok {
            return nil, false
        }
        //已删除或标记为删除的返回null
        return e.load()
    }

    func (e *entry) load() (value interface{}, ok bool) {
        p := atomic.LoadPointer(&e.p)
        if p == nil || p == expunged {
            return nil, false
        }
        return *(*interface{})(p), true
    }

    func (m *Map) missLocked() {
        m.misses++
        if m.misses < len(m.dirty) {
            return
        }
        m.read.Store(readOnly{m: m.dirty})
        m.dirty = nil
        m.misses = 0
    }

## Range

    func (m *Map) Range(f func(key, value interface{}) bool) {	
        read, _ := m.read.Load().(readOnly)
        if read.amended {		
            m.mu.Lock()
            read, _ = m.read.Load().(readOnly)
            if read.amended {
                read = readOnly{m: m.dirty}
                m.read.Store(read)
                m.dirty = nil
                m.misses = 0
            }
            m.mu.Unlock()
        }

        for k, e := range read.m {
            v, ok := e.load()
            if !ok {
                continue
            }
            if !f(k, v) {
                break
            }
        }
    }

## Store

    // 存储kv   
    func (m *Map) Store(key, value interface{}) {
        //获取read部分
        read, _ := m.read.Load().(readOnly)
        //read部分存在key,尝试自旋更新
        if e, ok := read.m[key]; ok && e.tryStore(&value) {
            //key存在，但是key对应的entry未被标记删除，直接保存到对应的entry上
            return
        }

        m.mu.Lock()
        //获取锁后重新获取一遍read
        read, _ = m.read.Load().(readOnly)
        //read中包含key
        if e, ok := read.m[key]; ok {
            // read中存在key，判断entry是否被标记删除(只有在dirty从read部分同步时才会标记)
            if e.unexpungeLocked() {               
                //read中存在key且被标记为删除，则dirty中一定不存在key
                m.dirty[key] = e
            }
            //entry没有被标记删除，直接保存值到read中
            e.storeLocked(&value)
        } else if e, ok := m.dirty[key]; ok {
             //read中不存在dirty中存在，这种情况发生在 保存一个新key的情况下，
             //如果load这个key或导致miss++
            e.storeLocked(&value)   //保存值dirty map
        } else {
            //dirty部分包含read中不存在的key
            //每次dirty部分被复制到read部分时会重置为false
            //第一次或复制后第一次则从read部分同步为被删除的key到dirty中
            if !read.amended {
                // We're adding the first new key to the dirty map.
                // Make sure it is allocated and mark the read-only map as incomplete.
                m.dirtyLocked()
                m.read.Store(readOnly{m: read.m, amended: true})
            }
            m.dirty[key] = newEntry(value)
        }
        m.mu.Unlock()
    }

## Delete

    func (m *Map) Delete(key interface{}) {
        read, _ := m.read.Load().(readOnly)
        e, ok := read.m[key]
        //read部分不存在key,dirty部分新增了key
        if !ok && read.amended {
            m.mu.Lock()
            read, _ = m.read.Load().(readOnly)
            e, ok = read.m[key]
            if !ok && read.amended {
                //read部分不存在 dirty部分直接删除
                delete(m.dirty, key)
            }
            m.mu.Unlock()
        }
        if ok {
            //将entry中的p标记为nil
            e.delete()
        }
    }

    //没有直接删除entry仅仅将p置为nil
    func (e *entry) delete() (hadValue bool) {
        for {
            p := atomic.LoadPointer(&e.p)
            if p == nil || p == expunged {
                return false
            }
            if atomic.CompareAndSwapPointer(&e.p, p, nil) {
                return true
            }
        }
    }