# 日期

JSON中没有日期类型，因此日期在es中可以表示为:

* 日期格式的字符串 ,如"2015-01-01" or "2015/01/01 12:10:30"
* 毫秒时间戳
* 秒时间戳

```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "date": {
        "type": "date" 
      }
    }
  }
}
'
curl -X PUT "localhost:9200/my-index-000001/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{ "date": "2015-01-01" }
'
curl -X PUT "localhost:9200/my-index-000001/_doc/2?pretty" -H 'Content-Type: application/json' -d'
{ "date": "2015-01-01T12:10:30Z" }
'
curl -X PUT "localhost:9200/my-index-000001/_doc/3?pretty" -H 'Content-Type: application/json' -d'
{ "date": 1420070400001 }
'
curl -X GET "localhost:9200/my-index-000001/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "sort": { "date": "asc"} 
}
'

```

## 指定多种日期格式

使用`||`指定多种日期格式。将依次尝试每个格式，直到找到匹配的格式。第一个匹配的格式将会被转为毫秒时间戳作为字符串存储。

```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "date": {
        "type":   "date",
        "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis"
      }
    }
  }
}
'

```



