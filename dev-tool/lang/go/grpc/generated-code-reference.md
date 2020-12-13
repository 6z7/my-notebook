# Go代码生成参考

这篇文章主要描述protoc编译器使用grpc插件[protoc-gen-go](https://github.com/golang/protobuf/tree/master/protoc-gen-go)编译.proto文件时的代码生成过程。

线程安全:客户端与服务端RPC调用都是线程安全的，可以进行并发操作；但需要注意的是，对于单个流，传入和传出数据是双向但是是串行的，因此单个流不支持并发读或并发写，但读写可以安全的并发。


## 生成的服务端接口中的方法

在服务端，`.proto`文件中定义的Bar服务，编译后会生成注册服务的方法

`func RegisterBarServer(s *grpc.Server, srv BarServer)`
  

在应用中使用BarService接口，使用RegisterBarServer将实现注册到gRPC上。


**一元方法**

一元简单方法在生成后拥有如下签名:

`Foo(context.Context, *MsgA) (*MsgB, error)`

其中MsgA是客户端发送的消息，MsgB是服务端返回的响应消息，这些消息都是pb中定义的message类型


**服务端流式方法**

生成的方法签名如下:

`Foo(*MsgA, <ServiceName>_FooServer) error`

其中MsgA是客户端发送的单一请求，\<ServiceName\>_FooServer参数代表服务端到客户端的MsgB类型的流。

\<ServiceName\>_FooServer是一个内嵌`grpc.ServerStream`的接口:

```go
type <ServiceName>_FooServer interface {
	Send(*MsgB) error
	grpc.ServerStream
}
```

服务端处理程序可以通过这个参数的Send方法向客户端发送protobuf消息流，服务端到客户端的流的结束是由处理程序中return触发的。


**客户端流式方法**

生成的方法签名:

`Foo(<ServiceName>_FooServer) error`

\<ServiceName>_FooServer可以读取客户端到服务端的消息流，并返回单一响应到客户端。

\<ServiceName\>_FooServer是一个内嵌`grpc.ServerStream`的接口:

```go
type <ServiceName>_FooServer interface {
	SendAndClose(*MsgA) error
	Recv() (*MsgB, error)
	grpc.ServerStream
}
```

服务端处理程序为了从客户端获取所有的消息可以反复调用Recv方法。Recv返回(nil,io.EOF)代表流中的消息已经读取完毕。通过SendAndClose方法从服务端发送单一的响应到客户端，注意该方法仅能调用一次。


**双向流式方法**

生成的方法签名如下:

`Foo(<ServiceName>_FooServer) erro`

\<ServiceName>_FooServer可用于访问客户端到服务端的消息流也可以用于访问服务端到客户端的消息流。\<ServiceName>_FooServer是一个内嵌grpc.ServerStream的接口:

```go
type <ServiceName>_FooServer interface {
	Send(*MsgA) error
	Recv() (*MsgB, error)
	grpc.ServerStream
}
```

服务端处理程序为了读取客户端到服务单的消息流，可以重复调用Recv方法。Recv返回(nil,io.EOF)代表流中的消息已经读取完毕。可以反复调用Send方法向服务端到客户端的消息流中写响应数据。服务端到客户端流的结束是由处理方法的return暗示的。

## 生成的客户端接口中的方法

对于客户端来说，.proto文件中定义的service Bar会生成对应的接口

**一元方法**

生成的方法如下签名:

`(ctx context.Context, in *MsgA, opts ...grpc.CallOption) (*MsgB, error)`

**服务端流式方法**

生成的方法签名:

`Foo(ctx context.Context, in *MsgA, opts ...grpc.CallOption) (<ServiceName>_FooClient, error)`

\<ServiceName>_FooClient代表服务端发送到客户端的MsgB类型的消息流。 它是一个内嵌`grpc.ClientStream`的接口：

```go
type <ServiceName>_FooClient interface {
	Recv() (*MsgB, error)
	grpc.ClientStream
}
```

客户端可以反复调用Recv方法读取服务端发送到客户端的响应消息流。当Recv方法返回(nil,io.EOF)时代表流结束


**客户端流式方法**

生成的方法签名:

`Foo(ctx context.Context, opts ...grpc.CallOption) (<ServiceName>_FooClient, error)`

\<ServiceName>_FooClient代表客户端到服务端的MsgA消息流，它是一个内嵌grpc.ClientStream的接口:

```go
type <ServiceName>_FooClient interface {
	Send(*MsgA) error
	CloseAndRecv() (*MsgA, error)
	grpc.ClientStream
}
```

客户端可以反复调用Send方法向服务端发送消息。CloseAndRecv方法仅能调用一次，用于关闭客户端到服务端的消息流并返回服务端返回的单一响应。

**双向流式方法**

生成的方法签名如下:

`Foo(ctx context.Context, opts ...grpc.CallOption) (<ServiceName>_FooClient, error)`

\<ServiceName>_FooClient代表客户端到服务端和服务端到客户端的双向消息流，接口定义:

```go
type <ServiceName>_FooClient interface {
	Send(*MsgA) error
	Recv() (*MsgB, error)
	grpc.ClientStream
}
```

客户端可以反复调用Send方法发送消息到服务端，也可以重复调用Recv方法接收服务端发挥的响应消息。

当Recv方法的返回值为(nil,io.EOF)时代表服务端到客户端的流结束；当客户单调用CloseSend方法时代表客户端到服务端的流式结束。