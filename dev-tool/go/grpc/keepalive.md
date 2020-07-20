# Keepalive

>原文: https://github.com/grpc/grpc-go/blob/master/Documentation/keepalive.md

gRPC在传输连接上发送http2 ping来检测连接是否关闭。如果在一定时间内另一方没有确认，连接将关闭。注意，只有在连接上没有活动时才需要ping。


## 客户端应该设置什么

对于大多数用户来说，将[客户端参数](https://pkg.go.dev/google.golang.org/grpc/keepalive?tab=doc#ClientParameters)设置为[dail选项](https://pkg.go.dev/google.golang.org/grpc?tab=doc#WithKeepaliveParams)就足够了。


## 客户端发生了什么

(以下行为是针对gRPC-go，其它语言可能稍有不同)

当连接上没有活动时，查过Time时间后，客户端将发送ping，服务端返回一个ping ack。客户端发送ping后将最多等待Timeout时间等待服务端的回复。

## 服务端发生了什么

服务端也有类似的Time和Timeout配置，可以配置其它[服务端连接参数](https://pkg.go.dev/google.golang.org/grpc/keepalive?tab=doc#ServerParameters)


## 强制策略

[强制策略](https://pkg.go.dev/google.golang.org/grpc/keepalive?tab=doc#EnforcementPolicy)是服务端上的一种特殊设置，用于保护服务器免受恶意客户端的攻击。

检测到以下行为时，服务端将发送GOAWAY并关闭连接

* 客户端频繁发送ping

* 客户端在没有活动的stream是仍然发送ping，并且服务端禁止这种行为时