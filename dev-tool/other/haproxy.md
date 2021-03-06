
[haproxy中的计时事件](http://cbonte.github.io/haproxy-dconv/2.2/configuration.html#8.4)


计时器在排查网络问题方面提供了很大的帮助。使用的单位是毫秒。

http层代理的时间事件：

```
               first request               2nd request
      |<-------------------------------->|<-------------- ...
      t         tr                       t    tr ...
   ---|----|----|----|----|----|----|----|----|--
      : Th   Ti   TR   Tw   Tc   Tr   Td : Ti   ...
      :<---- Tq ---->:                   :
      :<-------------- Tt -------------->:
                :<--------- Ta --------->:
```

tcp层代理的时间事件:

```
       TCP session
      |<----------------->|
      t                   t
   ---|----|----|----|----|---
      | Th   Tw   Tc   Td |
      |<------ Tt ------->|
```

Th：完成握手建立连接的时间

Ti：握手之后到收到第一个http请求的字节的空闲时间，-1代表连接上没有收到请求数据

TR：收到第一个字节到接收http头完成耗费的时间，-1说明没有收到htpp头结束的标记

Tq：距离从接收到数据或上一个输出的最后一个字节开始到获取到http头完成的时间。Tq=Th + Ti + TR

Tw：队列中等待能去建立连接花费的时间 -1说明请求在到达队列之前被结束

Tc：与后端服务建立连接花费的时间  -1说明未建立连接

Tr：与后端建立连接到后端发送完消息头到ha的时间

Ta：从接收到http第一个字节到后端响应完成返回到ha的时间

Td：后端传输数据到ha的时间 Td=Ta - (TR + Tw + Tc + Tr)

Tt：从接收到数据到后端处理完成返回ha的时间


一些字段说明:

srv_conn：记录日志时，当前服务上的活跃的并发连接数

