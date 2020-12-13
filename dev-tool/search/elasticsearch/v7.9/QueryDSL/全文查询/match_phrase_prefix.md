# match_phrase_prefix

返回包含查询所有词的文档，最后一个词作为一个前缀，匹配任何以其开头的词

```
curl -X GET "localhost:9200/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match_phrase_prefix": {
      "message": {
        "query": "quick brown f"
      }
    }
  }
}
'

```