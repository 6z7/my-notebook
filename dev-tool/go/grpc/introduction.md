# gRpc简介

&emsp;本文件将对gRpc和protocol buffers进行介绍。gRpc可以使用pb作为IDL(Interface Definition Language)，也可以作为底层消息交换的格式。

## 概述

&emsp; &nbsp; 在gRpc中，一个客户端应用可以直接调用在不同机器上的服务端接口，就像本地调用一样 。和许多RPC系统一样，gRPC基于定义服务的思想，指定可以用参数和返回类型远程调用的方法。在服务端，实现接口并运行一个gRpc服务用来处理客户端调用。在客户端，client有一个stub（在某些语言中称为client），它提供与服务端相同的方法。

![](./image/ingroduction1.svg)    

