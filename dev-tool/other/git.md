# git换行符问题

windows平台换行符使用的是`\r\n`,linux使用的是`\n`。

## windows平台

`git config --global core.autocrlf true`

当签出代码时将lf转为crlf，签入时将crlf转为lf。所以仓库中保存的文件都是以lf进行换行。但是，如果在设置之前将crlf文件提交到仓库，那么在设置之后文件也不会被修改为lf。

>开启autocrlf对包含混合换行符的文件，在签出时文件会变成已修改状态。这是因为autocrlf会将lf转为crlf，签出后与原文件对比发现发生了变化(有时也包含混合换行符但没报错??)。


`git config --global core.safecrlf true`

当设置了autocrlf为true后，如果文件中同时包含lf，在提交时会报错(混合反而不报错??)，需要将文件的换行符修改为crlf后才能保存。

 
 ## 指定文件夹使用单独的配置

 在.gitconfig中配置

 ```
 [includeIf "gitdir:C:/senki/work/"]
    path = ~/.gitconfig_work   #指定文件夹的项目使用单独的配置,gitdir大小写敏感，可以使用gitdir/i忽略大小写
 ```

 ## 全局ignore配置

 ```
 [core]	 
	# 全局ignore配置
	excludesfile = ~/.gitignore_global
 ```


 ## 查看文件匹配到的忽略规则

 ```
 git check-ignore -v  "C:\\senki\\work\\LastPackageCo.cs"
 ```