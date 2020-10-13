# 查询和过滤上下文

## 相关性分数

默认情况下，Elasticsearch根据相关性评分对匹配的搜索结果进行排序，相关性评分衡量每个文档与查询的匹配程度。

相关性得分是一个正浮点数，在搜索API的_score元数据字段中返回。_score越高，文档越相关。虽然每种查询类型计算相关性分数的方法不同，分数计算还取决于查询子句是在查询上下文中运行还是在筛选上下文中运行。


## 查询上下文

在查询上下文中，一个查询子句回答了这个问题“这个文档与这个查询子句匹配得有多好?”除了决定文档是否匹配之外，查询子句还在_score元数据字段中计算相关性分数。

查询上下文在将查询子句传递给查询参数(如搜索API中的查询参数)时起作用。


## 筛选上下文

在筛选器上下文中，查询子句回答问题“这个文档与这个查询子句匹配吗?”答案是简单的“是”或“不是”——没有计算分数。过滤上下文主要用于过滤结构化数据

频繁使用的过滤器将被Elasticsearch自动缓存，以提高性能。

## 查询与筛选上下文例子

下面是在搜索API的查询和筛选上下文中使用查询子句的示例。此查询将匹配满足以下所有条件的文档:

title包含""&&content包含"Elasticsearch"&&status包含"published"&&publish_date日期大于等于"2015-01-01"

```
curl -X GET "localhost:9200/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": { 
    "bool": { 
      "must": [
        { "match": { "title":   "Search"        }},
        { "match": { "content": "Elasticsearch" }}
      ],
      "filter": [ 
        { "term":  { "status": "published" }},
        { "range": { "publish_date": { "gte": "2015-01-01" }}}
      ]
    }
  }
}
'

```

query参数标志查询上下文

在查询上下文中使用bool和两个match子句，这意味着它们用于对每个文档的匹配程度进行评分。

filter参数指示筛选器上下文。它的term和range子句用于筛选上下文。它们将过滤掉不匹配的文档，但不会影响匹配文档的得分。







