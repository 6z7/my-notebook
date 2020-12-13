# match_phrase

查询字符串如果有多个词，分词后match_pharse需要文档中包含所有的词，match只要求文档中出现一个词既可。

```
curl -X GET "localhost:9200/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match_phrase": {
      "message": "this is a test"
    }
  }
}
'

```

参数：

slop：查询词条能够相隔多远时仍然将文档视为匹配，默认0，即分词需要紧邻