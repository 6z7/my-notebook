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

## 

即使没有failover在处理，哨兵也会一直尝试在被监视的master上设置当前配置。

* 副本声称是master,将会被配置为master的副本

* 副本连接了错误的master,将会被重新配置为正确的master的副本

要让哨兵重新配置副本，必须在一段时间内观察到错误的配置，这段时间比用于广播新配置的时间长。





