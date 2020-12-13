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

* cluster-node-timeout \<milliseconds> 在被标记为fail之前最大可以持续的不可用时间

* cluster-slave-validity-factor \<factor> 如果设置为0，则不管主从断开多久，从节点总是试图对master进行failover。如果是一个大于0的值则允许的主从最大断开时间为cluster-node-timeout+cluster-slave-validity-factor，如果主从断开超过这个阈值则从节点不会对master进行failover。如果master发生故障时没有从节点进行failover，则集群将不可用。

* cluster-migration-barrier \<count> 将master的从节点漂移到其它master时，当前master需要保留的最少从节点数量。

* cluster-require-full-coverage \<yes/no> 如果设置为yes,它也是默认值，如果部分key所在的节点不可用时整个集群将停止工作，反之slot还能继续提供服务。

* cluster-allow-reads-when-down \<yes/no> 如果设置为no,它也是默认值，当集群被标记为fail时集群将停止服务，反之将允许读。

## 搭建一个Redis集群

下面是搭建集群的最小配置

    port 7000
    cluster-enabled yes
    cluster-config-file nodes.conf
    cluster-node-timeout 5000
    appendonly yes

搭建集群最少需要3个master节点，对于当一次搭建建议启动6个节点3个master3个salve。

就这样做，我们先进入一个新目录，创建如下的目录，目录名为节点端口号。

    mkdir cluster-test
    cd cluster-test
    mkdir 7000 7001 7002 7003 7004 7005


在每个目录下创建redis.conf文件，按照上边提到的最小配置进行配置，注意端口要进行相应修改。在每个目录下执行类似下边的操作启动redis:

    cd 7000
    ../redis-server ./redis.conf

启动后将看到如下日志

    [82462] 26 Nov 11:56:55.329 * No cluster configuration found, I'm 97a3a64667477371c4479320d683e4c8db5858b1

这个ID将作为该reids实例在集群中的唯一名称，集群中的每个节点通过这个ID记住对方而不是通过ip或端口，ip和端口可能改变，但是节点的声明周期中这个ID不会改变，我们称之为Node ID。

### 创建一个集群

&emsp;&emsp;现在我们已经运行起来了几个redis实例，我们还需要向redis实例写入一些配置来创建一个集群出来。

&emsp;&emsp;如果你使用的是Redis 5，那么通过redis-cli中内置的集群帮助工具很容易实现，通过这个工具可以实现创建一个集群，对存在的集群进行检测或重新分片等操作。

&emsp;&emsp;对Redis 3或4，有个类似的工具叫redis-trib.rb。你可以在源码目录里找到，运行之前需要先安装redis gem

    gem install redis


&emsp;&emsp;第一个例子我将使用redis-cli和redis-trib创建集群。之后的例子将只使用redis-cli。需要注意的是完全可以使用Redis5的redis-cli工具操作Redis4集群。

使用Redis5的redis-cli创建集群:

    redis-cli --cluster create 127.0.0.1:7000 127.0.0.1:7001 \
    127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 \
    --cluster-replicas 1

使用redis-trib.rb创建集群:

    ./redis-trib.rb create --replicas 1 127.0.0.1:7000 127.0.0.1:7001 \
    127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005

create命令意味着是创建一个新的集群，--cluster-replicas 1选项代表每个master创建一个slave，其它参数是用于创建集群的redis地址

显然上边的操作是创建了一个3主3从的集群

Redis-cli将会给出一个建议配置，输入yes代表接受，之后集群将会被配置和连接，相互之间会进行通信。最后如果一切顺利将看到如下输出:

    [OK] All 16384 slots covered

### 使用create-cluster脚本创建集群

&emsp;&emsp;如果你不想向上边一样通过手动配置和运行各个实例来创建集群，还有一种更简单的方式。

在源码的utils/create-cluster目录中有一个名为create-cluster的bash脚本。为了启动由3个主节点3个从节点共6个节点组成的集群，只需要输入以下命令:

    1.create-cluster start
    2.create-cluster create

在执行第二步骤时需要输入yes同意redis-clie工具的集群分配方案。

