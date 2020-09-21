## binary

二进制类型接受Base64编码的二进制值。这种类型的字段默认不能进行排序和被检索的

```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "name": {
        "type": "text"
      },
      "blob": {
        "type": "binary"
      }
    }
  }
}
'

curl -X PUT "localhost:9200/my-index-000001/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{
  "name": "Some binary blob",
  "blob": "U29tZSBiaW5hcnkgYmxvYg==" 
}
'
```

## binary字段类型支持的参数

* doc_values:字段是否存以column-stride格式储到磁盘上，用于之后的排序、聚合和脚本，默认false

* store: 是否字段值应该被存储，用于单独检索，默认fasle

