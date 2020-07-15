# 基础教程

> Go语言中使用gRPC的基本教程

本教程提供了关于gRPC在Go中如何使用的基本指导。

通过该教程，你将学到:

* 在`.proto`文件中定义一个service
* 使用protocol buffer编译器生成客户端与服务端代码
* 使用Go中的gRPC API为定义的服务编写一个简单的客户端与服务端

本教程假设你已经阅读了gRPC简介和熟悉[protocol buffers](https://developers.google.com/protocol-buffers/docs/overview)。需要注意，教程中的例子使用的是proto3版本，你可以在[proto3语法教程](https://developers.google.com/protocol-buffers/docs/proto3)和[Go代码生成](https://developers.google.com/protocol-buffers/docs/reference/go-generated)中找到更多资料。


## 为什么使用gRPC

我们的示例是一个简单的路由映射应用程序，它允许客户端获取有关其路由特性的信息，创建其路由的摘要，并与服务器和其它客户端交换路由信息，如流量更新。

使用gRPC我们可以在`.proto`文件中定义我们的服务，生成gPRC所支持语言的任意客户端和服务端。gRPC屏蔽了不同语言之间通信的复杂问题。我们获得了使用PB带来的所有优点，包括高效的序列化、简单的IDL和容易更新的接口。

## 定义service

我们首先使用PB定义一个gPRC service、request和response方法，完整示参见：[grpc-go](https://github.com/grpc/grpc-go/blob/master/examples/route_guide/routeguide/route_guide.proto):


定义服务
```
service RouteGuide {
   ...
}
```
之后服务中定义在rpc方法，指定它们的request和response类型。gRPC中允许定义4种服务方法，所有这些都在RouteGuide服务中使用：

* 简单RPC:client使用stub发送请求到server，等待响应返回，就像一个普通的函数调用

`rpc GetFeature(Point) returns (Feature) {}`

* 服务端流式RPC:client发送请求到server,获得服务端的流用于读取返回的一系列消息。要获得流式响应，在返回类型之前放置`stream`关键字即可

`rpc ListFeatures(Rectangle) returns (stream Feature) {}`      

* 客户端流式RPC：client发送一系列消息到server,client发送完成，等待服务端响应

`rpc RecordRoute(stream Point) returns (RouteSummary) {}`

* 双向流式RPC

`rpc RouteChat(stream RouteNote) returns (stream RouteNote) {}`

我们的`.proto`文件也包含在服务方法中使用的请求响应类型的PB消息类型的定义。如Point消息类型

```
message Point {
  int32 latitude = 1;
  int32 longitude = 2;
}
```

## 生成client和server代码

下一步我们需要根据定义的`.proto`文件生成gRPC客户端与服务端代码。使用PB编译器`protoc`(需要安装gRPC Go插件)来生成代码。

在route_guid目录下，执行:

` protoc -I routeguide/ routeguide/route_guide.proto --go_out=plugins=grpc:routeguide`

运行上面的命令，在route_guid下的routeguide目录中会生成`route_guide.pb.go`文件。

生成的文件包含:

* 用于填充、序列化和检索请求和响应消息类型的所有PB代码

* 客户端使用RouteGuide服务中定义的方法调用的接口类型(或stub)

* 服务端实现RouteGuide服务中定义的方法












