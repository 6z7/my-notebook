## Go项目结构

[这是一个社区规范](https://github.com/golang-standards/project-layout)

## /cmd

项目的main应用程序。应用的名字应该和目录的名字一致(如:/cmd/myapp/main.go)。不要在这个目录放太多代码。

## /internal

应用私有的代码，此目录下的代码不能被其它应用使用，这是Go编译器强制执行的。如果项目中有多个应用程序，多个应用共用的私有代码可以放在`/internal/pkg`目录，每个应用特有的私有代码可以放在`/internal/app`目录下

## /pkg

其它应用可以使用的代码放到这个目录下

## /vendor

项目的依赖，使用了modules功能后，不在需要

## /api

协议文件

## /web

web应用的静态资源

## /configs

配置文件

## init

系统初始化和进程启动配置

## /scripts

编译、安装、分析等脚本

## /build

持续集成目录,云 (AMI), 容器 (Docker), 操作系统 (deb, rpm, pkg)等的包配置和脚本   

## /deployment

IaaS、Paas和容器编排部署配置和模板

## /test

测试目录

## /doc

设计和使用文档

## /tools

项目的支持工具和脚步

## /examples

事例

## /third_party

## /githooks

git hooks

## /assets

项目的静态资源(如:图片、logo等)

## /website

项目的web站点