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

## 创建服务端

为了让RouteGuide服务发挥作用需要做以下两件事:

* 实现根据我们的服务定义生成的接口，用来处理实际请求

* 运行gRPC服务监听来自客户端的请求，并将它们分发到正确的服务实现中

在[grpc-go/examples/route_guide/server/server.go](https://github.com/grpc/grpc-go/tree/master/examples/route_guide/server/server.go)的例子中可以看到服务的具体实现。下面我们来看它是如何工作的。

**RouteGuid服务实现**

如你所见,我们的服务端有一个routeGuideServer结构实现了生成的RouteGuideServer接口:

```go
type routeGuideServer struct {
        ...
}
...

func (s *routeGuideServer) GetFeature(ctx context.Context, point *pb.Point) (*pb.Feature, error) {
        ...
}
...

func (s *routeGuideServer) ListFeatures(rect *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
        ...
}
...

func (s *routeGuideServer) RecordRoute(stream pb.RouteGuide_RecordRouteServer) error {
        ...
}
...

func (s *routeGuideServer) RouteChat(stream pb.RouteGuide_RouteChatServer) error {
        ...
}
...
```

**简单RPC**

routeGuideServer实现了所有服务方法。我们首先看下简单类型的实现

```go
func (s *routeGuideServer) GetFeature(ctx context.Context, point *pb.Point) (*pb.Feature, error) {
	for _, feature := range s.savedFeatures {
		if proto.Equal(feature.Location, point) {
			return feature, nil
		}
	}
	// No feature was found, return an unnamed feature
	return &pb.Feature{"", point}, nil
}
```

一个RPC context对象和客户端的请求参数Point被传递给方法，返回一个Feature对象和error。这个方法中构造了一个合适的Feature对象返回，返回的错误为null，告诉gRPC我们已经完成处理RPC可以返回Feature给客户端。


**服务端流式RPC**

ListFeatures是一个服务端流式RPC，因此我们可以返回多个Feature到客户端

```go
func (s *routeGuideServer) ListFeatures(rect *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
	for _, feature := range s.savedFeatures {
		if inRange(feature.Location, rect) {
			if err := stream.Send(feature); err != nil {
				return err
			}
		}
	}
	return nil
}
```

在这个方法中，我们返回多个Feature对象，使用RouteGuide_ListFeaturesServer的send方法发送到客户端，返回nil错误告诉gRPC我们已经完成了写响应。如果有任何错误发生，我们返回一个非nil错误，gRPC会将错误转为合适的RPC状态发送给客户端。

**客户端流式RPC**

```go
func (s *routeGuideServer) RecordRoute(stream pb.RouteGuide_RecordRouteServer) error {
	var pointCount, featureCount, distance int32
	var lastPoint *pb.Point
	startTime := time.Now()
	for {
		point, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&pb.RouteSummary{
				PointCount:   pointCount,
				FeatureCount: featureCount,
				Distance:     distance,
				ElapsedTime:  int32(endTime.Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}
		pointCount++
		for _, feature := range s.savedFeatures {
			if proto.Equal(feature.Location, point) {
				featureCount++
			}
		}
		if lastPoint != nil {
			distance += calcDistance(lastPoint, point)
		}
		lastPoint = point
	}
}
```
方法中使用RouteGuide_RecordRouteServer的Recv方法，反复读入客户端的请求直到没有消息。

**双向流式RPC**

```go
func (s *routeGuideServer) RouteChat(stream pb.RouteGuide_RouteChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		key := serialize(in.Location)
                ... // look for notes to be sent to client
		for _, note := range s.routeNotes[key] {
			if err := stream.Send(note); err != nil {
				return err
			}
		}
	}
}
```
可以发现读写的方式和客户端流式RPC非常相似，处理服务端使用Send()方法而不是SendAndClose()，因为是要写多个响应返回

**启动服务端**

一旦我们实现了所有的服务方法，就可以启动gRPC服务来为客户端提供服务。

```go
flag.Parse()
lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
if err != nil {
        log.Fatalf("failed to listen: %v", err)
}
grpcServer := grpc.NewServer()
pb.RegisterRouteGuideServer(grpcServer, &routeGuideServer{})
... // determine whether to use TLS
grpcServer.Serve(lis)
```

构建和启动服务，需要做:

1. 指定兼容的端口

2. 使用`grpc.NewServer()`创建gRPC服务的实例

3. gRPC服务端上注册我们的服务实现

4. 在服务端调用`Serve()`来执行阻塞等待，直到进程被终止或Stop()被调用

## 创建客户端

在这一节，我们如何为RouteGuid服务创建一个客户端，完成的日子在[grpc-go/examples/route_guide/client/client.go](https://github.com/grpc/grpc-go/blob/master/examples/route_guide/client/client.go)

**创建一个stub**

为了调用服务端方法，我们首先需要创建一个gRPC通道用于与服务端通信。为了创建通道，需要将服务端的地址和端口传入`grpc.Dial()`方法 如下:

```go
conn, err := grpc.Dial(*serverAddr)
if err != nil {
    ...
}
defer conn.Close()
```

你也可以使用`DialOptions`设置身份认证凭证(如，TLS、GCE、JWT)传递给`grpc.Dial`。

一旦gRPC通道设置完毕,我们需要一个客户端stub来处理PRC。通过使用pb包中的`NewRouteGuideClient`方法可以获得stub。

```go
client := pb.NewRouteGuideClient(conn)
```

**调用服务端方法**

现在来看下如何调用服务端的方法。需要注意的是在gRPC-Go中，RPC操作使用的是同步阻塞模式，这意味着RPC将阻塞等待服务单的响应。

**简单RPC**

调用简单的RPC方法GetFeature，与直接调用本地方法很相似

```go
feature, err := client.GetFeature(context.Background(), &pb.Point{409146138, -746188906})
if err != nil {
        ...
}
```

在这个方法里，我们创建和填充一个请求的pb对象参数，也创建了一个`context.Context`对象，用于根据需要改变RPC行为，如超时取消RPC。如果调用没用返回错误，我们可以从返回的一个参数中读取服务端响应的消息。


**服务端流式RPC**

这里我们调用服务端流式RPC方法ListFeatures，该方法返回一个流式Feature。

```go
rect := &pb.Rectangle{ ... }  // initialize a pb.Rectangle
stream, err := client.ListFeatures(context.Background(), rect)
if err != nil {
    ...
}
for {
    feature, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
    }
    log.Println(feature)
}
```

方法不是直接返回响应结果，而是一个RouteGuide_ListFeaturesClient流式对象，使用Recv方法重复从中读取消息，直到消息被读取完毕。客户端需要检查Recv()返回的error信息，如果为null说明可以继续读取，如果是io.EOF说明消息读取完毕，其它错误说明有异常。


**客户端流式RPC**

客户端流式RPC方法RecordRoute与服务端流式RPC方法非常相似，除了传递context得到一个RouteGuide_RecordRouteClient流式响应，可以用于读写消息

```go
// Create a random number of random points
r := rand.New(rand.NewSource(time.Now().UnixNano()))
pointCount := int(r.Int31n(100)) + 2 // Traverse at least two points
var points []*pb.Point
for i := 0; i < pointCount; i++ {
	points = append(points, randomPoint(r))
}
log.Printf("Traversing %d points.", len(points))
stream, err := client.RecordRoute(context.Background())
if err != nil {
	log.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
}
for _, point := range points {
	if err := stream.Send(point); err != nil {
		if err == io.EOF {
			break
		}
		log.Fatalf("%v.Send(%v) = %v", stream, point, err)
	}
}
reply, err := stream.CloseAndRecv()
if err != nil {
	log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
}
log.Printf("Route summary: %v", reply)
```

RouteGuide_RecordRouteClient有一个Send()方法，可以使用该流式方法不断发送消息到服务端，需要调用CloseAndRecv()方法让gRPC知道我们已经完成了发送消息在等待响应，如果返回的错误是nil,则返回的第一个参数就是服务端的响应结果。


**双向流式RPC**

最后，来看下双向流式RPC方法RouteChat()，在这个方法中仅传递context对象作为参数，返回了一个可用与读写消息的流式对象。当获取到流方法的返回值后，服务端可能还在向流中写响应消息。

```go
stream, err := client.RouteChat(context.Background())
waitc := make(chan struct{})
go func() {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// read done.
			close(waitc)
			return
		}
		if err != nil {
			log.Fatalf("Failed to receive a note : %v", err)
		}
		log.Printf("Got message %s at point(%d, %d)", in.Message, in.Location.Latitude, in.Location.Longitude)
	}
}()
for _, note := range notes {
	if err := stream.Send(note); err != nil {
		log.Fatalf("Failed to send a note: %v", err)
	}
}
stream.CloseSend()
<-waitc
```

语法与客户端流式RPC很相似，除了在客户端流式RPC中我们仅调用一次CloseSend()。尽管每一方总是按照消息写入的顺序获取另一方的消息，但是客户端和服务端都可以以任何顺序读写消息—流操作完全独立。

## 试试看

切换到demo所在目录

```sh
$ cd $GOPATH/src/google.golang.org/grpc/examples/route_guide
```

运行服务端

```sh
$ go run server/server.go
```

换一个终端窗口运行客户端

```sh
$ go run client/client.go
```