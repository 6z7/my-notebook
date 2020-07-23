## 启动一个独立节点

```
./etcd

./etcdctl put foo bar

./etcdctl get foo
```

## 启动一个集群

1. 安装多进程管理工具goreman

`go get github.com/mattn/goreman`

2. 使用官方提供的[Procfile](https://github.com/etcd-io/etcd/blob/master/Procfile.v2)文件启动

`goreman -f Procfile start`

将在本地启动3个成员节点
localhost:2379, localhost:22379, 和 localhost:32379

## 与集群交互

打印集群中的所有成员

`etcdctl --write-out=table --endpoints=localhost:2379 member list`

```
+------------------+---------+--------+------------------------+------------------------+------------+
|        ID        | STATUS  |  NAME  |       PEER ADDRS       |      CLIENT ADDRS      | IS LEARNER |
+------------------+---------+--------+------------------------+------------------------+------------+
| 8211f1d0f64f3269 | started | infra1 | http://127.0.0.1:12380 |  http://127.0.0.1:2379 |      false |
| 91bc3c398fb3c146 | started | infra2 | http://127.0.0.1:22380 | http://127.0.0.1:22379 |      false |
| fd422379fda50e48 | started | infra3 | http://127.0.0.1:32380 | http://127.0.0.1:32379 |      false |
```

## 测试失败容错

1. kill掉一个成员

`goreman run stop etcd2`

2. 存储一个key

`etcdctl put key hello`

3. 从停止的节点查询key

`etcdctl --endpoints=localhost:22379 get key`

4. 重新启动节点

`goreman run restart etcd2`


