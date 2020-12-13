# 为什么使用gRPC网关

etcd v3使用[gRPC](https://www.grpc.io/)作为其消息传递协议。etcd项目包括一个基于gRPC的[Go客户端](https://github.com/etcd-io/etcd/tree/master/clientv3)和一个命令行实用程序[etcdctl](https://github.com/etcd-io/etcd/tree/master/etcdctl)，用于通过gRPC与etcd集群进行通信。对于不支持gRPC的语言，etcd提供了[JSON gRPC网关](https://github.com/grpc-ecosystem/grpc-gateway)。该网关提供一个RESTful代理，该代理将HTTP/JSON请求转换为gRPC消息。


## 使用gRPC网关

网关接收etcd [protocol buffer](https://etcd.io/docs/v3.4.0/dev-guide/api_reference_v3/)消息定义对应的[JSON映射](https://developers.google.com/protocol-buffers/docs/proto3#json)。由于etcd中的key和value字段是字节数组，所以JSON中需要使用base64编码。下边是一个使用curl的例子，其它任何HTTP/JSON客户端类似。

## 网关API变更记录

自etcd v3.3起，gRPC网关端点已更改：

* etcd v3.2或更低版本仅可使用[CLIENT-URL]/v3alpha/*

* etcd v3.3使用[CLIENT-URL]/v3beta/\*，同时保留[CLIENT-URL]/v3alpha/*

* etcd v3.4使用[CLIENT-URL]/v3/* ，同时保留[CLIENT-URL]/v3beta/*

    * [CLIENT-URL]/v3alpha/* 废弃

* etcd v3.5或更高版本仅使用[CLIENT-URL]/v3/*     
    
    * [CLIENT-URL]/v3beta/* 废弃

gRPC-gateway does not support authentication using TLS Common Name.


## 设置和读取key

使用`/v3/kv/range`和`/v3/kv/put`读取和写入key：

```curl
<<COMMENT
https://www.base64encode.org/
foo is 'Zm9v' in Base64
bar is 'YmFy'
COMMENT

curl -L http://localhost:2379/v3/kv/put \
  -X POST -d '{"key": "Zm9v", "value": "YmFy"}'
# {"header":{"cluster_id":"12585971608760269493","member_id":"13847567121247652255","revision":"2","raft_term":"3"}}

curl -L http://localhost:2379/v3/kv/range \
  -X POST -d '{"key": "Zm9v"}'
# {"header":{"cluster_id":"12585971608760269493","member_id":"13847567121247652255","revision":"2","raft_term":"3"},"kvs":[{"key":"Zm9v","create_revision":"2","mod_revision":"2","version":"1","value":"YmFy"}],"count":"1"}

# get all keys prefixed with "foo"
curl -L http://localhost:2379/v3/kv/range \
  -X POST -d '{"key": "Zm9v", "range_end": "Zm9w"}'
# {"header":{"cluster_id":"12585971608760269493","member_id":"13847567121247652255","revision":"2","raft_term":"3"},"kvs":[{"key":"Zm9v","create_revision":"2","mod_revision":"2","version":"1","value":"YmFy"}],"count":"1"}
```

## 监视key

使用`/v3/watch`监视key:

```
curl -N http://localhost:2379/v3/watch \
  -X POST -d '{"create_request": {"key":"Zm9v"} }' &
# {"result":{"header":{"cluster_id":"12585971608760269493","member_id":"13847567121247652255","revision":"1","raft_term":"2"},"created":true}}

curl -L http://localhost:2379/v3/kv/put \
  -X POST -d '{"key": "Zm9v", "value": "YmFy"}' >/dev/null 2>&1
# {"result":{"header":{"cluster_id":"12585971608760269493","member_id":"13847567121247652255","revision":"2","raft_term":"2"},"events":[{"kv":{"key":"Zm9v","create_revision":"2","mod_revision":"2","version":"1","value":"YmFy"}}]}}
```

......

