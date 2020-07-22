# gRPC名称解析(gRPC Name Resolution)

>原文：https://github.com/grpc/grpc/blob/master/doc/naming.md

gRPC支持DNS作为默认的名字系统(name-system)。在各种部署中使用了许多可替换的名称系统。我们提供一个足够通用的API来支持一系列的名称系统和相应的名称语法。各种语言的gRPC客户端库中将提供一个插件机制，以便可以插入不同名称系统的解析器。

## 名字语法

用于gRPC通道构造的完全限定的自包含名称使用[RFC 3986](https://tools.ietf.org/html/rfc3986)中定义的URI语法。

URI的scheme用来确定使用哪个解析器插件(resolver plugin)。如果没有指定scheme前缀或scheme是unknown，将默认使用dns scheme。

URI路径指示要解析的名字。

大多数的gRPC实现支持以下URI格式：

* `dns:[//authority/]host[:port]` -- DNS (默认)

    * host：dns要解析的主机
    * port: 地址端口，没有指定默认使用443(对于不安全的通道，默认80)
    * authority：指定要使用的DNS服务器，某些语言的实现才支持这个特性

* `unix:path`或`unix://absolute_path` -- Unix domain sockets

下面的scheme格式gRPC C-core的实现支持，其它语言不一定支持

* ipv4:address[:port][,address[:port],...] -- IPv4 addresses

* ipv6:address[:port][,address[:port],...] -- IPv6 addresses


未来可能有其它scheme会添加，如etcd。


## 解析器插件(Resolver Plugins)

gRPC客户端使用scheme去选择响应的解析器插件，并将完全限定的名称字符串传递给它进行解析。

解析器应该能够联系权威机构进行解析，然后将解析结果返回到gRPC客户端。返回内容包括:

* 地址列表(ip+port)。每个地址可能有一组与之关联的任意属性(键/值对)，这些属性可用于从解析器向负载平衡策略传递信息。

* 服务配置

插件API允许解析器持续监视端点，并根据需要返回更新的解析。