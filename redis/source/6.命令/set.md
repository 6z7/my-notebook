
set的流程比较简单,在进行保存value前会先尝试压缩


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