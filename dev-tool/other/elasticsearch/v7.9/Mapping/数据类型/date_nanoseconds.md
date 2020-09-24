# date nanoseconds

date数据类型精确到毫秒，date_nanos精确到纳秒。date_nanos限制日期范围在1970-2262

```
curl -X PUT "localhost:9200/my-index-000001?include_type_name=true&pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "_doc": {
      "properties": {
        "date": {
          "type": "date_nanos" 
        }
      }
    }
  }
}
'
curl -X PUT "localhost:9200/my-index-000001/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{ "date": "2015-01-01" }
'
curl -X PUT "localhost:9200/my-index-000001/_doc/2?pretty" -H 'Content-Type: application/json' -d'
{ "date": "2015-01-01T12:10:30.123456789Z" }
'
curl -X PUT "localhost:9200/my-index-000001/_doc/3?pretty" -H 'Content-Type: application/json' -d'
{ "date": 1420070400 }
'
curl -X GET "localhost:9200/my-index-000001/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "sort": { "date": "asc"} 
}
'
curl -X GET "localhost:9200/my-index-000001/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "script_fields" : {
    "my_field" : {
      "script" : {
        "lang" : "painless",
        "source" : "doc[\u0027date\u0027].value.nano" 
      }
    }
  }
}
'
curl -X GET "localhost:9200/my-index-000001/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "docvalue_fields" : [
    {
      "field" : "date",
      "format": "strict_date_time" 
    }
  ]
}
'

```