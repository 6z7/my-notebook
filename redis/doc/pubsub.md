# 发布与订阅

`SUBSCRIBE`、`UNSUBSCRIBE`和`PUBLISH`命令协作实现了发布订阅消息系统。发布者发布消息到通道，订阅者订阅感兴趣的通道，而不同关系发布者或订阅者是谁，这种分离允许更大的可伸缩性和更动态的网络拓扑。

举个例子，订阅foo和bar通道
```
SUBSCRIBE foo bar
```

客户端发送消息到这些通道，redis将通知所有订阅这些通道的客户端。对订阅和取消订阅操作的回复以消息的形式发送，这样客户端就可以读取一致的消息流，其中第一个元素指示消息类型。在订阅客户端的上下文中只能使用`SUBSCRIBE`、`PSUBSCRIBE`、`UNSUBSCRIBE`、`PUNSUBSCRIBE`、`PING`和`QUIT`命令。

需要注意的是redis-cli客户端在订阅模式下不接受任何命令，仅仅能使用<kbd>Ctrl</kbd>+<kbd>C</kbd>退出订阅模式。

## 发布消息的格式





