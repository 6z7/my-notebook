## 端口转发

目前发现不支持https?
netsh interface portproxy show all
netsh interface portproxy reset

```
--80端口转发到本地8086
netsh interface portproxy add v4tov4 listenport=80 listenaddress=m.benlai.com connectport=8086 connectaddress=127.0.0.1

netsh interface portproxy add v4tov4 listenport=80 listenaddress=0.0.0.0 connectport=8086 connectaddress=127.0.0.1

netsh interface portproxy add v4tov4 listenport=80 listenaddress=127.0.0.1 connectport=8086 connectaddress=127.0.0.1

netsh interface portproxy add v4tov4 listenport=80 listenaddress=* connectport=8086 connectaddress=127.0.0.1
```

kill所有nginx线程

`taskkill /F /IM nginx.*`

```
netstat -ano|findstr 80

tasklist|findstr pid

taskkill /F /PID pid
```

子系统文件路径

C:\Users\\{用户名}\AppData\Local\Packages\CanonicalGroupLimited.UbuntuonWindows_79rhkp1fndgsc\LocalState\rootfs

```
// 永久用户环境变量
setx env_name env_value

setx "path" "D:\test;%path%"

// 永久系统环境变量
setx env_name env_value /m

setx "Path" "%cd%;%path%" /m

// 删除永久环境变量
REG delete "HKCU\Environment" /F /V "abc"
```