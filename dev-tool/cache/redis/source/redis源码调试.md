[redis源码地址](https://github.com/antirez/redis/releases)

## 搭建调试环境

1. 安装gcc、gdb
2. make noopt 

### 编译过程中关闭优化

make noopt 

#清理生成的文件

make distclean

docker运行redis

    docker run -it --rm -v c:\:/c --network my-net registry.cn-shanghai.aliyuncs.com/6z7/ubuntu:14.04.01 bash

常用gdb命令

# 设置命令参数

set args -h 127.0.0.1 -p 6379  

# gdb tui 图形窗口
layout

#关闭layout

ctr+x a

#跳出方法

finish

#调用堆栈信息

backtrace








