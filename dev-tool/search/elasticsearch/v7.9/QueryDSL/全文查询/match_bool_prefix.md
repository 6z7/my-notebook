# match_bool_prefix

返回包含查询词的文档，以最后一个词作为前缀进行匹配

```
curl -X GET "localhost:9200/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match_bool_prefix" : {
      "message" : "quick brown f"
    }
  }
}
'

```