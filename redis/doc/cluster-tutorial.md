[Redis集群教程](https://redis.io/topics/cluster-tutorial)
---
&emsp;&emsp;本教程是对redis集群的简单介绍，没有使用复杂的概念来理解分布式系统。该文档是关于如何配置、测试、操作集群，深入的相关细节需要阅读[Redis Cluster specification]()。

&emsp;&emsp;本教程尝试从用户角度以简单易懂的方式介绍关于Redis集群的可用性和一致性相关信息。

&emsp;&emsp;需要注意的是教程中使用的是Redis 3.0或以上的版本。

&emsp;&emsp;如果你计划部署一台严谨的Redis集群系统，建议要阅读下Redis Cluster specification,即使部署要求没那么要个也建议你阅读下。先从该文档开始熟悉下redis集群之后在在阅读集群规范也是一个不错的方式。

### Redis集群概述

&emsp;&emsp;Redis集群提供一种运行redis实例的方式，其中数据自动在多个redis节点上进行分片。

&emsp;&emsp;Redis集群还在分区期间提供了一定程度的可用性，实际上就是在某些节点失败或无法通信时能够继续进行服务的能力。但是如果集群发生较大故障将停止运转(如集群中大多数主节点不可用)。

&emsp;&emsp;使用Redis集群实际能获得什么呢?

* 在多个节点间进行数据分片的能力
* 少数节点发生故障或不能与集群其它节点通信时能够继续对外提供服务的能力

### Redis集群TCP端口

&emsp;&emsp;每个Redis集群节点都需要开放2个tcp端口，一个用于与客户端通信的常规端口如6379，另一个在这个端口上加10000用于集群节点间的数据通信如16379。

&emsp;&emsp;第二个高端口用于节点间使用二进制协议进行通信,称为集群总线Cluster bus,Cluster bus可用于节点间的故障检测、配置更新、故障转移授权等等操作。客户端不应该尝试与Cluster bus端口进行通信，请确保在防火墙中打开了者两个端口，否则节点间将无法通信。

&emsp;&emsp;命令端口与cluster bus端口相差是固定的必须是10000。

&emsp;&emsp;为了使Redis集群正常工作，对于每个节点都需要满足以下条件:

1. 节点需要打开命令端口(通常是6379)，能够与客户端进行通信
2. 节点间能够通过Cluster bus端口(命令端口+10000)进行通信

&emsp;&emsp;cluster bus使用二进制协议进行节点间的数据交换，用于节约带宽和处理时间

### Redis集群与Docker

### Redis集群数据分片

&emsp;&emsp;Redis集群没有使用一致性hash进行数据分片，而是使用了一种称为hash slot(hash槽)的分片方式，每个key逻辑上属于某个slot。

&emsp;&emsp;Redis集群有16384个hash slot，计算key属于哪个slot使用crc16(key)模16384的方式。

&emsp;&emsp;每个节点负责一部分hash slot，举个例子如果有3个节点则:

* nodeA负责0-5500
* nodeB负责5501-11000
* nodeC负责11001-16383

&emsp;&emsp;集群中可以轻松的添加或删除节点。如果想增加一个新节点D，那需要从节点A、B、C移动一些slot到节点D。类似的如果想删除一个节点A,需要将节点A上的slot移动到B和C。当节点上的slot全部移走后就可以从集群中移除该节点。

&emsp;&emsp;因为从一个节点移动hash slot到另一个节点不需要停止服务，因此新增或移除节点或改变节点持有的slot的比例也不会影响服务。

&emsp;&emsp;Redis集群支持多key操作原子执行，但需要所涉及的key属于同一个hash slot。也可以使用hash tag的方式强制多key分配到同一个slot。

&emsp;&emsp;hash tag在Redis Cluster specification中有描述，要点是key字符中有一个大括号{},仅对大括号中的内容进行hash。举个例子this{foo}key和another{foo}key会被分配到同一个hash slot。

### Redis集群主从模式

&emsp;&emsp;当集群中的部分主节点出现故障或不能与集群中的大多数节点通信时，为了提高可用性，Redis集群使用主从模式，每个hash slot有1个(主节点)到N个副本(N-1个从节点)。

&emsp;&emsp;在我们作为举例的集群中有个A、B、C三个节点，如果B节点失败则集群将不能继续工作，因为我们不能提供范围5501-11000的槽了。

&emsp;&emsp;如果我们在创建时或之后我们为每个master节点添加一个slave节点，此时集群由A、B、C作为master，A1、B1、C1作为salve节点组成，如果B节点失败集群仍能够继续正常运转。

&emsp;&emsp;节点B1是B的副本，如果B节点失败，集群将提升节点B1成为新的master，集群将继续正常运行。

&emsp;&emsp;如果节点B和B1同时失败，集群将不能继续工作。

### Redis集群一致性保证

&emsp;&emsp;Redis集群不保证强一致性，因此在某些情况下集群会丢失一些数据。

&emsp;&emsp;Redis集群会丢失数据的第一个原因是由于异步复制。在写数据期间会发生如下操作:

* client写数据到master B
* master B回复OK给client
* master B将写入的数据发送到它的从节点B1、B2、B3

&emsp;&emsp;如上所示，master B在回复client之前不会等待B1、B2、B3的确认，因此如果master在发送数据到slave之前发生crash，slave升级为master时将丢失数据。
&emsp;&emsp;这种情况类似每秒刷新数据到磁盘，为了提高一致性我们可以在返回客户端之前强制刷新数据到磁盘，但会影响性能，同理集群采取同步复制的方式也一样。

&emsp;&emsp;通常在性能与一致性之间需要平衡。

&emsp;&emsp;Redis集群支持同步写,通过WAIT命令实现,这可以降低数据丢失的概率，但是尽管使用了同步复制集群也没有实现强一致性，在复杂的故障场景下还是可能有没有接收到master节点数据的从节点被选举为master。

&emsp;&emsp;这里有一种集群丢失数据的情况，当发生网络分区时，一个client与至少包含一个master节点的少数节点被隔离。

&emsp;&emsp;我们以有6个节点A、B、C、A1、B1、C1的集群举例说明，其中三个主节点三个从节点，还有一个客户端Z1。

&emsp;&emsp;当网络分区发生后，分区的一侧可能有A、C、A1、B1、C1，另一侧有B和Z1。

&emsp;&emsp;此时Z1依然能写数据到B，B也会接受Z1的命令。如果分区在很短的时间内恢复，集群将能够继续正常工作。然而如果分区持续了一段时间，此期间B1在拥有大多数节点的一侧被选举为master，那么Z1发送到B的数据将会丢失。

&emsp;&emsp;需要注意的是，Z1能够发送数据到B有一个最大窗口期限制。当拥有大多数节点的一侧选举一个从节点成为新的主节点后，在拥有少数节点的另一侧里的每个master将停止接收客户端的命令。

&emsp;&emsp;这个最大窗口期对Redis集群是一个很重要的配置被称为node timeout。

&emsp;&emsp;当节点与集群失去联系超过node timeout后，master节点将会被标记为fail，会被它的副本中的某一个替换，同时这个与集群失去联系的master将停止接收客户端的命令。


## Redis集群配置参数

&emsp;&emsp;我们将部署一个集群事例。在这之前先了解下redis.conf文件中关于集群的相关配置。

* cluster-enabled \<yes/no> 是否开启Redis集群功能
* cluster-config-file \<filename> 节点的配置文件，这个不需要用户编辑，记录一些集群的配置，方便节点重启时使用，当集群的信息发生变换时会刷新到该文件
* cluster-node-timeout \<milliseconds>

















