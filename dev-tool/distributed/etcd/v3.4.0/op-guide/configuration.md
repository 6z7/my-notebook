# 配置

etcd可以通过配置文件、命令行参数和环境变量进行配置。

配置文件是一个YAML格式的文件。`--config-file`或环境变量`ETCD_CONFIG_FILE`指定文件。

命令行参数优先于环境变量。如果指定了配置文件，命令行参数和环境变量将被忽略。如，`etcd --config-file etcd.conf.yml.sample --data-dir /tmp`中的`--data-dir`选项将被忽略。

命令行参数对应的环境变量格式`-my-flag`为`ETCD_MY_FLAG`，该规则适用于所有命令行参数。


客户端使用2379与集群通信，集群节点间使用2380进行通信。

## Member flags

–name

* 节点成员的名称
* 默认：default
* 环境变量 ETCD_NAME
* 


–data-dir

* 数据保存目录
* 默认:${name}.etcd
* 环境变量：ETCD_DATA_DIR

–wal-dir

* wal日志专用目录，默认使用dataDir
* 默认值：""
* 环境变量：ETCD_WAL_DIR



## Clustering flags

`--initial-advertise-peer-urls`、 `--initial-cluster`、`--initial-cluster-state`和`--initial-cluster-token`仅在启动集群时会使用到。