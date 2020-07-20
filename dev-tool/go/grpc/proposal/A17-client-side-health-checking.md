# Client-Side Health Checking

> 原文：https://github.com/grpc/proposal/blob/master/A17-client-side-health-checking.md

## 摘要

本文提出了一种在客户端支持应用程序健康检查的设计。

## 背景

gRPC has an existing health-checking mechanism, which allows server applications to signal that they are not healthy without actually tearing down connections to clients. This is used when (e.g.) a server is itself up but another service that it depends on is not available.

Currently, this mechanism is used in some look-aside load-balancing implementations to provide centralized health-checking of backend servers from the balancer infrastructure. However, there is no support for using this health-checking mechanism from the client side, which is needed when not using look-aside load balancing (or when falling back from look-aside load balancing to directly contacting backends when the balancers are unreachable).

## Related Proposals

N/A


## 提议(Proposal)

gRPC客户端将能够配置可以将健康检查rpc发送到它所连接的每个后端。每当后端响应为不健康时，客户机的LB策略将停止向该后端发送请求，直到该后端再次报告健康为止。

注意，由于健康检查服务需要一个服务名称，因此客户端将需要配置一个要使用的服务名称。但是，按照约定，它可以使用空字符串，这意味着给定主机/端口上所有服务的运行状况将由一个开关控制。从语义上说，空字符串用于表示服务器的总体运行状况，而不是服务器上运行的任何单个服务的运行状况。

## Watch-Based Health Checking Protocol

当前的健康检查协议是一个请求-响应模式，其中客户端需要定期轮询服务器。这对于通过负载均衡器进行集中运行状况检查已经足够了，其中运行状况检查来自少数客户机，并且存在定期轮询每个客户机的现有基础设施。但是，如果有大量客户端发起健康检查请求，那么出于可伸缩性和带宽使用的原因，我们将需要将健康检查协议转换为基于流监视的API。

注意，这种方法的一个缺点是，服务器端健康检查代码在变得不健康时可能无法发送更新。如果问题是由于服务器停止了对I/O的轮询，那么问题将被keepalives捕获，此时客户机将断开连接。但是，如果问题是由健康检查服务中的错误引起的，那么服务器可能仍然在响应，但未能通知客户机它不健康。


