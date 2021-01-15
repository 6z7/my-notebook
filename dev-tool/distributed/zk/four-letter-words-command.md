# four letter words command

ZooKeeper3.4.6支持某些特定的四字命令字母与其的交互。它们大多是查询命令，用来获取 ZooKeeper 服务的当前状态及相关信息。用户在客户端可以通过 telnet 或 nc 向 ZooKeeper 提交相应的命令。

echo srvr | nc localhost 12081

srvr: 输出服务器的详细信息。zk版本、接收/发送包数量、连接数、模式（leader/follower）、节点总数。

ruok: 测试服务是否处于正确运行状态。如果正常返回"imok"，否则返回空。

conf: 服务配置的详细信息