# gRpc核心概念、架构和生命周期

> gRpc中的关键概念、架构概览和RPC生命周期的介绍

## 简介

**Service定义**

&emsp; &nbsp;像许多RPC系统一样，gRPC基于定义服务的思想，指定可以用参数和返回类型远程调用的方法。默认情况下使用protocol buffers作为IDL(interface definition language)，用来描述service接口和有效负载(payload)消息的结构。如果需要可以使用其它的序列化工具替代。

```
service HelloService {
  rpc SayHello (HelloRequest) returns (HelloResponse);
}

message HelloRequest {
  string greeting = 1;
}

message HelloResponse {
  string reply = 1;
}
```

gRpc允许你定义四种service方法

* 一元RPC(Unary RPCs):client发送一个请求到server，server返回一个结果，就像一个函数调用。

`rpc SayHello(HelloRequest) returns (HelloResponse);`

* 服务器流RPC(Server streaming RPCs):client发送一个请求到server，获取一个流以读回一系列消息。client从返回的流中读取消息直到读完。gRpc保证单个RPC调用中消息的顺序。

`rpc LotsOfReplies(HelloRequest) returns (stream HelloResponse);`

* 客户端流RPC(Client streaming RPCs):client使用流写一系列消息发送到server，client写入消息完成后将等待服务器读取消息并返回响应。同样，gRpc保证单个RPC调用中消息的顺序。

`rpc LotsOfGreetings(stream HelloRequest) returns (HelloResponse);`

* 双向流RPC(Bidirectional streaming RPCs):两端使用读写流发送消息。这两个流独立运行，因此client和server可以按它们喜欢的顺序读写。

`rpc BidiHello(stream HelloRequest) returns (stream HelloResponse);`


## 使用API

&emsp; &nbsp;从.proto文件中的服务定义开始，gRPC提供生成client和server代码的pb编译器插件。gRpc通常在client端调用这些API,并在server端实现相应的API。

* 在server端，server实现服务声明的方法并运行gRpc服务器处理client的调用。gRpc的基础设施解码传入的请求，执行服务方法，并对服务的响应进行编码。

* 在client端，client有一个称为stub的本地对象(对某些语言，首选术语是client)，它实现与服务相同的方法。这样client就可以在本地对象上调用这些方法，将调用的参数包装在适当的协议pb消息类型中

## 同步与异步

&emsp; &nbsp;同步的RPC调用在响应从服务器到达之前一直被阻塞。另一方面，网络本质上是异步的，在许多情况下，能够在不阻塞当前线程的情况下启动RPC是很有用的。

&emsp; &nbsp;大多数语言中的gRPC编程API有同步和异步两种风格。您可以在每种语言的教程和参考文档中找到更多信息（完整的参考文档很快就会提供）。

## RPC生命周期

&emsp; &nbsp;在本节中，你可以更详细地了解当gRPC客户端调用gRPC服务器方法时会发生什么。有关完整的实现详细信息，请参阅特定于语言的页面。

**一元RPC(Unary RPC)**

首先考虑最简单的RPC类型，client发送一个请求并返回一个响应。


**服务端流式RPC(Server streaming RPC)**

**Client streaming RPC**

**Bidirectional streaming RPC**

**截止/超时时间(Deadlines/Timeouts )

&emsp; &nbsp;gRPC允许客户端指定在RPC因`DEADLINE_EXCEEDED`错误而终止RPC之前，它愿意等待RPC完成的时间。在服务器端，服务器可以查询某个特定的RPC是否超时，或者还有多少时间来完成RPC。

&emsp; &nbsp;指定截止时间或超时是特定于语言的：有些语言使用超时时间(持续时间)，有些语言使用截止时间(固定时间点)，可能有也可能没有默认的截止时间。

**RPC终止(RPC termination)**

&emsp; &nbsp;在gRPC中，client和server都对调用成功与否做出独立和局部的判断，它们的结论可能不一致。这意味着，你可以在服务器端成功地完成RPC（“我已经发送了我所有的响应！”）但在客户端失败（“响应在我的截止日期之后到达！”）。服务器也有可能在客户端发送完所有请求之前决定完成。

**取消RPC**

&emsp; &nbsp;client和server可以随时取消RPC。取消操作会立即终止RPC，因此还未进行的操作将不会被执行。

**元数据(Metadata)**

&emsp; &nbsp;元数据是与特定RPC调用有关的，以键值对列表形式出现的信息(如，认证详细信息)。其中键是字符串，值通常是字符串，也可以是二进制数据。元数据对gRPC本身是不透明的，它允许客户端提供与调用服务器相关的信息，反之亦然。

&emsp; &nbsp;对元数据的访问依赖于语言。

**通道(Channels)**

&emsp; &nbsp;gRPC通道提供到指定主机和端口上的gRPC服务器的连接。它在创建客户机stub时使用。客户端可以指定通道参数来修改gRPC的默认行为，例如打开或关闭消息压缩。通道有状态，包括已连接和空闲。

&emsp; &nbsp;gRPC如何关闭通道取决于语言。有些语言还允许查询通道状态。