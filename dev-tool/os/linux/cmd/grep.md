## grep

grep -Ev "^$|#" redis.conf

* -E
* -v  不包含匹配行数的所有行
* -c  匹配的行数
* -o  只显示匹配的行
* -i  忽略大小写
* -n  显示行号