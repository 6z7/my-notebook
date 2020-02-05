处于SYN_RECV半连接状态的TCP最大连接数，当处于SYN_RECV状态的TCP连接数超过tcp_max_syn_backlog后，
会丢弃后续的SYN报文。

![](./../../../../image/backlog.png)