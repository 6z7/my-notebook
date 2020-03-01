# Go中常用的命令

* go build
* go install
* go clean
* go env
* go fmt
* go generate
* go get
* go list
* go run
* go test
* go bug
* go doc
* go fix
* go version
* go vet
* go tool
* go mod

## go build

编译包及其依赖项，但是不会安装包 

`go build [-o output] [-i] [build flags] [packages]` 

编译包时将忽略其中的`_test.go`结尾的测试文件  

当编译可执行文件时时，可执行文件将以第一个文件的名字命名。
`go build app.go xx.go`的可执行文件名app或app.exe；当编译可执行包时，将以包的名字命名。`go build mycmd/cmd`输出结果为cmd或cmd.exe

当同时编译多个包或编译一个非main包时，仅仅是校验是否能编译，将不会有任何输出。`go build mycmd/cmd mycmd/demo`与`go build mycmd/demo`

`-o` 取代默认行为将编译结果输出到指定文件或目录。
```go
//保存到文件 aa.exe
go build -o ./aa/aa.exe mycmd/cmd
//保存到文件  bb.exe
go build -o ./aa/bb mycmd/cmd
//如果bb目标不存在 则报错
go build -o ./aa/bb/ mycmd/cmd
//cc目录存在  输出到文件cmd.exe
go build -o ./aa/cc/ mycmd/cmd
```

`-i` 安装依赖的包


build flags是一些可以被`build`、`clean`、`get`、`install`、`list`、`run`、`test`通用的标志

| build flag                       | 说明                                                         |
| -------------------------------- | ------------------------------------------------------------ |
| -a                               | 强制重新生成包                                               |
| -n                               | 仅打印需要用到的命令，但是不运行它们                         |
| -p n                             | 编译或测试时的并发数量，默认是机器的CPU核数                  |
| -race                            | 允许竞争检测，仅这些平台支持 linux/amd64, freebsd/amd64, darwin/amd64和windows/amd64 |
| -msan                            | enable interoperation with memory sanitizer                  |
| -v                               | 打印出被编译的包名                                           |
| -work                            | 打印临时工作目录，在退出时不删除                             |
| -x                               | 打印用到的命令，并运行                                       |
| -asmflags  '[pattern=]arg list'  | 传递给go tool asm用到的参数                                  |
| -buildmode mode                  | 编译模式 参见 go help buildmode                              |
| -compiler name                   | 使用的编译器名称(gccgo or gc)                                |
| -gccgoflags '[pattern=]arg list' | 传递给gccgo编译/链接器参数                                   |
| -gcflags '[pattern=]arg list'    | 传递给go编译器参数                                           |
| -installsuffix suffix            | a suffix to use in the name of the package installation directory, in order to keep output separate from default builds. 		If using the -race flag, the install suffix is automatically set to race or, if set explicitly, has _race appended to it. Likewise for the -msan flag. Using a -buildmode option that requires non-default compile flags has a similar effect. |
| -ldflags '[pattern=]arg list'    | 传递给go tool link的参数                                     |
| -linkshared                      | link against shared libraries previously created with -buildmode=shared. |
| -mod mode                        | module download mode to use: readonly or vendor.	See 'go help modules' for more. |
| -pkgdir dir                      | 从指定位置而不是约定位置安装和加载包                         |
| -tags tag,list                   | 逗号分割的编译时用到的tag用于条件编译                        |
| -trimpath                        | remove all file system paths from the resulting executable. Instead of absolute file system paths, the recorded file names will begin with either "go" (for the standard library), or a module path@version (when using modules), or a plain import path (when using GOPATH). |
| -toolexec 'cmd args'             | a program to use to invoke toolchain programs like vet and asm. For example, instead of running asm, the go command will run 'cmd args /path/to/asm \
<arguments for asm>'. |

`-asmflags`, `-gccgoflags`, `-gcflags`和`-ldflags`标志接受空格分割的多个参数，但需要用单引号或双引号包裹。参数列表前可能有一个模式`[pattern=]`,它限制哪些包可以使用这些参数。如果没有模式则只有命令行上声明的包可以使用这些参数。如 `go build -gcflags=-S fmt` 仅fmt包打印反汇编信息,`go build -gcflags=all=-S fmt` fmt和所有它的依赖打印反汇编信息

## go install

编译并安装包

`go install [-i] [build flags] [packages]`

可执行包被安装到GOBIN环境变量下，如果没有设置GOBIN则默认安装到$GOPATH/bin或$HOME/go/bin

当mod功能被禁用时，其它包(非可执行)被安装到$GOPATH/pkg/$GOOS_$GOARCH；当开启mod功能时，其它包被编译和缓存，但不会被安装

