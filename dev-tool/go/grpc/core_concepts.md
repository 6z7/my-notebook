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

