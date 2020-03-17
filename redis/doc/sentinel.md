# Redis哨兵(Redis Sentinel)

Redis哨兵为Redis提供了高可用的保证。这意味着使用哨兵可以在没有人为干预的情况下进行redis部署用来抵抗某些类型的失败。

Redis哨兵也提供了一些辅助功能，如监视、通知和为客户端提供配置。

哨兵具有以下功能：

* 监视：  检查主节点和副本节点是否正常工作
* 通知：  通过API通知系统管理员或其它程序，被监视的节点有问题
* 自动故障转移：  如果主节点未如期工作，哨兵将会启动自动故障转移，将其中一个副本提升为主节点，其它副本被重新配置使用新的主节点，并且应用程序会被通知使用新的地址连接主节点
* 提供配置：  哨兵充当客户端服务发现的权威来源:客户端连接哨兵获取当前主节点的地址

## 哨兵的分布性

Redis哨兵是一个分布式系统:

哨兵被设计为使用相同的配置协调运行多个哨兵进程。多个哨兵进程协调运行具有以下优势:

1. 当多数哨兵同意主节点不可用时才执行故障转移，减低了误报的可能性
2. 即使不是所有的哨兵进程都在工作，哨兵也能工作，从而使系统能够抵抗故障。毕竟，拥有一个本身就是单点故障的故障转移系统是没有意义的。

# 快速入门

## 获得哨兵

当前版本的哨兵被称为Sentinel 2。它是对最初的Sentinel实现的重写，使用更强大和更简单的算法。

Redis Sentinel的稳定版本是从Redis 2.8开始发布的。

在不稳定的分支中执行新的开发，并且新的特性有时在被认为是稳定的时被重新移植到最新的稳定分支中。

Redis 2.6附带的Redis Sentinel版本1已弃用，不应使用。

## 运行哨兵
```
redis-sentinel /path/to/sentinel.conf
或
redis-server /path/to/sentinel.conf --sentinel
```

启动哨兵时必须使用配置文件，这个文件用来保存系统当前的状态，在重启时会直接加载该文件。如果没有配置文件或配置文件不可写，哨兵将拒绝启动。

默认情况下，哨兵通过`26379`端口进行通信，所以需要保证端口没有被使用，否则，哨兵之间无法进行交流，也就不能就如何操作达成一致，从而无法进行故障转移操作。

## 部署前需要了解的哨兵的基本知识

1. 至少需要三个哨兵节点
2. 分开独立部署
3. 哨兵+Redis构成的分布式系统，由于redis使用的是异步复制，所以不保证在失败转移期间已经写入的数据会保留
4. 客户端需要支持哨兵机制
5. ...
6. ...

## 配置哨兵

Redis源码中包含一个哨兵的示例配置`sentinel.conf`文件，最小配置如下:
```
sentinel monitor mymaster 127.0.0.1 6379 2
sentinel down-after-milliseconds mymaster 60000
sentinel failover-timeout mymaster 180000
sentinel parallel-syncs mymaster 1

sentinel monitor resque 192.168.1.3 6380 4
sentinel down-after-milliseconds resque 10000
sentinel failover-timeout resque 180000
sentinel parallel-syncs resque 5
```

仅仅需要指定需要监视的主节点，并指定一个唯一的名字。不需要指定副本，会自动发现。哨兵会自动更新副本的信息到配置文件中。在副本被提升为主节点或发现了新的哨兵时会重写配置文件。

`sentinel monitor <master-group-name> <ip> <port> <quorum>`

法定人数(quorum)的含义:

* 判定主节点不可达时，需要的哨兵数量
* quorum仅用于判定主节点不可用。为了实际执行故障转移，其中一个哨兵需要被选为故障转移的领导者并被授权继续。选举为领导者需要大多数哨兵同意。

举个例子，如果有5个哨兵，quorum等于2:
* 如果2个哨兵同时认为主节点不可用，两者之一将尝试进行故障转移
* 如果有至少3个节点同意，则故障转移被授权将会被执行

