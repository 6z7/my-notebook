# _source

`_source`字段包含创建索引时传入的原始的JSON定义。_source字段本身不能被索引，但是存储了该字段，因此在查询操作时可以返回。

虽然非常方便，但_source字段在索引中确实会产生存储开销。因此可以禁用
```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "_source": {
      "enabled": false
    }
  }
}
'
```

> 禁用_source字段会导致一系列功能不能使用，操作时需要注意




