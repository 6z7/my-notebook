# Match查询

查询匹配指定文字、数字、日期和bool值。在匹配查询前会进行分词

```
curl -X GET "localhost:9200/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "message": {  // 字段
        "query": "this is a test" // 字段匹配的值
      }
    }
  }
}
'

```

支持的参数:

query: 匹配的文本，必须

analyzer：分词器，可选

operator：用于解释查询值中的文本的布尔逻辑，支持
 
 * OR ： "capital of Hungary"会被解释为 capital OR of OR Hungary
 * AND ： "capital of Hungary"会被解释为 capital AND of AND Hungary




通过组合字段和查询参数，可以简化匹配查询语法。如:
```
curl -X GET "localhost:9200/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "message": "this is a test"
    }
  }
}
'

```

match查询的类型是布尔型。它意味着对提供的文本进行分析，分析过程从提供的文本构造一个布尔查询。`opertor`参数可以设置为OR或AND用来控制bool查询子句，默认OR。