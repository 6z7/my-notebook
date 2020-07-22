# 服务配置(Service Config in gRPC)

>原文：https://github.com/grpc/grpc/blob/master/doc/service_config.md

## 目的

服务配置是一种机制，允许服务所有者将发布的参数自动应用到服务的所有客户端。

## 格式

服务配置的格式通过protocol buffer消息[grpc.service_config.ServiceConfig](https://github.com/grpc/grpc-proto/blob/master/grpc/service_config/service_config.proto)定义。

请注意，随着新功能的引入，将来可能会添加新字段。

## 结构

服务配置与服务器名相关联。当[名称解析插件](https://github.com/grpc/grpc/blob/master/doc/naming.md)解析特定的服务器名称时，它将返回解析后的地址和服务配置。

名字解析时以JSON格式返回服务配置到gRPC客户端。如果解析器获得的是protobuf形式的配置，需要根据[映射规则](https://developers.google.com/protocol-buffers/docs/proto3#json)转为对应的JSON。如果解析器获得的是JSON格式的配置，则可以直接返回。

有关DNS解析插件如何支持服务配置的详细信息，请参阅[gRFC A2: Service Config via DNS](https://github.com/grpc/proposal/blob/master/A2-service-configs-in-dns.md)。


## 例子


protobuf格式的服务配置:

```
{
  // Use round_robin LB policy.
  load_balancing_config: { round_robin: {} }
  // This method config applies to method "foo/bar" and to all methods
  // of service "baz".
  method_config: {
    name: {
      service: "foo"
      method: "bar"
    }
    name: {
      service: "baz"
    }
    // Default timeout for matching methods.
    timeout: {
      seconds: 1
      nanos: 1
    }
  }
}
```

json格式的服务配置:

```
{
  "loadBalancingConfig": [ { "round_robin": {} } ],
  "methodConfig": [
    {
      "name": [
        { "service": "foo", "method": "bar" },
        { "service": "baz" }
      ],
      "timeout": "1.0000000001s"
    }
  ]
}
```

## APIs

服务配置被以下API：

* In the resolver API, used by resolver plugins to return the service config to the gRPC client.

* In the gRPC client API, where users can query the channel to obtain the service config associated with the channel (for debugging purposes).

* In the gRPC client API, where users can set the service config explicitly. This can be used to set the config in unit tests. It can also be used to set the default config that will be used if the resolver plugin does not return a service config.