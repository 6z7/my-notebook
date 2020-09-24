# Arrays

ES中，没有专门的array数据类型。默认，任何字段都能包含0个或多个值，在array中的值需要是相同的数据类型。如：

* string类型数组 ["one","two"]
* int类型数组  [1,2]
* array类型数组   [ 1, [ 2, 3 ]]，等价与[ 1, 2, 3 ]
* object类型数组   [ { "name": "Mary", "age": 12 }, { "name": "John", "age": 10 }]

>对于数组对象，不能单独查询其中的对象，如果需要，可以使用nested类型
 

 ```
 curl -X PUT "localhost:9200/my-index-000001/_doc/1?pretty" -H 'Content-Type: application/json' -d'
{
  "message": "some arrays in this document...",
  "tags":  [ "elasticsearch", "wow" ], 
  "lists": [ 
    {
      "name": "prog_list",
      "description": "programming list"
    },
    {
      "name": "cool_list",
      "description": "cool stuff list"
    }
  ]
}
'
curl -X PUT "localhost:9200/my-index-000001/_doc/2?pretty" -H 'Content-Type: application/json' -d'
{
  "message": "no arrays in this document...",
  "tags":  "elasticsearch",
  "lists": {
    "name": "prog_list",
    "description": "programming list"
  }
}
'
curl -X GET "localhost:9200/my-index-000001/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "tags": "elasticsearch" 
    }
  }
}
'

 ```