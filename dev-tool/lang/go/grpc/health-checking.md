# 健康检查协议

>原文: https://github.com/grpc/grpc/blob/master/doc/health-checking.md


健康检查用于探测服务是否可用。客户端到服务端的健康检查可以点对点进行，也可以通过某些控制系统进行。如果服务还没准备好响应请求,可以回复"unhealthy"。如果客户端在一段时间内没有收到响应，或者响应显示不健康(unhealthy)，客户端可以相应地采取行动。

GRPC服务被用作客户端到服务端场景和其它控制系统(如负载平衡)的健康检查机制。作为一个高级服务提供了一些好处。首先，由于它本身是一个GRPC服务，所以运行状况检查的格式与普通rpc相同。其次，它具有丰富的语义，如按服务进行健康状态检查。第三，作为GRPC服务，它能够重用所有现有的基础设施，因此服务器可以完全控制健康检查服务的访问。

## Service定义

```
syntax = "proto3";

package grpc.health.v1;

message HealthCheckRequest {
  string service = 1;
}

message HealthCheckResponse {
  enum ServingStatus {
    UNKNOWN = 0;
    SERVING = 1;
    NOT_SERVING = 2;
    SERVICE_UNKNOWN = 3;  // Used only by the Watch method.
  }
  ServingStatus status = 1;
}

service Health {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);

  rpc Watch(HealthCheckRequest) returns (stream HealthCheckResponse);
}
```

客户端可以通过调用Check方法来查询服务器的运行状况，并且应该在rpc上设置一个截止时间(deadline)。客户端可以选择设置要查询健康状况的服务名称。建议的服务名称格式为package_name.ServiceName，例如grpc.health.v1.Health。

服务端需要手动注册所有的服务并设置状态，包括空的服务名和服务状态。对于每个请求，如果能在注册的服务里找到对应请求的服务名称，则成功响应Ok状态，失败响应SERVING或NOT_SERVING状态；如果对应请求的服务名称没有注册，则返回NOT_FOUND状态

服务器使用一个空字符串代表查询所有服务器状态的的服务器名称。服务器可以对服务名进行精确匹配，而无需支持任何类型的通配符匹配，但服务所有者可以自由实现客户端和服务端都同意的更复杂的匹配语义。

如果rpc在一段时间后没有完成，客户端可以声明服务器不健康。客户端应该能够处理服务器没有健康检查的情况。


客户端可以调用Watch方法来执行流式健康检查。服务器将立即返回一条消息，指示当前的服务状态。随后，当服务的服务状态发生变化时，它将发送一条新消息。











