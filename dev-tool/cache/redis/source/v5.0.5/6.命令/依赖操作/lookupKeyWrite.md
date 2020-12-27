## lookupKeyWrite

从数据中查找指定key,查找前先删除过期的key

    robj *lookupKeyWrite(redisDb *db, robj *key) {
        //释放过期的key
        expireIfNeeded(db,key);
        return lookupKey(db,key,LOOKUP_NONE);
    }

### expireIfNeeded    

删除过期key

    //key有效返回0，过期返回1
    //1.判断key是否过期
    //2.如果是主节点 广播过期key的del操作到AOF file与从节点
    //3.发送事件通知
    //4.同步(异步)释放kv
    int expireIfNeeded(redisDb *db, robj *key) {
        //检查key是否过期
        if (!keyIsExpired(db,key)) return 0;

        /* If we are running in the context of a slave, instead of
        * evicting the expired key from the database, we return ASAP:
        * the slave key expiration is controlled by the master that will
        * send us synthesized DEL operations for expired keys.
        *
        * Still we try to return the right information to the caller,
        * that is, 0 if we think the key should be still valid, 1 if
        * we think the key is expired at this time. */
        //从节点 会通过同步主节点的del操作来删除过期key
        if (server.masterhost != NULL) return 1;

        /* Delete the key */
        server.stat_expiredkeys++;
        //主节点广播过期key的del操作到AOF file与从节点
        propagateExpire(db,key,server.lazyfree_lazy_expire);
        //发送事件
        notifyKeyspaceEvent(NOTIFY_EXPIRED,
            "expired",key,db->id);
        //释放kv
        return server.lazyfree_lazy_expire ? dbAsyncDelete(db,key) :
                                            dbSyncDelete(db,key);
    }

### dbAsyncDelete

异步释放kv

    //异步释放kv
    int dbAsyncDelete(redisDb *db, robj *key) {
        /* Deleting an entry from the expires dict will not free the sds of
        * the key, because it is shared with the main dictionary. */
        if (dictSize(db->expires) > 0) dictDelete(db->expires,key->ptr);

        /* If the value is composed of a few allocations, to free in a lazy way
        * is actually just slower... So under a certain limit we just free
        * the object synchronously. */
        //从字典中删除key未释放节点
        dictEntry *de = dictUnlink(db->dict,key->ptr);
        if (de) {
            robj *val = dictGetVal(de);
            size_t free_effort = lazyfreeGetFreeEffort(val);

            /* If releasing the object is too much work, do it in the background
            * by adding the object to the lazy free list.
            * Note that if the object is shared, to reclaim it now it is not
            * possible. This rarely happens, however sometimes the implementation
            * of parts of the Redis core may call incrRefCount() to protect
            * objects, and then call dbDelete(). In this case we'll fall
            * through and reach the dictFreeUnlinkedEntry() call, that will be
            * equivalent to just calling decrRefCount(). */
            //如果释放的对象的大小大于阈值,则创建一个后台job进行释放
            if (free_effort > LAZYFREE_THRESHOLD && val->refcount == 1) {
                atomicIncr(lazyfree_objects,1);
                //创建一个job释放内存
                bioCreateBackgroundJob(BIO_LAZY_FREE,val,NULL,NULL);
                dictSetVal(db->dict,de,NULL);
            }
        }

        /* Release the key-val pair, or just the key if we set the val
        * field to NULL in order to lazy free it later. */
        if (de) {
            dictFreeUnlinkedEntry(db->dict,de);
            if (server.cluster_enabled) slotToKeyDel(key);
            return 1;
        } else {
            return 0;
        }
    } 

 ### dbSyncDelete 

 同步删除kv

    /* Delete a key, value, and associated expiration entry if any, from the DB */
    int dbSyncDelete(redisDb *db, robj *key) {
        /* Deleting an entry from the expires dict will not free the sds of
        * the key, because it is shared with the main dictionary. */
        if (dictSize(db->expires) > 0) dictDelete(db->expires,key->ptr);
        if (dictDelete(db->dict,key->ptr) == DICT_OK) {
            if (server.cluster_enabled) slotToKeyDel(key);
            return 1;
        } else {
            return 0;
        }
    } 


### lookupKey

从db中查找key,更新lfu/lru相关信息

    robj *lookupKey(redisDb *db, robj *key, int flags) {
        dictEntry *de = dictFind(db->dict,key->ptr);
        if (de) {
            robj *val = dictGetVal(de);

            /* Update the access time for the ageing algorithm.
            * Don't do it if we have a saving child, as this will trigger
            * a copy on write madness. */
            if (server.rdb_child_pid == -1 &&
                server.aof_child_pid == -1 &&
                !(flags & LOOKUP_NOTOUCH))
            {
                //LFU策略
                if (server.maxmemory_policy & MAXMEMORY_FLAG_LFU) {
                    updateLFU(val);
                } else {
                    val->lru = LRU_CLOCK();
                }
            }
            return val;
        } else {
            return NULL;
        }
    }