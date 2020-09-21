# boolean

boolean字段接收true或fasle和可以解释为true或false的字符串

|       |                        |
| ----- | ---------------------- |
| False | false`, `"false"`, `"" |
| True  | true`, `"true"         |

```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "is_published": {
        "type": "boolean"
      }
    }
  }
}
'
curl -X POST "localhost:9200/my-index-000001/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{
  "is_published": "true" 
}
'
curl -X GET "localhost:9200/my-index-000001/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "term": {
      "is_published": true 
    }
  }
}
'

```

## 支持的参数

boost: 查询时提升权重，一个float类型的值，默认1.0

doc_values: 是否已column-stride格式存储到磁盘，默认false

index: 是否可以被检索  默认true

null_value:

store:

meta: 