# 别名

```
curl -X PUT "localhost:9200/trips?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "distance": {
        "type": "long"
      },
      "route_length_miles": {
        "type": "alias",
        "path": "distance" 
      },
      "transit_mode": {
        "type": "keyword"
      }
    }
  }
}
'
curl -X GET "localhost:9200/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "range" : {
      "route_length_miles" : {
        "gte" : 39
      }
    }
  }
}
'

```

> 别名字段的path必须是目标字段的完整路径