实际上，这意味着在发生故障时，如果大多数Sentinel无法进行对话，Sentinel将不会启动故障转移。

## 哨兵的其它配置

其它的一些配置格式:
`sentinel <option_name> <master_name> <option_value>`

常用配置:

* down-after-milliseconds：哨兵判定节点不可用的(不响应ping或返回错误)时间阀值，单位毫秒
* parallel-syncs：故障转移之后同时进行同步新的主节点的副本数量。值越小，故障转移完成耗时越长。虽然复制过程对于副本来说基本上是非阻塞的，但是有一段时间为了从主服务器加载大容量数据它会停止。所以越大就意味着越多的从服务器因为复制而不可用。可以通过将这个值设为1来保证每次只有一个从服务器处于不能处理命令请求的状态。

所有的配置参数在可以在运行时通过`SENTINEL SET`命令修改。

## 哨兵部署示例

...

## Sentinel、Docker、NAT和可能的问题

Docker使用一种端口映射的技术:docker容器中运行的程序可能暴露一个与它认为的不同端口。对于在同一个服务器上同时运行多个使用相同端口的程序的场景，这种方式是有用的。

并不是只有Docker才会产生这种情况，NAT也会造成端口映射，有时IP也会被重新映射。

端口和IP映射，在哨兵中会产生两种问题:

1. 哨兵自动发现其它哨兵的机制将不能正常工作，因为自动发现是基于每个哨兵发出的包含它监听的ip和端口的hello消息，然而哨兵不知道经过了地址或端口映射，所以它发出的对于其它哨兵是不正确的不能用于建立连接
2. 副本在master节点的info命令的输出中列出，地址通过tcp连接拿到的，但是端口是副本在握手时自己确定的，所以也存在端口映射的问题

由于哨兵自动发现副本是通过主节点的Info命令的输出信息发现的。由于发现的副本无法连接，哨兵将不能对主节点进行故障转移，因为从哨兵的角度来看是没有可用的副本还进行转移的。除非Docker使用端口1:1映射，不然哨兵无法正常工作。

对于第一个问题，如果你使用docker运行哨兵并进行了端口转发(后其它任何端口映射之类的操作)，可以通过两个配置强制哨兵使用特定的IP和端口进行声明:
```
sentinel announce-ip <ip>
sentinel announce-port <port>
```

需要注意的是，如果Docker运行时使用了host的网络模式(--network=host)则不会有问题。


# 快速入门教程

假设有3个哨兵实例，端口分别为5000、5001和5002，和一个端口为6379的主节点，主机节点的副本6380，ip地址使用127.0.0.1.

三个哨兵的配置文件类似这样:
```
port 5000
sentinel monitor mymaster 127.0.0.1 6379 2
sentinel down-after-milliseconds mymaster 5000
sentinel failover-timeout mymaster 60000
sentinel parallel-syncs mymaster 1
```
其它两个哨兵的配置使用端口5001和5002.

上面的配置文件中有些注意事项:

* mymater是主节点与它的副本的唯一标识。每个master和它的副本都有一个不同的名字，哨兵可以同时监视不同的master集合
* down-after-milliseconds设置为了5000毫秒，如果在5秒内没有收到主节点对ping的回复，主节点将会被认为主观下线 

一旦启动三个哨兵，将会看到如下日志:

`+monitor master mymaster 127.0.0.1 6379 quorum 2`

这是哨兵事件，如果订阅了相应的通道可以收到。

哨兵在故障检测和故障转移期间生成并记录不同的事件。

## 询问哨兵关于主节点的状态

