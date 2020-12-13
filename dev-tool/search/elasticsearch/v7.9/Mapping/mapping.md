# Mapping

映射是定义一个文档包含那些字段、索引和存储的一个过程。

映射定义包含：

* 元数据字段：元数据字段用于自定义如何处理文档的关联元数据

* 字段：映射包含一系列与文档的相关的字段或属性，每个字段都有自己的数据类型

## 防止映射爆炸的设置

在索引中定以太多的字段会导致映射爆炸，可能会导致内存错误并难以恢复

通过下边的设置可以防止映射爆炸

* index.mapping.total_fields.limit

* index.mapping.depth.limit

* index.mapping.nested_fields.limit

* index.mapping.nested_objects.limit

* index.mapping.field_name_length.limit

## 动态映射

字段和映射类型不需要在使用之前进行定义。使用动态映射，新字段在索引文档时可以自动被添加，新字段的类型是映射中的顶级类型object和nested。

## 显示映射

在创建索引时指定字段和向以存在的索引中添加字段

### 创建索引时指定字段

```
curl -X PUT "localhost:9200/my-index-000001?pretty" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "age":    { "type": "integer" },  
      "email":  { "type": "keyword"  }, 
      "name":   { "type": "text"  }     
    }
  }
}
'
```

### 向已存在索引中添加字段

```
curl -X PUT "localhost:9200/my-index-000001/_mapping?pretty" -H 'Content-Type: application/json' -d'
{
  "properties": {
    "employee-id": {
      "type": "keyword",
      "index": false
    }
  }
}
'

```

>index:false表示该字段只被存储，不能被搜索

### 更新映射的字段

更新字段的名字会使已经使用该字段索引的数据无效，可以通过使用别名来替代重命名。

### 查看索引映射

```
curl -X GET "localhost:9200/my-index-000001/_mapping?pretty"
```

### 查看指定字段的映射

```
curl -X GET "localhost:9200/my-index-000001/_mapping/field/employee-id?pretty"
```

