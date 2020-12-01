/proc/sys/net/ipv4/ 下的参数临时修改重启后无效

名称 | 默认值 | 描述 
-|-|-
tcp_keepalive_time | 7200 | TCP发送keepalive探测消息的间隔时间（秒），用于确认TCP连接是否有效。防止两边建立连接但不发送数据的攻击。 |
tcp_tw_reuse | 0 | 表示是否允许重新应用处于TIME-WAIT状态的socket用于新的TCP连接(这个对快速重启动某些服务,而启动后提示端口已经被使用的情形非常有帮助) |
tcp_tw_recycle | 0 | 打开快速 TIME-WAIT sockets 回收 |
tcp_wmem：min default max |4096<br>16384<br>131072 | 发送缓存设置<br>min：为TCP socket预留用于发送缓冲的内存最小值。每个tcp socket可以在建议以后都可以使用它。默认值为4096(4K)。<br/>default：为TCP socket预留用于发送缓冲的内存数量，默认情况下该值会影响其它协议使用的et.core.wmem_default 值，一般要低于net.core.wmem_default的值。默认值为16384(16K)。<br/>max: 用于TCP socket发送缓冲的内存最大值。该值不会影响net.core.wmem_max，"静态"选择参数SOSNDBUF则不受该值影响。默认值为131072(128K)。（对于服务器而言，增加这个参数的值对于发送数据很有帮助,在我的网络环境中,修改为了51200 131072 204800） |