```
$ redis-cli -p 5000
127.0.0.1:5000> sentinel master mymaster
 1) "name"
 2) "mymaster"
 3) "ip"
 4) "127.0.0.1"
 5) "port"
 6) "6379"
 7) "runid"
 8) "953ae6a589449c13ddefaee3538d356d287f509b"
 9) "flags"
10) "master"
11) "link-pending-commands"
12) "0"
13) "link-refcount"
14) "1"
15) "last-ping-sent"
16) "0"
17) "last-ok-ping-reply"
18) "735"
19) "last-ping-reply"
20) "735"
21) "down-after-milliseconds"
22) "5000"
23) "info-refresh"
24) "126"
25) "role-reported"
26) "master"
27) "role-reported-time"
28) "532439"
29) "config-epoch"
30) "1"
31) "num-slaves"
32) "1"
33) "num-other-sentinels"
34) "2"
35) "quorum"
36) "2"
37) "failover-timeout"
38) "60000"
39) "parallel-syncs"
40) "1"
```

如你所见，输出了大量信息，我们对其中的一些感兴趣:

1. num-other-sentinels等于2，所以我们知道哨兵已经发现了主机节点的其它2个哨兵。如果查看日志将能看到`+sentinel`事件生成
2. flags主节点的状态，master, s_down, o_down
3. num-slaves 副本的数量

```
// 查看监视的主节点的信息
sentinel master mymaster
// 查看所有副本的信息
SENTINEL slaves mymaster
// 查看所有哨兵的信息 不包括执行当前命令的哨兵
SENTINEL sentinels mymaster
```

## 获取主节点的地址

哨兵为客户端提供主节点和副本的地址配置信息，当进行了失败转移或重新配置，客户端通过api可以查询到新的主节点地址信息。
```
127.0.0.1:5000> SENTINEL get-master-addr-by-name mymaster
1) "127.0.0.1"
2) "6379"
```

## 测试失败转移

现在我们搭建的哨兵可以进行测试了。可以通过kill掉master，查看配置是否改变。也可以通过下面的命令实现：

`redis-cli -p 6379 DEBUG sleep 30`

这个命令会导致master不可达，睡眠30秒。模拟了master因为某些原因别挂起。

如果查看日志会看到:

1. 每个哨兵检测到master不可达，生成`+sdown`事件
2. 多数节点认为master不可达，升级为`+odown`
3. 哨兵投票选举一个leader开始第一次失败转移尝试
4. 失败转移

```
1458:X 17 Mar 2020 06:31:19.490 # +sdown master my-master 127.0.0.1 6379
1458:X 17 Mar 2020 06:31:19.550 # +odown master my-master 127.0.0.1 6379 #quorum 2/2
1458:X 17 Mar 2020 06:31:19.552 # +new-epoch 2
1458:X 17 Mar 2020 06:31:19.553 # +try-failover master my-master 127.0.0.1 6379
1458:X 17 Mar 2020 06:31:19.564 # WARNING: Sentinel was not able to save the new configuration on disk!!!: Permission denied
1458:X 17 Mar 2020 06:31:19.566 # +vote-for-leader 99484e2f62d8bcd516bc4e3050fdf0301d228e9a 2
1458:X 17 Mar 2020 06:31:19.605 # c3b7bebb0e7ed9ad3beb267e4333243304dd736b voted for 99484e2f62d8bcd516bc4e3050fdf0301d228e9a 2
1458:X 17 Mar 2020 06:31:19.608 # 7be0fdaa6a44a229d592535fe747284abe5999cb voted for 99484e2f62d8bcd516bc4e3050fdf0301d228e9a 2
1458:X 17 Mar 2020 06:31:19.621 # +elected-leader master my-master 127.0.0.1 6379
1458:X 17 Mar 2020 06:31:19.623 # +failover-state-select-slave master my-master 127.0.0.1 6379
1458:X 17 Mar 2020 06:31:19.677 # +selected-slave slave 127.0.0.1:6479 127.0.0.1 6479 @ my-master 127.0.0.1 6379
1458:X 17 Mar 2020 06:31:19.679 * +failover-state-send-slaveof-noone slave 127.0.0.1:6479 127.0.0.1 6479 @ my-master 127.0.0.1 6379
1458:X 17 Mar 2020 06:31:19.738 * +failover-state-wait-promotion slave 127.0.0.1:6479 127.0.0.1 6479 @ my-master 127.0.0.1 6379
1458:X 17 Mar 2020 06:31:19.776 # WARNING: Sentinel was not able to save the new configuration on disk!!!: Permission denied
1458:X 17 Mar 2020 06:31:19.778 # +promoted-slave slave 127.0.0.1:6479 127.0.0.1 6479 @ my-master 127.0.0.1 6379
1458:X 17 Mar 2020 06:31:19.780 # +failover-state-reconf-slaves master my-master 127.0.0.1 6379
1458:X 17 Mar 2020 06:31:19.819 # +failover-end master my-master 127.0.0.1 6379
1458:X 17 Mar 2020 06:31:19.821 # +switch-master my-master 127.0.0.1 6379 127.0.0.1 6479
1458:X 17 Mar 2020 06:31:19.823 * +slave slave 127.0.0.1:6379 127.0.0.1 6379 @ my-master 127.0.0.1 6479
1458:X 17 Mar 2020 06:31:19.832 # WARNING: Sentinel was not able to save the new configuration on disk!!!: Permission denied
1458:X 17 Mar 2020 06:31:49.872 # +sdown slave 127.0.0.1:6379 127.0.0.1 6379 @ my-master 127.0.0.1 6479
```