现在你可以与集群进行交互了，第一个节点启动的默认端口是30001，输入以下命令可以停止集群:

    create-cluster stop

关于如何使用这个脚步的详细信息可以阅读脚本所在目录中的README文件。

### 测试集群

&emsp;&emsp;下面是一个使用redis-cli与redis集群交互的事例:

    $ redis-cli -c -p 7000
    redis 127.0.0.1:7000> set foo bar
    -> Redirected to slot [12182] located at 127.0.0.1:7002
    OK
    redis 127.0.0.1:7002> set hello world
    -> Redirected to slot [866] located at 127.0.0.1:7000
    OK
    redis 127.0.0.1:7000> get foo
    -> Redirected to slot [12182] located at 127.0.0.1:7002
    "bar"
    redis 127.0.0.1:7000> get hello
    -> Redirected to slot [866] located at 127.0.0.1:7000
    "world"

redis-cli对集群的支持非常基础，Redis集群需要通过客户端将请求重定向正确的节点。一个严谨的客户端能够做的更好，它能够缓存hash slot与节点地址间的映射关系，因此能够直接从定向正确的节点。当建群的配置发生变换，如发生了failover或管理员新增或移除了节点，此时需要刷新节点的映射关系。

### 集群重新分片

&emsp;&emsp;重新分片是移动某些节点上的slot到其它节点的上，使用redis-cli工具启动resharding如下:

    redis-cli --cluster reshard 127.0.0.1:7000

只需要指定一个节点，redis-clie将自动发现其它节点。

当前redis-cli只能在管理员的支持下才能进行resharding，我们不能指定说只移动5%的slot到其它节点，命令执行后会出现一个问题:

    How many slots do you want to move (from 1 to 16384)?

我们可以尝试移动1000个slot。

redis-cli需要知道resharding的目标，这里使用第一个master节点，即127.0.0.1:7000，但是需要指定该实例的Node ID。使用如下命令可以找到某个节点的ID:

    $ redis-cli -p 7000 cluster nodes | grep myself
    97a3a64667477371c4479320d683e4c8db5858b1 :0 myself,master - 0 0 0 connected 0-5460

 resharding结束后，可以使用如下命令测试集群是否健康:

    redis-cli --cluster check 127.0.0.1:7000

### 脚本化重新分片操作

&emsp;&emsp;resharding可以自动化进行，不用手动交互在命令行中输入参数。使用像下面的命令:

    redis-cli reshard <host>:<port> --cluster-from <node-id> --cluster-to <node-id> --cluster-slots <number of slots> --cluster-yes

当前redis还不支持根据节点上key的分布自动进行在平衡，这个功能在未来会加入。

### 更多有意思的事例

