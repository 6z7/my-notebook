# Query DSL

Elasticsearch提供了一个完整的基于JSON的查询DSL(领域特定语言)来定义查询。可以将查询DSL看作查询的AST(抽象语法树)，它由两种类型的子句组成:

## 叶子查询

叶查询子句在特定字段中查找特定值，如match、term或range查询

## 复合查询

复合查询子句包装其它叶查询或复合查询，并以逻辑方式组合多个查询(如bool或dis_max查询)，或用于改变它们的行为(如constant_score查询)。