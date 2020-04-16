```shell
sudo groupadd docker     #添加docker用户组
sudo gpasswd -a $USER docker     #将登陆用户加入到docker用户组中
newgrp docker     #切换当前用户到docker组
docker ps    #测试docker命令是否可以使用sudo正常使用
```