`-i`:安装包的依赖项

## go clean

`go clean [clean flags] [build flags] [packages]` 

`-i`:清除安装的包和可执行文件(go install生成的，包括当前所在目录和安装目录)

`-r`:递归清理所有的包依赖项(仅清理当前命令所在目录)

`-n`:仅仅打印将要执行的清理命令

`-x`:清理同时打印清理的命令(仅清理当前命令所在目录)

`-cache`:清理go build的缓存

`-testcache`:使go build cache中的测试结果过期

`-modcache`:清理mod缓存

## go env

打印Go环境变量信息

`go env [-json] [-u] [-w] [var ...]`

`-json`:以json格式输出

`-u`: 将变量重置为默认值

`-w`: 设置变量,格式name=value

设置的信息保存在`GOENV`对应的目录下，windows默认C:\Users\用户名\AppData\Roaming\go\env;linux默认/home/用户/.config/go/env

## go fmt

格式化代码

`go fmt [-n] [-x] [packages]`

`-n`:仅打印要执行的命令

`-x`:打印要执行的命令并执行

## go generate

运行文件中描述的命令，这些命令可以是任何进程，目的是创建或更新Go源文件。

Go generate不会自动被go build、go test、go get等命令调用执行，需要手动执行。

Go generate扫描文件中如下所示形式的指令`//go:generate command argument...`。//go之间没有空格，要执行的命令需要在可执行路径中(shell path)或指定完整的路径(/usr/you/bin/mytool)

匹配指定的正则:`^// Code generated .* DO NOT EDIT\.$`

Go generate可以出现在文件的任何位置，为了容易被发现，通常放在靠近开头的位置

指令的参数使用空格进行分割，或双引号包裹的字符串作为整体传递给指令

Go generate用到的几个环境变量:

* $GOARCH  处理器
* $GOOS    操作系统
* $GOFILE  文件名
* $GOLINE  指令所在的文件的函数
* $GOPACKAGE 包含指令的文件所在的包
* $DOLLAR

`go generate [-run regexp] [-n] [-v] [-x] [build flags] [file.go... | packages]`

`-run`:指定从文件中匹配出指令的正则

`-v`:打印出正在处理的包和文件名

`-n`:仅打印将要指定的命令

`-x`:打印要执行的命令并执行

## go get

下载并编译安装包

`go get [-d] [-f] [-t] [-u] [-v] [-fix] 
[-insecure] [build flags] [packages]`

`-d`:仅下载包不要安装

`-u`:更新包和依赖，默认情况下仅下载缺失的包，不会更新已经存在的包

`-f`:只有设置`-u`才会生效，它强制`get -u`不验证每个包都从远程仓库中更新。如果源是一个本地仓库时会用到

`-fix`:在解析依赖项和生成代码之前，先在下载的包上运行修复工具

`-insecure`:允许使用http连接远程仓库

`-t`:下载包编译所需的测试

`-v`:详细输出

代码编译时的签出规则:使用匹配本地版本的分支或tag，如果没有则使用默认分支

Get不会更新`vendor`中的目录

## go list

列出包的详细信息

`go list [-f format] [-json] [-m] [list flags] [build flags] [packages]`

`-f '{{.ImportPath}}'`:指定输出，根据是否指定模块(-m)可选择选项不同

包:

* Dir    包含包的源码目录
* ImportPath  
* Name
* ...

模块:

* Path
* Version
* ...

`-json`:以josn格式打印包信息

`-m`:列出模块信息而不是包信息 `go list -m all`

`-u`: 如果有的话，展示可用的升级信息 `go list -m -u -json all`

## go run

编译并运行

`go run [build flags] [-exec xprog] package [arguments...]`

默认情况下，直接运行编译好的可执行文件`a.out arguments...`，如果指定了`-exec`，则运行`xprog a.out arguments...`

## go test

自动测试包

`go test [build/test flags] [packages] [build/test flags & test binary flags]`

`-args`:

`-c`:编译测试文件保存到包名.test文件中，不会运行

`-exec xprog`:运行编译后的文件`xprong 二进制文件 参数...`

`-i`:安装测试需要的包，但不允许

`-json`:以json格式输出

`-o`:指定编译文件的名称和位置，会运行文件除非指定了`-c`或`-i`

`-v`:

`-bench regexp`:运行满足条件的的基准测试，默认情况下没有基准测试会执行；运行所有的基准测试`-bench .`或`-bench=.` 

`-benchtime t`:基准运行指定的时间(单位:time.Duration,如: -benchtime 1h30s)，默认是1s;有一种特殊的语法`Nx`意味着运行基准N次(如: -benchtime 100x)

