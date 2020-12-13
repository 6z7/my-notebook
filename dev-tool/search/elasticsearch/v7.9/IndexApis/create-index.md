# Create index API

`PUT /<index>`

```
curl -X PUT "localhost:9200/my-index-000001?pretty"

```

创建index时可以指定：

* index的设置
* index中字段的映射
* index别名

## Path参数:

<index> 索引名称，必需，名称需要满足以下条件：

* 小写
* 不能包含\, /, *, ?, ", <, >, |, ` ` (space character), ,, #
* 7.0+之后不在支持冒号(:)
* 不能以 -, _, + 开头
* 不能是 .或..
* 不能超过255字节
* 不能以.开头，除了隐藏索引和插件管理的内部索引

## Query参数


wait_for_active_shards：需要等待N个分片同步完成，默认1，只写入主分片即可，最大number_of_replicas+1

master_timeout：连接主节点超时时间，默认30s

timeout：指定未响应超时时间，默认30s

## Request Body

aliases：索引名称，可选

mappings：索引中字段映射，可选，包括
   - 字段名
   - 字段数据类型
   - 映射参数

settings：索引设置

## Examples

设置索引
```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "settings": {
    "index": {
      "number_of_shards": 3,  
      "number_of_replicas": 2 
    }
  }
}
'
// 可以简化为

curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "settings": {
    "number_of_shards": 3,
    "number_of_replicas": 2
  }
}
'

```

映射

```
curl -X PUT "localhost:9200/test?pretty" -H 'Content-Type: application/json' -d'
{
  "settings": {
    "number_of_shards": 1
  },
  "mappings": {
    "properties": {
      "field1": { "type": "text" }
    }
  }
}
'

```

别名

```
curl -X PUT "localhost:9200/test?pretty" -H 'Content-Type: application/json' -d'
{
  "aliases": {
    "alias_1": {},
    "alias_2": {
      "filter": {
        "term": { "user.id": "kimchy" }
      },
      "routing": "shard-1"
    }
  }
}
'

```

默认情况下，创建索引后当所有分片的主副本处理完成后或超时才返回响应。
```
{
  "acknowledged": true,
  "shards_acknowledged": true,
  "index": "test"
}
```

acknowledged：索引是否成功创建

shards_acknowledged：是否在超时之前为索引中的每个分片启动了所需数量的分片副本

```
curl -X PUT "localhost:9200/test?pretty" -H 'Content-Type: application/json' -d'
{
  "settings": {
    "index.write.wait_for_active_shards": "2"
  }
}
'

```

