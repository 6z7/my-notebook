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

客户端订阅哨兵上的通道获取事件通知。        

通道的名称和事件名称一致。如名字为+sdown的通道将收到所有实例进入SDOWN状态(从哨兵角度看主机节点不可达)的通知。

下面是你可以使用api收到的所有通道和消息格式。第一个是通道/事件名称，后边是数据格式。

`<instance-type> <name> <ip> <port> @ <master-name> <master-ip> <master-port>`

@后边的是master信息，这部分是可选的，只有当实例不是主节点时才出现。

* +reset-master \<instance details> -- 主节点被重置.

* +slave \<instance details> -- 新的副本加入.

* +failover-state-reconf-slaves \<instance details> -- Failover state changed to reconf-slaves state.

* +failover-detected \<instance details> -- A failover started by another Sentinel or any other external entity was detected (An attached replica turned into a master).

* +slave-reconf-sent \<instance details> -- The leader sentinel sent the SLAVEOF command to this instance in order to reconfigure it for the new replica.

* +slave-reconf-inprog \<instance details> -- The replica being reconfigured showed to be a replica of the new master ip:port pair, but the synchronization process is not yet complete.

* +slave-reconf-done \<instance details> -- The replica is now synchronized with the new master.

* -dup-sentinel \<instance details> -- One or more sentinels for the specified master were removed as duplicated (this happens for instance when a Sentinel instance is restarted).

* +sentinel \<instance details> -- 新的哨兵加入

* +sdown \<instance details> -- 主观下线状态

* -sdown \<instance details> -- 退出主观下线状态

* +odown \<instance details> -- 客观下线状态

* -odown \<instance details> -- 退出客观下线装填

* +new-epoch \<instance details> -- epoch被更新

* +try-failover \<instance details> -- 哨兵准备进行failover,等待被选举为leader

* +elected-leader \<instance details> -- 哨兵赢得当前的epoch阶段的leader，可以进行failover

* +failover-state-select-slave \<instance details> -- New failover state is select-slave: we are trying to find a suitable replica for promotion.

* no-good-slave \<instance details> -- There is no good replica to promote. Currently we'll try after some time, but probably this will change and the state machine will abort the failover at all in this case.

* selected-slave \<instance details> -- We found the specified good replica to promote.

* failover-state-send-slaveof-noone \<instance details> -- We are trying to reconfigure the promoted replica as master, waiting for it to switch.

* failover-end-for-timeout \<instance details> -- The failover terminated for timeout, replicas will eventually be configured to replicate with the new master anyway.

* failover-end \<instance details> -- The failover terminated with success. All the replicas appears to be reconfigured to replicate with the new master.

* switch-master \<master name> \<oldip> \<oldport> \<newip> \<newport> -- 主机节点新的ip和端口

* +tilt -- Tilt mode entered.

* -tilt -- Tilt mode exited.

## -BUSY状态处理

当Lua脚本执行时间超过配置的限制时，redis讲返回-BUSY错误。当进行failover之前发生了这种错误，哨兵将先发送`script kill`命令，但是只有在脚本是只读的情况下才会成功。如果在尝试之后，redis实例还是处于-BUSY状态，那么failover将失败。            

## 副本优先级

Redis实例有一个配置`replica-priority`，在info命令的输出可以看到，哨兵将使用这个参数排序挑选出一个副本升级为主节点。

1. 如果副本优先级设置为0，则该副本不会被提升为主节点
2. 值越小优先级越高

## 哨兵和Redis认证

为了安全，master配置了客户端连接需要密码，那么副本需要为了与主节点连接需要知道密码。

这是通过以下配置文件中的指令实现的:

* requirepass  设置主节点的密码
* masterauth   副本配置主节点需要的密码

由于使用了哨兵，master节点不在固定。由于副本可以升级为master,master可以配置为新的master的副本，所以，如果要认证，需要在所有实例上都配置以上2个指令。

