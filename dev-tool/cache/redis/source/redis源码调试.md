[redis源码地址](https://github.com/redis/redis)

## 搭建调试环境

1. 安装gcc、gdb
2. make noopt 

// 编译过程中关闭优化

make noopt 

// 清理生成的文件

make distclean

## vscode调试

1. 安装C/C++插件
2. 配置调试,[参见](https://gitee.com/glzsk/redis/commit/ffa6c43538744ed9c20ae920f8a620a21e3a5780)

## CMake编译

cmake编译配置，[参见](https://gitee.com/glzsk/redis/commit/e36cb6cf10a6c3885864a4b87d1bc1960222f5f7)

## 常用gdb命令

// 设置命令参数

set args -h 127.0.0.1 -p 6379  

// gdb tui 图形窗口

layout

// 关闭layout

ctr+x a

// 跳出方法

finish

// 调用堆栈信息

backtrace








