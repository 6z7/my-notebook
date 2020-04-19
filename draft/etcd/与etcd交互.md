etcdctl是一个与etcd服务端交互的命令行工具。与etcd交互的api可以使用version 2和version 3，通过`ETCDCTL_API`环境变量指定。默认情况下，3.4版本的etcdctl使用V3版本的API，之前的版本默认使用V2 API。

使用V2 API创建的key，不能被V3 API查询到。

`export ETCDCTL_API=3`
```
./etcdctl version
etcdctl version: 3.4.5
API version: 3.4


$ etcdctl put foo bar
OK

// 创建一个10秒过期的key, 1234abcd创建的10秒租约的命令返回的id
$ etcdctl put foo1 bar1 --lease=1234abcd
OK

//假设etcd存入了以下key
foo = bar
foo1 = bar1
foo2 = bar2
foo3 = bar3

$ etcdctl get foo
foo
bar

// 以十六进制输出
$ etcdctl get foo --hex
\x66\x6f\x6f          # Key
\x62\x61\x72          # Value


// 只打印值
$ etcdctl get foo --print-value-only
bar


// 查询[foor,foo3)范围内的key
$ etcdctl get foo foo3
foo
bar
foo1
bar1
foo2
bar2

// 查询以foo为前缀的key
$ etcdctl get --prefix foo
foo
bar
foo1
bar1
foo2
bar2
foo3
bar3

// 查询以foo为前缀的key并且只返回前2个
$ etcdctl get --prefix --limit=2 foo
foo
bar
foo1
bar1
```

## 访问旧版本的key

访问未来的版本号会报错
```
foo = bar         # revision = 2
foo1 = bar1       # revision = 3
foo = bar_new     # revision = 4
foo1 = bar1_new   # revision = 5

$ etcdctl get --prefix foo # access the most recent versions of keys
foo
bar_new
foo1
bar1_new

$ etcdctl get --prefix --rev=4 foo # access the versions of keys at revision 4
foo
bar_new
foo1
bar1

$ etcdctl get --prefix --rev=3 foo # access the versions of keys at revision 3
foo
bar
foo1
bar1

$ etcdctl get --prefix --rev=2 foo # access the versions of keys at revision 2
foo
bar

$ etcdctl get --prefix --rev=1 foo # access the versions of keys at revision 1
```

## 读取大于或等于指定键值的二进制值
```
$ etcdctl get --from-key b
b
456
z
789
```

## 删除key

```
$ etcdctl del foo
1 # one key is deleted

//删除foo-foo9范围内的
$ etcdctl del foo foo9
2 # two keys are deleted

//返回删除的kv
$ etcdctl del --prev-kv zoo
1   # one key is deleted
zoo # deleted key
val # the value of the deleted key

//按前缀删除
$ etcdctl del --prefix zoo
2 # two keys are deleted

//删除键值大于等于b的
$ etcdctl del --from-key b
2 # two keys are deleted
```

## 监听键的变换

```
//输出监听的键的变换情况
$ etcdctl watch foo
# in another terminal: etcdctl put foo bar
PUT
foo
bar

$ etcdctl watch foo --hex
# in another terminal: etcdctl put foo bar
PUT
\x66\x6f\x6f          # Key
\x62\x61\x72          # Value


//监听范围的键值[foo,foo9)
$ etcdctl watch foo foo9
# in another terminal: etcdctl put foo bar
PUT
foo
bar
# in another terminal: etcdctl put foo1 bar1
PUT
foo1
bar1

//监听foo开头的键
$ etcdctl watch --prefix foo
# in another terminal: etcdctl put foo bar
PUT
foo
bar
# in another terminal: etcdctl put fooz1 barz1
PUT
fooz1
barz1

//监听多个键
$ etcdctl watch -i
$ watch foo
$ watch zoo
# in another terminal: etcdctl put foo bar
PUT
foo
bar
# in another terminal: etcdctl put zoo val
PUT
zoo
val
```

## 监听键的历史变换

```
//返回自版本2以后的所有修改
$ etcdctl watch --rev=2 foo
PUT
foo
bar
PUT
foo
bar_new

//监视key的变换并返回最后一次修改之前的值
$ etcdctl watch --prev-kv foo
# in another terminal: etcdctl put foo bar_latest
PUT
foo         # key
bar_new     # last value of foo key before modification
foo         # key
bar_latest  # value of foo key after modification
```

## Watch progress

```
$ etcdctl watch -i
$ watch a
$ progress
progress notify: 1
# in another terminal: etcdctl put x 0
# in another terminal: etcdctl put y 1
$ progress
progress notify: 3
```

## 压缩key的历史记录

```
//压缩修订版小于5的历史记录
$ etcdctl compact 5
compacted revision 5

# any revisions before the compacted one are not accessible
$ etcdctl get --rev=4 foo
Error:  rpc error: code = 11 desc = etcdserver: mvcc: required revision has been compacted


//查看当前的修订版本
$ etcdctl get mykey -w=json
{"header":{"cluster_id":14841639068965178418,"member_id":10276657743932975437,"revision":15,"raft_term":4}}
```

## 租约
```
# grant a lease with 60 second TTL
//创建一个60秒的租约，当租约到时后，所有与该租约绑定的key都会被删除
$ etcdctl lease grant 60
lease 32695410dcc0ca06 granted with TTL(60s)

//通过id绑定指定的租约
# attach key foo to lease 32695410dcc0ca06
$ etcdctl put --lease=32695410dcc0ca06 foo bar
OK

//撤销租约，租约撤销时关联的key也被移除了
$ etcdctl lease revoke 32695410dcc0ca06
lease 32695410dcc0ca06 revoked

//保持租约不过期
$ etcdctl lease keep-alive 32695410dcc0ca06
lease 32695410dcc0ca06 keepalived with TTL(60)

//查看租约剩余存活时间
$ etcdctl lease timetolive 694d5765fc71500b
lease 694d5765fc71500b granted with TTL(500s), remaining(258s)

//查看租约剩余存活时间和与该租约绑定的key
$ etcdctl lease timetolive --keys 694d5765fc71500b
lease 694d5765fc71500b granted with TTL(500s), remaining(132s), attached keys([zoo2 zoo1])
```
