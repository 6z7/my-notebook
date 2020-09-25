# fields

对同一个字段进行不同方式的索引。如，字符串可以映射成text类型进行全文分析，也可以作为keyword类型用于排序或聚合。

```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "city": {
        "type": "text",
        "fields": {
          "raw": { 
            "type":  "keyword"
          }
        }
      }
    }
  }
}
'
curl -X PUT "localhost:9200/my-index-000001/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{
  "city": "New York"
}
'
curl -X PUT "localhost:9200/my-index-000001/_doc/2?pretty" -H 'Content-Type: application/json' -d'
{
  "city": "York"
}
'
curl -X GET "localhost:9200/my-index-000001/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "city": "york" 
    }
  },
  "sort": {
    "city.raw": "asc" 
  },
  "aggs": {
    "Cities": {
      "terms": {
        "field": "city.raw" 
      }
    }
  }
}
'

```