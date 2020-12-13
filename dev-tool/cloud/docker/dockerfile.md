FROM

RUN  运行在工作目录

ARG 定义一些宏，在run中使用

ENV 设置环境变量

COPY 复制本地文件到镜像

ADD  添加本地文件到镜像，对压缩文件进行提取和解压

WORKDIR 工作目录

ENTRYPOINT  镜像执行入口 ["可执行文件","参数1","参数2"]  工作目录

CMD  容器启动时执行的命令

cmd与entrypoint的区别:

a. cmd命令会被docker run命令覆盖entrypoint不会，entrypoint会把docker run的输入作为参数传入  
b. cmd与entrypoint都存在时，cmd会作为entrypoint的参数

FROM ..AS  镜像别名，后边直接使用别名

MAINTAINER  维护人员

EXPOSE	暴露端口

ONBUILD 触发操作

USER  以什么用户执行

VOLUME
