# token count

token_count字段类型，可以接受一个字符串值，进行分词，获得字符串中token的数量。

```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "name": { 
        "type": "text",
        "fields": {
          "length": { 
            "type":     "token_count",
            "analyzer": "standard"
          }
        }
      }
    }
  }
}
'
curl -X PUT "localhost:9200/my-index-000001/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{ "name": "John Smith" }
'
curl -X PUT "localhost:9200/my-index-000001/_doc/2?pretty" -H 'Content-Type: application/json' -d'
{ "name": "Rachel Alice Williams" }
'
curl -X GET "localhost:9200/my-index-000001/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "term": {
      "name.length": 3 
    }
  }
}
'

```