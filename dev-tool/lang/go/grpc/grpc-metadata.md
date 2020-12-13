# 元数据(Metadata)

>原文：https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md

gRPC支持在客户端与服务端之间发送元数据。这个文档说明了gRPC-go如何进行发送和接收在元数据。


## metadata据结构

元数据可以通过[metadata](https://pkg.go.dev/google.golang.org/grpc/metadata?tab=doc)包创建。MD类型定义：

```go
type MD map[string][]string
```

Metadata可以像正常的map一样进行读取。需要注意的是，元数据的值类型是一个字符串切片([]string)，所以一个key可以有多个值。

## 创建metadata

```go
md := metadata.New(map[string]string{"key1": "val1", "key2": "val2"})
```

另一种方式，使用Paris，相同的key将会合并成一个list

```go
md := metadata.Pairs(
    "key1", "val1",
    "key1", "val1-2", // "key1" will have map value []string{"val1", "val1-2"}
    "key2", "val2",
)
```

>所有key将会被自动转为小写，如"key1"和"KEy1"是相同的key，他们的值将会被合并到同一个list中。这个规则应用于New和Pirs


## metadata中存储二进制数据

元数据中，key是字符串类型，但是值可以是字符串或二进制数据。将二进制数据存储到metadata中，只需要在key中添加"-bin"后缀。带有"-bin"后缀的key对应的值在创建元数据时会被编码。

```go
md := metadata.Pairs(
    "key", "string value",
    "key-bin", string([]byte{96, 102}), // this binary data will be encoded (base64) before sending
                                        // and will be decoded after being transferred.
)
```

## 从上下文中检索元数据

可以通过`FromIncomingContext`从context中检索元数据

```go
func (s *server) SomeRPC(ctx context.Context, in *pb.SomeRequest) (*pb.SomeResponse, err) {
    md, ok := metadata.FromIncomingContext(ctx)
    // do something with metadata
}
```


# 客户端发送和接收元数据

客户端发送和接收元数据的例子参见[这里](https://github.com/grpc/grpc-go/blob/master/examples/features/metadata/client/main.go)。


## 发送元数据

客户端有两种方式发送元数据到服务端。推荐的使用`AppendToOutgoingContext`。当上下文中不存在元数据直接添加，返回则将添加的kv合并到现有的元数据中。

```go
// create a new context with some metadata
ctx := metadata.AppendToOutgoingContext(ctx, "k1", "v1", "k1", "v2", "k2", "v3")

// later, add some more metadata to the context (e.g. in an interceptor)
ctx := metadata.AppendToOutgoingContext(ctx, "k3", "v4")

// make unary RPC
response, err := client.SomeRPC(ctx, someRequest)

// or make streaming RPC
stream, err := client.SomeStreamingRPC(ctx)
```

或者可以通过`NewOutgoingContext`将元数据附加到上下文中。这种方式将替换已经存在的元数据，因此如果需要必须小心保存现有的元数据。

```go
// create a new context with some metadata
md := metadata.Pairs("k1", "v1", "k1", "v2", "k2", "v3")
ctx := metadata.NewOutgoingContext(context.Background(), md)

// later, add some more metadata to the context (e.g. in an interceptor)
md, _ := metadata.FromOutgoingContext(ctx)
newMD := metadata.Pairs("k3", "v3")
ctx = metadata.NewContext(ctx, metadata.Join(metadata.New(send), newMD))

// make unary RPC
response, err := client.SomeRPC(ctx, someRequest)

// or make streaming RPC
stream, err := client.SomeStreamingRPC(ctx)
```


## 接收元数据

客户端可以通过heder和trailer接收元数据

**Unary call**

Header和trailer随一元调用作为选项传递

```go
var header, trailer metadata.MD // variable to store header and trailer
r, err := client.SomeRPC(
    ctx,
    someRequest,
    grpc.Header(&header),    // will retrieve header
    grpc.Trailer(&trailer),  // will retrieve trailer
)
// do something with header and trailer
```

**Streaming call**

流式调用包括：

* 服务端流式调用
* 客户端流式调用
* 双向流式调用

Header和trailer可以从返回的stream(ClientStream)中使用Header和Trailer方法进行接收

```go
stream, err := client.SomeStreamingRPC(ctx)

// retrieve header
header, err := stream.Header()

// retrieve trailer
trailer := stream.Trailer()
```


# 服务端发送和接收元数据

...



