# keyword

keyword种类包括以下三种字段类型：

* keyword：结构化字符串

* constant_keywrod：索引中包含相同的值

* wildcard：通配符

> keyword与text都表示字符串类型，text可以分词，keword是一个整体


### wildcard类型

```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "my_wildcard": {
        "type": "wildcard"
      }
    }
  }
}
'
curl -X PUT "localhost:9200/my-index-000001/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{
  "my_wildcard" : "This string can be quite lengthy"
}
'
curl -X GET "localhost:9200/my-index-000001/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "wildcard": {
      "my_wildcard": {
        "value": "*quite*lengthy"
      }
    }
  }
}
'

```