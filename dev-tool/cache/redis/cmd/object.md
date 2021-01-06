## object

> 从2.2.3版本

查看key在redis内部的使用的object信息

https://redis.io/commands/object

`object subcommand [arguments [arguments ...]]`

支持的子命令:

* refcount key对应的值引用次数
* encoding 编码方式
* idletime key空闲时间，单位秒
* freq key访问频率
* help 帮助信息

对象编码方式:

* string：raw、int(64位有符号整数)
* list:ziplist(小list使用)、linkedlist
* set:intset(整数构成的小集合)、hashtable
* hash:ziplist(小hash)、hashtable
* sorted set:ziplist(小有序集合)、skiplist