哨兵连接需要认证的节点时，需要配置以下指令:  
`sentinel auth-pass <master-group-name> <pass>`

## 配置哨兵需要认证

从Redis 5.0.1可以配置哨兵需要密码认证

`requirepass "your_password_here"`

所有的哨兵需要使用相同的密码，同时连接哨兵的客户端需要支持向哨兵发送`AUTH`命令

## 哨兵客户端实现

哨兵客户端的实现参见： [Sentinel clients guidelines](https://redis.io/topics/sentinel-clients)

# 更深入的概念

下面的章节中，我们将讨论集群是如何工作的。

## SODWN和ODOWN故障状态

Redis哨兵有两个不同的概念，主观下线(Subjectively Down,SDOWN)和客观下线(Objectively Down,ODOWN)。主观下线时哨兵从自己的视角观察到主节点在有限的时间内不可达，客观下线是配置的quorum个哨兵认为主节点不在有限的时间内不可达。

哨兵在配置的(is-master-down-after-milliseconds)有限时间内没有收到PING的回复，会认为主节点主观下线。

可以接受的PING的响应有:

* +PONG
* -LOADING error
* -MASTERDOWN error

除此之外，其它的回复或没有回复则认为无效。

SDOWN转为ODOWN的过程没有使用强一致性算法，而是使用了gossip:哨兵在有限的时间内收到了足够多的其它哨兵的报告主节点不可达，则SDOWN被提升为ODOWN。

ODOWN只针对master节点，副本节点只有SDOWN状态。

处于SDOWN状态的副本不会被提升为主节点

## 哨兵与副本的自动发现

哨兵与其它哨兵保持连接，以便相互检查对方的可用性，并交换消息。不需要在每个哨兵上配置其它哨兵的地址，哨兵使用Redis的发布/订阅功能来发现其它监视同一个master的哨兵和副本。

这个功能的实现是通过发送`hello messages`到`__sentinel__:hello`通道实现的。

类似的，也不需要配置连接主节点的副本，哨兵通过查询Redis自动发现。

* 每个哨兵发布消息到其监视的master和副本上的` __sentinel__:hello`通道，每隔2秒发送一次，消息内容包含哨兵的ip、端口、runid

* 每个哨兵订阅master和副本上的`__sentinel__:hello`通道，查找自己还不知道的哨兵。当检测到新的哨兵，则加入到监视当前master的哨兵集群

* Hello消息中包含主节点的当前配置信息，如果哨兵上的关于主节点的配置比收到的配置信息旧，则更新哨兵上的配置

* 添加新的哨兵监视主节点之前，哨兵会检查是否有其它哨兵与其拥有相同的runid或地址(ip和端口)。如果有相同的哨兵则先移除，在添加新的。

## 故障转移过程之外哨兵的重新配置

即使没有failover在处理，哨兵也会一直尝试在被监视的节点上进行以下配置:

* 副本声称是master(根据它自己的配置),将会被配置为当前master的副本

* 副本连接了错误的master,将会被重新配置为正确的master的副本

要让哨兵重新配置副本，必须在一段时间内观察到错误的配置，这段时间要比广播新配置的时间长。

关于本节要记住的重要一课是：Sentinel是一个系统，在这个系统中，每个进程总是试图将最后一个逻辑配置强加于被监视的实例集。

## 副本选择与优先级

当master处于ODOWN状态，哨兵收到大多数其它已知的哨兵授权，准备准备进行failover，需要选择一个合适的副本。

副本的选择，依据以下规则：
1. 与mater主从复制断开的时间
2. 副本的优先级
3. 副本的复制偏移
4. RunID

副本与主节点的断开时间(info输出)大于
`(down-after-milliseconds * 10) + milliseconds_since_master_is_in_SDOWN_state`，那么该副本将会被排除。  
down-after-milliseconds：哨兵判断主节点主观下线的时间  
milliseconds_since_master_is_in_SDOWN_state：哨兵观察到的主节点已处于SDOWN状态的时间

副本选择只考虑通过上述条件的副本，并根据上述条件按以下顺序对其排序。

1. 副本的redis.conf文件中配置的replica-priority，值越小优先级越高
2. 如果优先级相同，选择副本复制偏移量大的节点
3. 如果优先级和复制偏移量都相同，则字典顺序选择RunID较小的。选择的RunID小的节点不一定是最合适的，但是可以使副本的选择过程更具有确定性

Redis主服务器(在故障转移后可以转换为副本)和副本，如果有强优先权的机器，则都必须配置`replica-priority`。

`replica-priority`如果配置为0，那么该节点将不会被哨兵提升为新的master，但是仍然会被哨兵配置为新的master的副本。

# 内部算法

在下面的章节中，我们将探讨哨兵行为的细节。用户并不需要知道所有的细节，但是对哨兵的深入了解有助于更有效地部署和操作哨兵。

## Quorum

每个哨兵监控的master都需要配置一个quorum。它指定了需要多少个哨兵判定master不可达才触发failover。

failover触发之后，为了执行failover流程，需要至少大多数哨兵授权某个哨兵执行哨兵。在少数哨兵的分区中不会执行failover。

* Quorum：需要多少个哨兵检测到master节点不可达，将其标记为ODOWN
* 进入ODWN状态后触发failover
* 一旦触发failover，哨兵需要获得大多数哨兵(或者超过大多数，如果quorum设置的大于大多数)的授权后尝试进行failover

这种差异看起来很微妙，但实际上很容易理解和使用。例如，若果有5个哨兵，quorum设置了为2，只要2个哨兵认为master不可达就会触发failover，但是只有两个哨兵中一个获得了至少3个哨兵的授权后才能进行failover。

如果quorum配置了5，必须所有的哨兵都认为master不可达，同时需要所有节点的授权才能进行failover。

这意味着quorum可以通过两种方式调整哨兵：

1. 如果quorum的值小于部署的哨兵的大多数，我们让哨兵对master的失败更敏感，只要有少数哨兵无法与master正常通信，就会触发failover
2. 如果quorum的值大于部署的哨兵的大多数，则仅当存在大量(大于大多数)的哨兵认为master不可达时才会触发failover

## epochs配置

哨兵需要获得多数哨兵的授权才能启动故障转移，有几个重要的原因：

当一个哨兵被授权，为将要被进行failover的master获得了一个唯一的配置epoch。这个数据是failover完成之后新配置的版本。因为大多数哨兵都同意这个指定的版本分配给指定的哨兵，所以其它哨兵不能在获得这个版本。这意味着每次failover的配置都有一个唯一的版本。

哨兵有一个规则：如果一个哨兵投票给另一个哨兵进行指定master的failover，那么该哨兵将等待一些时间再次对同一个master进行failove(即,如果failover还没完成，哨兵投票给了其它哨兵并没有投给自己，那么该哨兵会等待一段时间后才会再次发起failover)。这个延迟等待的时间配置通过`sentinel.conf`文件中的`failover-timeout`进行配置。这意味着哨兵不会尝试同时对相同的master进行failover，第一个哨兵将尝试获得授权，如果失败等待一段时间后另一个哨兵开始尝试，以此类推。

哨兵提供了liveness保证，即大多数哨兵能够正常通信，那么，当master下线时，其中一个哨兵将会被授权进行failover。

Redis哨兵提供了安全保证，每个哨兵在对相同的master进行filover时,使用不同的配置epoch。

## 配置传播

一旦哨兵failover成功，它将广播新的配置以便其它哨兵能更新关于master的信息。

要使failover被认为是成功的，它要求Sentinel能够向所选副本发送`SLAVEOF NO ONE`命令，并且随后在master的`info`输出中观察到切换到master。

此时，即使副本的重新配置还在进行中，也会认为故障转移成功，并且所有哨兵都需要开始报告新配置。

每个哨兵连续不断的通过Pub/Sub消息向所有的master和副本广播它的版本的关于master的配置。同时所有的哨兵等待消息，看看其它哨兵的配置是什么(也会接受到自己发送的消息)

配置信息通过`__sentinel__:hello `Pub/Sub通道广播。

因为每个配置有一个不同的版本号，版本号大的将胜过版本号小的。

举个例子，所有的哨兵的配置开始时都认为mymaster主节点在192.168.1.50:6379。此时这个配置的版本号是1。一段时间后某个哨兵被授权使用version 2进行了failover。如果failover成功，该哨兵将广播新的配置，比如说，version 2中主机节点在192.168.1.50:9000,由于具有更高的version，其它哨兵看到这个配置后将会更新它们自己的配置。

## 分区下的一致性

Redis哨兵的会最终保持一致。所以每个分区都会收敛到可用的更高配置。然而，在使用哨兵的真实系统中，会有三种不同的角色：

* Redis实例
* 哨兵实例
* 客户端

为了定义系统的行为，我们必须考虑这三个因素。

下面是一个3个节点组成的简单网络，每个节点都运行一个Redis实例和哨兵实例。

```
           +-------------+
            | Sentinel 1  |----- Client A
            | Redis 1 (M) |
            +-------------+
                    |
                    |
+-------------+     |          +------------+
| Sentinel 2  |-----+-- // ----| Sentinel 3 |----- Client B
| Redis 2 (S) |                | Redis 3 (M)|
+-------------+                +------------+
```

在这个系统中初始状态，Redis 3是master,Redis 1和Rdis 2是副本。发生分区隔离了master。哨兵1和2启动failover提升Redis 1成为新的master。

哨兵的机制保证，哨兵1和2拥有master的最新配置，由于发生了分区，哨兵3仍然是旧的配置。

我们知道在分区恢复后，哨兵2将更新它的配置，但是在分区时，如果客户端和旧的master分区在一起，会发生什么？

客户端仍然可以写数据到Redis 3，当分区恢复，Redis 3将成为Redis 1的副本，所有在分区期间的写数据将会丢失。

根据配置，您可能希望或不希望发生以下情况：

* 如果使用Redis作为cache，客户端B仍然写数据到旧的master，即使数据丢失也没什么问题
* 如果使用Redis作为存储，可能不需要发生这种情况，需要通过配置，以部分防止这种问题

由于Redis是异步复制，在这种场景下没有办法完全防止数据丢失，但是在Redis 3和Redis 1上使用下面的配置减少这种问题的影响

```
// 如果master的有效副本少于指定值，则将拒绝写操作
min-replicas-to-write 1
// master收到的副本的ping的最大延迟N秒，超过将拒绝写操作
min-replicas-max-lag 10
```

使用了以上配置，Redis 3在10秒后将不用，当分区恢复后，哨兵3的配置将更新，客户端B将能获取到新的有效配置继续运行。

## 哨兵状态持久化

哨兵状态被持久化到哨兵的配置文件中。如，每次收到新的配置，配置会和配置epoch一块被持久化。这意味着停止和重启哨兵是安全的。

## TILT模式

Redis哨兵在很大程度上都依赖计算机时间：例如，为了了解一个实例是否可用，它会记住对PING命令的最新成功应答时间，并将其与当前时间进行比较，以了解它有多旧。

如果计算机时间被意外改变，或者如果计算机很忙，或者进程由于某种原因被阻塞，Sentinel可能会工作不正常。

当检测到可能降低系统可靠性的异常情况时，TILT模式是一种哨兵可进入的特别的保护模式。哨兵的定时任务每秒调用10次，因此，我们预计在两次调用之间大约会经过100毫秒。

哨兵会记录上定时任务被调用的时间，与当前时间比较：如果差值是负值或大于2秒，则进入TILT模式。

当哨兵处于TILT模式时，仍会继续监视，但是：
1. 停止所有活动
2. 对`SENTINEL is-master-down-by-addr`的请求，不在回复有效的信息

若果30秒内恢复正常，则退出TILT模式。
