## go modules 

[Modules](https://github.com/golang/go/wiki/Modules)

GO111MODULE=auto, 如果发现go.mod文件则启用module模式，即使在GOPATH目录下。在Go 1.13之前，如果在GOAPTH目录下，不会自动启动module模式

 GO111MODULE=on 启动module功能

go get -u ./... or 升级所有的包到最新

get -u=patch ./... 

go build ./...  编译所有的包

go test ./...  测试所有包

go mod tidy 移除不需要的依赖，安装缺失的依赖

go mod vendor   创建一个vendor文件夹

go.mod文件中有4个指令:

* module  ：包路径

* require   ：

* replace

* exclude


require M v1.2.3   表示可以安装   v1.2.3<=M版本< v2 范围内的版本，v2版本和v1不兼容

选择依赖的最小版本号


语义导入版本控制，主版本号被包含到包路径中

如果新的包和旧的包有相同的包路径，新包必须向后兼容旧的包

为了实现语义导入版本控制，使用module功能的包必须符合以下规则:

* 满足 semver语义的规则(如，v1.2.3)
* 如果包是v2或更高的版本，主版本号/vN必须包含到包路径的结尾(如，require github.com/my/mod/v2 v2.0.1，go get github.com/my/mod/v2@v2.0.1)
* 如果包的版本是v0或v1，则不需要在包的路径中包含主版本号


使用go get进行包的升降级，也可以直接编译go.mod文件进行修改

go get example.com/package    升级依赖到最新版本

go get -u example.com/package  升级依赖和依赖的依赖到最新版本


go list -u -m all  查看直接依赖和间接依赖可用的次要升级和修补程序升级


go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' -m all 2> /dev/null    仅查看直接依赖可用的次要升级和修补程序升级


go get -u ./...    升级到最新的次要版本或修补(-t 也升级测试的依赖)

go get -u=patch ./...    升级到最新的修补(-t 也升级测试的依赖)



go get foo与go get foo@latest等价 升级到最新版本


go get foo@v1.6.2, go get foo@e3702bed2, or go get foo@'<v1.6.2'

go get foo@master 使用分支名称，获取最新的提交，而不管是否有版本号
 


