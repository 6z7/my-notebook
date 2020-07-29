# etcd客户端设计

## 介绍

etcd服务器已经通过多年的故障注入测试证明了它的健壮性。复杂的应用程序逻辑已经由etcd服务器及其数据存储处理。尽管服务器组件是正确的，但它与客户端的组合需要一组不同的复杂协议，以保证其在错误条件下的正确性和高可用性。理想情况下，etcd服务器提供许多物理机的一个逻辑集群视图，并且客户端在副本之间实现自动故障转移。该文档介绍了客户端架构决策及其实现细节。

## 词汇表 

clientv3: etcd Official Go client for etcd v3 API.

clientv3-grpc1.0: Official client implementation, with grpc-go v1.0.x, which is used in latest etcd v3.1.

clientv3-grpc1.7: Official client implementation, with grpc-go v1.7.x, which is used in latest etcd v3.2 and v3.3.

clientv3-grpc1.23: Official client implementation, with grpc-go v1.23.x, which is used in latest etcd v3.4.

平衡器(Balancer):etcd客户端负载均衡实现了重试和故障转移。etcd客户端会自动在多个节点之间自动负载均衡。

端点(Endpoints):客户端可以连接到的etcd服务器端点列表。

固定端点(Pinned endpoint):当配置了多个服务端点时，为了减少与集群的连接数，<=v3.3的客户端的平衡器会选择其中的一个建立tcp连接。在v3.4中，对每个请求平衡器会询调度一个端点进行连接，从而使负载更均匀地分配。

客户端连接(Client Connection):已经通过gRPC Dial与etcd服务端建立连接

子连接(Sub Connection):gRPC SubConn接口。每个子连接都包含一个地址列表。平衡器从已解析的地址列表中选择一个创建SubConn。gRPC ClientConn可以映射到到多个SubConn(如，example.com解析为10.10.10.1与10.10.10.2两个子连接)。etcd v3.4均衡器使用内部解析器为每个端点建立一个子连接。

暂时断开连接(Transient disconnect):当gRPC服务器返回状态错误码Unavailable。


## 对客户端的要求

正确性(Correctness)

活性(Liveness)

有效性(Effectiveness)

可移植性(Portability)


## 客户概况

etcd客户端实现以下组件:

* 建立与etcd集群的gRPC连接的平衡器

* 将RPC发送到etcd服务器的API客户端

* 错误处理程序，用于确定是重试失败的请求还是切换端点

语言可能在如何建立初始连接（例如配置TLS），如何编码和发送协议缓冲区消息到服务器，如何处理流RPC等方面有所不同。但是，从etcd服务器返回的错误将是相同的。因此错误处理和重试策略也应该如此。

例如，etcd服务器可能返回"rpc error: code = Unavailable desc = etcdserver: request timed out"，这是临时错误，需要重试。或return rpc error: code = InvalidArgument desc = etcdserver: key is not provided，这意味着请求无效，不应重试。Go客户端可以使用解析错误`google.golang.org/grpc/status.FromError`，而Java客户端可以使用解析错误`io.grpc.Status.fromThrowable`。


**clientv3-grpc1.0：平衡器概述**