参见 [redis-rb-cluster](https://github.com/antirez/redis-rb-cluster)。

其中的consistency-test.rb是一个简单的一致性检测脚本，它能告诉你是否集群丢失了数据或没有收到集群的响应。

    $ ruby consistency-test.rb
    925 R (0 err) | 925 W (0 err) |
    5030 R (0 err) | 5030 W (0 err) |
    9261 R (0 err) | 9261 W (0 err) |
    13517 R (0 err) | 13517 W (0 err) |
    17780 R (0 err) | 17780 W (0 err) |
    22025 R (0 err) | 22025 W (0 err) |
    25818 R (0 err) | 25818 W (0 err) |

## 测试故障转移

&emsp;&emsp;在测试期间，最好运行起一致性检测脚本consistency-test.rb方便观察。

为了触发failover，最简单的方式是使进程crash，在我们的例子中就是使某个节点crash。

我们先通过下边的命令找到一个master节点:

    $ redis-cli -p 7000 cluster nodes | grep master
    3e3a6cb0d9a9a87168e266b0a0b24026c0aae3f0 127.0.0.1:7001 master - 0 1385482984082 0 connected 5960-10921
    2938205e12de373867bf38f1ca29d31d0ddb3e46 127.0.0.1:7002 master - 0 1385482983582 0 connected 11423-16383
    97a3a64667477371c4479320d683e4c8db5858b1 :0 myself,master - 0 0 0 connected 0-5959 10922-11422

可以看到7000、70001和70005，现在使用DEBUG SEGFAULT命令将7002节点：

    $ redis-cli -p 7002 debug segfault
    Error: Server closed the connection

现在在一致性脚本运行的窗口可以看到如下输出:

    18849 R (0 err) | 18849 W (0 err) |
    23151 R (0 err) | 23151 W (0 err) |
    27302 R (0 err) | 27302 W (0 err) |

    ... many error warnings here ...

    29659 R (578 err) | 29660 W (577 err) |
    33749 R (578 err) | 33750 W (577 err) |
    37918 R (578 err) | 37919 W (577 err) |
    42077 R (578 err) | 42078 W (577 err) 

通过输出结果发现集群不能进行读写，因而没有不一致的数据产生。这听起来可能有些出乎意料，因为在本教程的第一部分中，我们提到了Redis集群在故障转移期间会丢失写操作，因为它使用异步复制，我们没有说的是这是不太可能发生的,因为Redis发送回复到客户端，和命令复制到slave大约在同一时间，所以有一个很小的丢失数据的窗口。虽然概率很小但是也能会发生。现在我们看下failover后的集群状况(已经手动重启了crash的节点,重新加入了集群变成了从节点)。

    $ redis-cli -p 7000 cluster nodes
    3fc783611028b1707fd65345e763befb36454d73 127.0.0.1:7004 slave 3e3a6cb0d9a9a87168e266b0a0b24026c0aae3f0 0 1385503418521 0 connected
    a211e242fc6b22a9427fed61285e85892fa04e08 127.0.0.1:7003 slave 97a3a64667477371c4479320d683e4c8db5858b1 0 1385503419023 0 connected
    97a3a64667477371c4479320d683e4c8db5858b1 :0 myself,master - 0 0 0 connected 0-5959 10922-11422
    3c3a0c74aae0b56170ccb03a76b60cfe7dc1912e 127.0.0.1:7005 master - 0 1385503419023 3 connected 11423-16383
    3e3a6cb0d9a9a87168e266b0a0b24026c0aae3f0 127.0.0.1:7001 master - 0 1385503417005 0 connected 5960-10921
    2938205e12de373867bf38f1ca29d31d0ddb3e46 127.0.0.1:7002 slave 3c3a0c74aae0b56170ccb03a76b60cfe7dc1912e 0 1385503418016 3 connected

现在主节点运行在7000、7001和7005端口，之前的master现在运行在7002端口上，现在成为了7005的从节点。

CLUSTER NODES命令的输出格式:

* NODE ID
* ip:port
* falgs: master, slave, myself, fail, ...
* 如果是从节点则是主节点的NODE ID
* 最近一个在等待回复的ping时间
* 最近一个收到的PONG时间
* 这个节点的config epoch
* 连接到节点的连接状态
* 负责的slot

## 手动故障转移

&emsp;&emsp;有时进行手动failover是需要的，如升级集群中的master节点，先把它转变成从节点在进行升级，这样对集群的影响就很小了。

通过CLUSTER FAILOVER命令实现手动failover，这个命令必须是在你想要进行failover的master节点下的其中一个slave节点上进行的。

手动进行failover与实际发生的failover相比是特殊和安全的。手动failver在新的master从旧的master复制完数据后才会通知client切换到新的master。

下面是进行手动failover时产生的日志:

    # Manual failover user request accepted.
    # Received replication offset for paused master manual failover: 347540
    # All master replication stream processed, manual failover can start.
    # Start of election delayed for 0 milliseconds (rank #0, offset 347540).
    # Starting a failover election for epoch 7545.
    # Failover election won: I'm the new master.

连接到进行failover的master节点上的连接会被暂停，同时master发送它的偏移信息到slave，等待slave复制数据完成后，q启动进行failover，旧的master会被通知进行配置切换。当旧的master上的被暂停的连接放行后，将会重定向新的master。

## 新增节点

&emsp;&emsp;添加新节点基本上就是添加一个空节点，然后将一些数据移动到其中的过程，如果设置了成为某个主节点的副本则成为一个slave节点，否则则是master节点。

两种情况都会演示，先看下成为一个master的情况。

按照之前的手动启动一个节点的步骤启动一个端口为7006的节点。使用rdis-cli将节点加入集群:

    redis-cli --cluster add-node 127.0.0.1:7006 127.0.0.1:7000

第一个参数是新节点的地址，第二个参数是随机选择的一个已知节点地址。

实际上redis-cli在这里并没有做什么，仅是发送CLUSTER MEET消息到节点，操作之前会先检查集群的状态。

查询集群的节点信息可以看到新节点已经加入集群:

    redis 127.0.0.1:7006> cluster nodes
    3e3a6cb0d9a9a87168e266b0a0b24026c0aae3f0 127.0.0.1:7001 master - 0 1385543178575 0 connected 5960-10921
    3fc783611028b1707fd65345e763befb36454d73 127.0.0.1:7004 slave 3e3a6cb0d9a9a87168e266b0a0b24026c0aae3f0 0 1385543179583 0 connected
    f093c80dde814da99c5cf72a7dd01590792b783b :0 myself,master - 0 0 0 connected
    2938205e12de373867bf38f1ca29d31d0ddb3e46 127.0.0.1:7002 slave 3c3a0c74aae0b56170ccb03a76b60cfe7dc1912e 0 1385543178072 3 connected
    a211e242fc6b22a9427fed61285e85892fa04e08 127.0.0.1:7003 slave 97a3a64667477371c4479320d683e4c8db5858b1 0 1385543178575 0 connected
    97a3a64667477371c4479320d683e4c8db5858b1 127.0.0.1:7000 master - 0 1385543179080 0 connected 0-5959 10922-11422
    3c3a0c74aae0b56170ccb03a76b60cfe7dc1912e 127.0.0.1:7005 master - 0 1385543177568 3 connected 11423-16383

这时新节点已经连接到集群，能够重定向client到正确的节点，但是新节点相比于老节点有两个特点:

* 由于没有分配slot所以没有数据
* 由于没有分配slot，不能参与选举投票

使用redis-cli的重新分片命令可以为新的master节点分配slot。

#### 新增一个节点作为从节点

&emsp;&emsp;有两种方式将新增的节点成为从节点，最明显的方法是再次使用redis-cli:

    redis-cli --cluster add-node 127.0.0.1:7006 127.0.0.1:7000 --cluster-slave

随机从副本较少的主节点中选择一个成为其副本。

当然也可以指定主节点:

    redis-cli --cluster add-node 127.0.0.1:7006 127.0.0.1:7000 --cluster-slave --cluster-master-id 3c3a0c74aae0b56170ccb03a76b60cfe7dc1912e

还有其它的手动方式可以指定成为某个master的副本，使用CLUSTER REPLICATE命令。这个命令也可以用于将当前节点变成新指定的master的副本。

    redis 127.0.0.1:7006> cluster replicate 3c3a0c74aae0b56170ccb03a76b60cfe7dc1912e

现在集群中新加了一个副本，集群中的其它节点也知道了这个节点(配置更新后)。

    $ redis-cli -p 7000 cluster nodes | grep slave | grep 3c3a0c74aae0b56170ccb03a76b60cfe7dc1912e
    f093c80dde814da99c5cf72a7dd01590792b783b 127.0.0.1:7006 slave 3c3a0c74aae0b56170ccb03a76b60cfe7dc1912e 0 1385543617702 3 connected
    2938205e12de373867bf38f1ca29d31d0ddb3e46 127.0.0.1:7002 slave 3c3a0c74aae0b56170ccb03a76b60cfe7dc1912e 0 1385543617198 3 connected

## 移除节点

移除一个从节点直接使用del-node命令:

    redis-cli --cluster del-node 127.0.0.1:7000 `<node-id>`

第一个参数是集群中的任意一个节点，第二个参数是你想移除的节点。

移除master节点也可以用同样的方式，但是master节点必须是一个空节点。如果是一个非空节点，首先需要将数据分片到其它节点。另外一种方式是进行手动failover在其变成从节点后在直接移除，但是如果你是想减少集群中master的数量，这种方式是无效的只能进行重新分片。

## 副本漂移

&emsp;&emsp;在Redis集群中将一个salve迁移到其它master下，通过执行以下命令即可:

    CLUSTER REPLICATE <master-node-id>

这个操作被称为副本漂移(replicas migration)，能够提高集群的可靠性。

副本偏移的详细信息才Redis Cluster Specification，这里仅仅是提供大概的思路和应该怎样做。

进行副本偏移的原因是，集群对故障的抵抗能力与附加到master上的副本数量相关。

例如，如果集群每个master只有一个副本，当master与副本同时失效时，集群将不能正常运转，原因很简单，因为没有其它实例拥有master提供的hash slot的副本。网络分区可能会隔离一些节点，还有其它一些故障，类似节点的硬件与软件故障，这些都会导致集群不能正常工作。

为了提高可靠性可以为每个master增加副本，但是成本较高。我们可以为一些master添加更多的从节点，当某个master没有从节点时会触发副本漂移，将拥有多个副本的主节点下的一个副本漂移到这个孤立的master。

关于副本漂移需要知道以下几点:

* 集群尝试从某时刻拥有最多副本的主节点下漂移一个副本
* 使用副本漂移，只需要为单个master添加少量的副本
* 通过一个redis.conf中的配置cluster-migration-barrier来控制副本漂移功能

## 升级集群中的节点

&emsp;&emsp;升级slave节点很简单，只需要停止节点重新启动一个新版本的redis。如果有client从salve节点读取数据，当这个slave不可用时这些client将会重新连接其它的slave。

升级master节点有点复杂，可参考下面的步骤:

* 在master的一个副本上执行CLUSTER FAILOVER命令手动触发failover
* 等待master转为slave
* 和升级从节点的方式一样升级这个节点
* 如果你想将该节点在恢复为master，则在该节点上手动触发failover即可

按照以上步骤，你可以逐个升级集群中的节点，直到所有节点升级完成。

## 迁移到Redis集群

&emsp;&emsp;想要迁移到Redis集群的用户可能只有一个主节点，或者使用一些内置算法、客户端或代理实现的分片算法将key分配到多个节点上。

以上两种情况都可以很容易的迁移到Redis集群，最重要的是是否使用了多key操作，有以下三种不同的情况:

1. 不使用多个键操作、事务或涉及多个键的Lua脚本，key是独立访问
2. 使用了多个键操作、事务或涉及多个键的Lua脚本，但是这些key具有相同的hash tag
3. 使用了多个键操作、事务或涉及多个键的Lua脚本，这些key没有相同的hash tag

这三种情况Redis集群无法处理，需要修改程序不使用多key操作或者使用相同的hash tag。

假如将预先存在的数据集分片到N个master，N=1时则不需要进行分片，为了将数据迁移到Redis集群需要以下步骤:

1. 停止client。当前无法自动实时迁移到Redis群集。你可能能够在应用程序/环境的上下文中协调实时迁移。
2. 对所有节点使用BGREWRITEAOF命令生成aof文件
3. 将所有的aof文件保存到某处，停止旧的reids实例
4. 创建一个由N个master和0个slave组成的Redis集群。稍后在添加salve，确保所有节点使用aof进行持久化
5. 停止所有的集群节点，使用之前保存的aof替换集群中的aof文件
6. 重启启动集群，可能会出现根据配置key不应该属于当前节点的问题
7. 使用redis-cli --cluster fix命令修复集群，根据节点的hash slot配置对key进行迁移
8. 使用redis-cli --cluster check命令检查集群是否ok
9. 修改客户端使用一个支持集群功能的库

还有另一种方式从外部运行的redis实例导入数据到集群，即使用redis-cli --cluster import命令。

这个命令移动redis实例中所有的key(从源实例中删除key)到集群。如果源实例使用的是Redis2.8操作可能慢，因为Reis2.8没有实现迁移连接的缓存。因此在执行此操作之前，你可能需要使用Redis3.x版本重新启动源实例。