
set的流程比较简单,在进行保存value前会先尝试压缩后在保存


    /* SET key value [NX] [XX] [EX <seconds>] [PX <milliseconds>] */
    void setCommand(client *c) {
        int j;
        robj *expire = NULL;
        int unit = UNIT_SECONDS;
        int flags = OBJ_SET_NO_FLAGS;

        //解析命令参数
        for (j = 3; j < c->argc; j++) {
            char *a = c->argv[j]->ptr;
            robj *next = (j == c->argc-1) ? NULL : c->argv[j+1];

            if ((a[0] == 'n' || a[0] == 'N') &&
                (a[1] == 'x' || a[1] == 'X') && a[2] == '\0' &&
                !(flags & OBJ_SET_XX))
            {
                flags |= OBJ_SET_NX;
            } else if ((a[0] == 'x' || a[0] == 'X') &&
                    (a[1] == 'x' || a[1] == 'X') && a[2] == '\0' &&
                    !(flags & OBJ_SET_NX))
            {
                flags |= OBJ_SET_XX;
            } else if ((a[0] == 'e' || a[0] == 'E') &&
                    (a[1] == 'x' || a[1] == 'X') && a[2] == '\0' &&
                    !(flags & OBJ_SET_PX) && next)
            {
                flags |= OBJ_SET_EX;
                unit = UNIT_SECONDS;
                expire = next;
                j++;
            } else if ((a[0] == 'p' || a[0] == 'P') &&
                    (a[1] == 'x' || a[1] == 'X') && a[2] == '\0' &&
                    !(flags & OBJ_SET_EX) && next)
            {
                flags |= OBJ_SET_PX;
                unit = UNIT_MILLISECONDS;
                expire = next;
                j++;
            } else {
                addReply(c,shared.syntaxerr);
                return;
            }
        }

        //尝试压缩
        c->argv[2] = tryObjectEncoding(c->argv[2]);
        setGenericCommand(c,flags,c->argv[1],c->argv[2],expire,unit,NULL,NULL);
    }


    void setGenericCommand(client *c, int flags, robj *key, robj *val, robj *expire, int unit, robj *ok_reply, robj *abort_reply) {
        long long milliseconds = 0; /* initialized to avoid any harmness warning */

        if (expire) {
            //获取过期时间
            if (getLongLongFromObjectOrReply(c, expire, &milliseconds, NULL) != C_OK)
                return;
            if (milliseconds <= 0) {
                addReplyErrorFormat(c,"invalid expire time in %s",c->cmd->name);
                return;
            }
            if (unit == UNIT_SECONDS) milliseconds *= 1000;
        }
        //查找key前会进行释放过期key的操作
        if ((flags & OBJ_SET_NX && lookupKeyWrite(c->db,key) != NULL) ||
            (flags & OBJ_SET_XX && lookupKeyWrite(c->db,key) == NULL))
        {
            addReply(c, abort_reply ? abort_reply : shared.nullbulk);
            return;
        }
        setKey(c->db,key,val);
        server.dirty++;
        if (expire) setExpire(c,c->db,key,mstime()+milliseconds);
        //事件通知
        notifyKeyspaceEvent(NOTIFY_STRING,"set",key,c->db->id);
        if (expire) notifyKeyspaceEvent(NOTIFY_GENERIC,
            "expire",key,c->db->id);
        //写响应数据到缓冲中
        addReply(c, ok_reply ? ok_reply : shared.ok);
    }

## tryObjectEncoding

    //尝试编码 为了压缩空间
    //1. 要是 OBJ_ENCODING_RAW或OBJ_ENCODING_EMBSTR编码方式
    //2. 没有被共享
    //3. 能转为long类型 则使用共享对象
    //4. 非OBJ_ENCODING_EMBSTR编码则尝试转为OBJ_ENCODING_EMBSTR编码
    //5. OBJ_ENCODING_RAW编码方式 尝试释放多余的空闲空间
    robj *tryObjectEncoding(robj *o) {
        long value;
        sds s = o->ptr;
        size_t len;

        /* Make sure this is a string object, the only type we encode
        * in this function. Other types use encoded memory efficient
        * representations but are handled by the commands implementing
        * the type. */
        serverAssertWithInfo(NULL,o,o->type == OBJ_STRING);

        /* We try some specialized encoding only for objects that are
        * RAW or EMBSTR encoded, in other words objects that are still
        * in represented by an actually array of chars. */
        //只处理RAW or EMBSTR编码的数据，应为只有这些数据还是原始的字符数组
        if (!sdsEncodedObject(o)) return o;

        /* It's not safe to encode shared objects: shared objects can be shared
        * everywhere in the "object space" of Redis and may end in places where
        * they are not handled. We handle them only as values in the keyspace. */
        //共享对象
        if (o->refcount > 1) return o;

        /* Check if we can represent this string as a long integer.
        * Note that we are sure that a string larger than 20 chars is not
        * representable as a 32 nor 64 bit integer. */
        len = sdslen(s);
        //是否能转为long类型数字
        if (len <= 20 && string2l(s,len,&value)) {
            /* This object is encodable as a long. Try to use a shared object.
            * Note that we avoid using shared integers when maxmemory is used
            * because every object needs to have a private LRU field for the LRU
            * algorithm to work well. */
            //(没有最大内存限制||LRU算法不会回收共享对象)&& value>=0&&value<=10000
            if ((server.maxmemory == 0 ||
                !(server.maxmemory_policy & MAXMEMORY_FLAG_NO_SHARED_INTEGERS)) &&
                value >= 0 &&
                value < OBJ_SHARED_INTEGERS)
            {
                decrRefCount(o);
                incrRefCount(shared.integers[value]);
                return shared.integers[value];
            } else {
                if (o->encoding == OBJ_ENCODING_RAW) sdsfree(o->ptr);
                o->encoding = OBJ_ENCODING_INT;
                o->ptr = (void*) value;
                return o;
            }
        }

        /* If the string is small and is still RAW encoded,
        * try the EMBSTR encoding which is more efficient.
        * In this representation the object and the SDS string are allocated
        * in the same chunk of memory to save space and cache misses. */
        //是否能使用OBJ_ENCODING_EMBSTR类型编码
        if (len <= OBJ_ENCODING_EMBSTR_SIZE_LIMIT) {
            robj *emb;

            if (o->encoding == OBJ_ENCODING_EMBSTR) return o;
            emb = createEmbeddedStringObject(s,sdslen(s));
            //ref count减去1
            decrRefCount(o);
            return emb;
        }

        /* We can't encode the object...
        *
        * Do the last try, and at least optimize the SDS string inside
        * the string object to require little space, in case there
        * is more than 10% of free space at the end of the SDS string.
        *
        * We do that only for relatively large strings as this branch
        * is only entered if the length of the string is greater than
        * OBJ_ENCODING_EMBSTR_SIZE_LIMIT. */
        //使用OBJ_ENCODING_RAW编码的方式的字符，如果剩余空间大于已经使用空间的10%,则释放剩余的空间
        trimStringObjectIfNeeded(o);

        /* Return the original object. */
        return o;
    }