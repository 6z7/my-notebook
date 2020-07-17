# 生成Go代码

>原文：https://developers.google.com/protocol-buffers/docs/reference/go-generated

这篇文章详细描述了protocol buffer编译器为任何给定protocol定义生成Go代码的过程。对proto2和proto3生成代码之间的任何差异都会突出显示出来。在阅读本文之前，应该先阅读proto2或[proto3指南](./proto3.md)。

## Compiler 

protocol buffer编译器生成Go代码需要一个插件。
```
go install google.golang.org/protobuf/cmd/protoc-gen-go
```

将会在`$GOBIN`目录下安装一个二进制文件`protoc-gen-go`。安装目录需要在$PATH中让protocol buffer编译器可以找到插件。当指定了--go_out标志时，编译器将生成Go代码到指定的目录。编译器为每一个.profo文件生成一个源码文件，生成的源码文件的名字是对应的profo文件名字加上扩展名.pb.go。profo文件中需要包含一个`go_package`选项指定Go包的完整路径。

```
option go_package = "example.com/foo/bar";
```

生成的文件所在的输出目录的子目录取决于`go_package`选项和编译时指定的命令行参数：

* 默认情况下，生成的文件放在以Go包的导入路径命名的目录中。如，protos/foo.proto文件中有上述的go_package选项，最终生成一个example.com/foo/bar/foo.pb.go文件

* 如果给protoc设置了`--go_opt=paths=source_relative`标志，输出文件与输入文件将放在相同的目录中，如protos/foo.proto文件生成的代码放在protos/foo.pb.go。

当你像这样运行proto编译器时：

```
protoc --proto_path=src --go_out=build/gen --go_opt=paths=source_relative src/foo.proto src/bar/baz.proto
```

编译器将读取src/foo.proto和src/bar/baz.proto文件，生成两个文件build/gen/foo.pb.go和build/gen/bar/baz.pb.go

如果需要，编译器会自动创建目录build/gen/bar，但不会创建build或build/gen，它们必须已经存在。

## Packages

.proto文件应该包含一个go_package选项，指定文件的完整的Go包路径。如果没有这个选项，编译器将尝试猜测一个。之后的编译器版本将要求go_package选项必填。The Go package name of generated code will be the last path component of the go_package option.


## Messages
