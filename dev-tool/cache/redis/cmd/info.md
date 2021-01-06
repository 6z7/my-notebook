## info

> 从1.0.0版本

查询redis服务器统计信息

https://redis.io/commands/info

`Info [section]`

可选部分包括：

* server 服务器相关信息
* clients 客户端连接信息
* memory 
* persistence
* stats  统计信息
* replication 主从副本信息
* cpu
* commandstats  使用的命令统计信息
* cluster
* modules
* keyspace 存储的key统计信息
* errorstats

---

* all 返回所有的section，处了module
* defalut  仅仅返回默认的section
* everything 包括all和module