# _routing

使用以下公式将文档路由到特定分片:

```
shard_num = hash(_routing) % num_primary_shards
```

默认_routing使用的是文档的_id

可以通过为每个文档指定自定义路由值来实现自定义路由模式。例如:

```
curl -X PUT "localhost:9200/my-index-000001/_doc/1?routing=user1&refresh=true&pretty" -H 'Content-Type: application/json' -d'
{
  "title": "This is a document"
}
'
curl -X GET "localhost:9200/my-index-000001/_doc/1?routing=user1&pretty"

```

在查询中访问指定_routing字段的值:
```
curl -X GET "localhost:9200/my-index-000001/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "terms": {
      "_routing": [ "user1" ] 
    }
  }
}
'

```

自定义路由可以减少搜索的影响。不必将搜索请求扇出到索引中的所有分片，请求可以只发送到与特定路由值匹配的分片:
```
curl -X GET "localhost:9200/my-index-000001/_search?routing=user1,user2&pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "title": "document"
    }
  }
}
'

```

将自定义路由字段配置成必选值
```
curl -X PUT "localhost:9200/my-index-000002?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "_routing": {
      "required": true 
    }
  }
}
'
curl -X PUT "localhost:9200/my-index-000002/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{
  "text": "No routing value provided"
}
'

```


当索引文档指定自定义_routing时，不能保证索引中的所有分片都具有唯一的_id。实际上，如果使用不同的_routing值进行索引，那么具有相同_id的文档可能会位于不同的分片上。

自定义路由时需要由用户来确保id在索引中是唯一的。


可以配置一个索引，以便自定义路由值转到一组分片，而不是单个分片。这有助于降低集群不平衡的风险，同时降低搜索的影响。

通过在创建索引时设置` index.routing_partition_size`可以影响路由分片的情况。值越大分布的越均匀，代价是每个请求必须搜索更多的分片。

```
shard_num = (hash(_routing) + hash(_id) % routing_partition_size) % num_primary_shards

```

启用这个功能，需要index.routing_partition_size的值大于1且小于index.number_of_shards







