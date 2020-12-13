# _index

在跨多个索引执行查询时，有时需要添加只与某些索引的文档关联的查询子句。_index字段允许在索引中匹配文档。它的值可在某些查询和聚合中访问，在排序或编写脚本时:

```
curl -X PUT "localhost:9200/index_1/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{
  "text": "Document in index 1"
}
'
curl -X PUT "localhost:9200/index_2/_doc/2?refresh=true&pretty" -H 'Content-Type: application/json' -d'
{
  "text": "Document in index 2"
}
'
curl -X GET "localhost:9200/index_1,index_2/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "terms": {
      "_index": ["index_1", "index_2"] 
    }
  },
  "aggs": {
    "indices": {
      "terms": {
        "field": "_index", 
        "size": 10
      }
    }
  },
  "sort": [
    {
      "_index": { 
        "order": "asc"
      }
    }
  ],
  "script_fields": {
    "index_name": {
      "script": {
        "lang": "painless",
        "source": "doc[\u0027_index\u0027]" 
      }
    }
  }
}
'

```

_index字段是一个虚拟字段，它不会作为一个真实字段添加到Lucene索引中。