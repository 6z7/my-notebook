# Redis版本

3.0
---

主要变更：

* 集群
* ...

详细：https://github.com/redis/redis/blob/3.0.0/00-RELEASENOTES#L17

2.8迁移到3.0

详细：https://github.com/redis/redis/blob/3.0.0/00-RELEASENOTES#L578

4.0
---

主要变更：

* 新增Module系统
* 支持部分复制 PSYNC2
* 添加NAT/Docker支持
* ...

详细：https://github.com/redis/redis/blob/4.0.0/00-RELEASENOTES#L908

3.2迁移到4.0有部分是不兼容的:

* Reis集群协议不兼容
* RDB格式变换，4.0可以解析之前的版本，反之不行
* ...

详细：https://github.com/redis/redis/blob/4.0.0/00-RELEASENOTES#L3628

5.0
---

主要变更：

* 新增Stream数据类型
* 集群管理工具从Ruby(redis-trib.rb)移植到C(redis-cli，`redis-cli --cluster help` 查询更多信息)
* ...

详细：https://github.com/redis/redis/blob/5.0.0/00-RELEASENOTES#L2060

4.0迁移到5.0

详细：https://github.com/redis/redis/blob/5.0.0/00-RELEASENOTES#L2107


6.0
---

主要变更：

* 支持SSL
* 支持ACL
* RESP3
* 客户端缓存
* 多线程I/O
* 集群代理
* ...

详细：https://github.com/redis/redis/blob/6.0.0/00-RELEASENOTES#L1475

5.0迁移到6.0

详细：https://github.com/redis/redis/blob/6.0.0/00-RELEASENOTES#L1562