# Sentinel API

哨兵提供了检查哨兵状态、查看主节点与副本的健康状态、运行时修改配置以及订阅哨兵的通知的api

## 哨兵命令

* PING 
* SENTINEL masters：显示监视的主节点状态
* SENTINEL master \<master name>：显示指定主节点的状态
* SENTINEL slaves \<master name>：显示副本信息
* SENTINEL sentinels \<master name>：显示监视的哨兵
* SENTINEL get-master-addr-by-name \<master name>：查询主机点的ip和端口
* SENTINEL reset \<pattern>: 重置匹配的主节点的哨兵的所有状态
* SENTINEL failover \<master name>：强制进行故障转移，不要其它哨兵投票
* SENTINEL ckquorum \<master name>；检查是否满足故障转移需要的quorum
* SENTINEL flushconfig ：保存当前配置和状态到磁盘

从Redis2.8.4开始，哨兵提供了新增、修改、产出配置的api。如果有多个哨兵，需要每个哨兵都执行一遍相同的操作。

* SENTINEL MONITOR \<master name> \<ip> \<port> \<quorum> ：监控一个新的master,与sentinel.conf文件中的sentinel monitor完全相同，除了host只能使用ip地址外
* SENTINEL REMOVE \<master name> ：移除监控
* SENTINEL SET \<master name> \<option> \<value>  ：修改配置，sentinel.conf中的参数都可以通过该命令修改

## 新增或移除哨兵

新增一个哨兵到部署的哨兵集群中很容易，因为哨兵实现了自动发现机制。你只需要在启动哨兵时配置监视一个有效的master即可。10s内，新增的哨兵就能获得其它的哨兵和主节点的副本。

如果一次需要添加多个哨兵，建议一个接一个添加，等待所有其它哨兵都知道前一个添加的哨兵的情况后再添加下一个哨兵。

新增完成后可以使用` SENTINEL MASTER mastername`检查是否所有的哨兵都监视相同的主节点。

删除哨兵稍微有点复杂，因为哨兵不会忘记它看到过的哨兵，即使它已经不可达。删除哨兵需要以下步骤:

1. 停止哨兵进程
2. 发送`SENTINEL RESET *` 到所有的哨兵节点上(可以使用特定的master name替换\*)
3. 检查所有哨兵的`SENTINEL MASTER mastername`返回值是否一致

## 删除旧的主节点或不可达的副本

哨兵不会忘记主节点的副本，即使它们长时间不可达。这对于发生网络分区或失败后的，哨兵重新配置是有用的。而且，失败转移后，被转移的主节点会成为新的主节点的副本。通过在所有的哨兵上执行`SENTINEL RESET mastername`命令，可以刷新当前主节点的副本列表。

## 发布/订阅消息



