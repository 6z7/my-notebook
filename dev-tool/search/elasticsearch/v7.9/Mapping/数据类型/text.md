# text

text类型的字段，可以进行全文索引。在创建索引时，会使用analyzer将字符串转为独立的terms。text字段不用于排序，也很少用于聚合。

```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "full_name": {
        "type":  "text"
      }
    }
  }
}
'

```

