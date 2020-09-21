# doc-values

大多数字段默认都能被索引，这样它们就能被搜索到。通过倒排索引可找到搜索词所在的文档。

排序、聚合和访问脚本中的字段值需要不同的数据访问模式。这种模式和通过在倒排索引中查找term找到文档不同，我们需要查找文档找到其中包含的term。

doc values为保存在磁盘上的数结构，在文档建索引时被创建。doc values存储的值和_source中的一致。几乎所有字段类型都支持doc values,除了text和annotated_text类型。

所有的字段默认都支持doc values。如果确认不需要对字段进行排序或聚合或通过脚本访问，则可以禁用doc values来节约磁盘空间。

```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "status_code": { 
        "type":       "keyword"
      },
      "session_id": { 
        "type":       "keyword",
        "doc_values": false
      }
    }
  }
}
'
```

> status_code默认使用doc_values
> session_id禁用doc_values，但是可以被检索到