`-count n`:运行测试和基准测试n次(默认1);如果设置了`-cpu`，则每个cpu运行n次

`-cover`:启动代码覆盖率分析

`-covermode set,count,atomic`:代码覆盖率分析方式

`-cpu 1,2,4`:设置GOMAXPROCS，默认是当前GOMAXPROCS

`-failfast`:第一个测试失败就终止

`-parallel n`:并行数量



以`_.go`和`_test.go`开头的文件会被编译器忽略

`testdata`目录用于保存测试用的辅助数据，会被编译器忽略

`go test`运行时会先运行`go vet`，使用`-vet=off`可禁用

Go test有2种不同的运行模式:

* local directory mode(本地目录模式)

  在当前目录直接运行无包参数的`go test`或`go test -v`，会运行目录下的所有测试实例。这种模式下不会缓存测试结果

* package list model(包列表模式)

  当显示指定包参数(如 `go test main`、`go test ./..`和`go test .`)会运行在此模式下。缓存测试通过的包避免重复执行，当go test运行时发现有之前成功的缓冲结果在，会直接展示之前的结果而不是再次运行


## go bug

打开默认浏览器在github上报告bug,并会携带一些系统信息

## go doc

`go doc [-u] [-c] [package|[package.]symbol[.methodOrField]]`

查看文档

## go fix

`go fix [packages]`

修复包中的旧版本的语法

## go version

`go version [-m] [-v] [file ...]`

查询版本信息，如果没有指定文件或目录则返回的是当前安装的go版本信息。

`-v`:输出无法识别的文件

`-m`:输出可执行文件中的模块信息

## go vet

分析并报告代码存在的问题

`go vet [-n] [-x] [-vettool prog] [build flags] [vet flags] [packages]`

`-n`:仅打印要执行的命令

`-x`:打印要执行的命令并执行

`-vettool=prog`:设置分析工具

## go tool

`go tool [-n] command [args...]`

查看命令文档 `go doc cmd/<command>`

常用命令:

### addr2line

`go tool addr2line binary`

将调用栈地址转为文件和行号，只有在pprof中用到，未来可能会移除

### api

`go tool api`

### asm

`go tool asm [flags] file`

编译指定的汇编文件

`-S`:打印汇编和机器码

`-V`:打印汇编器版本

`-o file`:将输出保存到指定文件(.o文件)

`go tool buildid [-w] file`

打印文件中的build id(文件内容hash)

`-w`:根据文件重新计算build id并更新到文件


### cgo

`go tool cgo [cgo options] [-- compiler options] gofiles...`
 

### compile

`go tool compile [flags] file...`

编译Go文件，这些文件必须属于同一个包

`-D path`:设置文件的相对路径

`-N`:禁止优化

`-S`:打印汇编到标准输出

`-V`:仅打印编译器版本

`-asmhdr file`:

`-L`:在错误信息中展示文件路径

`-l`:禁止内联

`-lang version`:设置Go语言版本(如:-lang=go1.12)

`-o file`:保存编译成的链接文件(默认.o文件，如果使用了-pack则是.a文件)

`-pack`:编译成一个静态库文件而不是一个链接文件

`-shared`:动态链接库

Go源文件中的编译器指令

`//go:noescape`   
`//go:nosplit`  
`//go:linkname localname [importpath.name]`

### cover

`go tool cover -help`

代码覆盖情况分析工具

### dist

`go tool dist [command]`

### doc

`go doc`

查看文档

### fix

`go tool fix [-r name,...] [path ...]`

修复包中的旧版本的语法

### link

`go tool link [flags] main.a`

链接器


### nm

`go tool nm [options] file...`

列出文件中使用的符号信息


### objdump

`go tool objdump [-s symregexp] binary`

反汇编，输出所有的符号

`-s symregexp`:仅输出满足条件的符号

### pack

`go tool pack op file.a [name...]`

创建静态链接库(.a)

### pprof

`go tool pprof binary profile`

性能分析

### test2json

`go tool test2json [-p pkg] [-t] [./pkg.test -test.v]`

将go test的输出转为json

`-p pkg`:

`-t`:追加时间到输出结果中


### trace

### vet

报告代码存在的问题
 

## go mod

`go mod <command> [arguments]`

Go模块管理

`go help mod <command>`查看详细信息

command:

### download

下载包到本地缓存  
`go mod download [-json] [modules]`



### edit

  编辑go.mod

### graph

  打印依赖的模块

### init

初始化模块   
`go mod init [module]`  

### tidy

添加缺失或移除无用的模块

### vendor

将依赖项复制到`vendor`文件夹中

### verify

验证依赖是否修改过

### why

解释为什么包或模块被需要