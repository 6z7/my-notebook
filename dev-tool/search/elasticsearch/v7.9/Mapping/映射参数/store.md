# store

默认，字段值被索引可以被搜索到，但是不会被存储。这意味着这个字段可以被查询，但是无法检索原始的字段值。

通常这并不重要，因为字段已经是_source字段的一部分，默认是会被存储的。如果只是检索时不想看到_source，则可是使用source过滤功能。

在某些情况下，存储字段是有意义的。For instance, if you have a document with a title, a date, and a very large content field, you may want to retrieve just the title and the date without having to extract those fields from a large _source field:

```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "title": {
        "type": "text",
        "store": true 
      },
      "date": {
        "type": "date",
        "store": true 
      },
      "content": {
        "type": "text"
      }
    }
  }
}
'
curl -X PUT "localhost:9200/my-index-000001/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{
  "title":   "Some short title",
  "date":    "2015-01-01",
  "content": "A very long content field..."
}
'
curl -X GET "localhost:9200/my-index-000001/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "stored_fields": [ "title", "date" ] 
}
'

```

