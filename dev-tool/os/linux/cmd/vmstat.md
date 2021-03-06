## vmstat

`vmstat`查看系统负载详细信息,包括进程、内存、交换区、io、cpu信息

`vmstat 2 1`:每2秒采集一次数据，总共采集一次
`vmstat 2`:每2秒采集一次数据，一直采集直到手动结束

输出:
```shell
 procs -----------memory---------- ---swap-- -----io---- -system-- ------cpu-----
 r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st
 0  0 286208 793064 137032 4282296    2    6   150   139  143  370  7  8 85  0  0
 0  0 286208 794296 137032 4280740    0    0     0     0 1292 11140  3  2 95  0  0
 0  0 286208 799504 137032 4275416    0    0     0    78 1199 10728  3  3 94  0  0
 1  0 286208 799256 137040 4275360    0    0     0    26 1255 11088  3  3 94  0  0
 0  0 286208 806092 137040 4268304    0    0     0     0 1206 10779  2  3 95  0  0
 ```

* proc(进程)
  - r:运行队列中的进程数量
  - b:阻塞的进程数量

* memory(内存)
  - swpd:使用的虚拟内存大小
  - free:空闲物理内存
  - buff:用作缓冲的内存
  - cache:用作缓存的内存

* swap
  - si:每秒从磁盘读入虚拟内存的大小
  - so:每秒虚拟内存写入磁盘的大小

* io
  - bi:每秒读取的块数
  - bo:每秒写入的块数

* system
  - in:每秒CPU中断次数
  - cs:每秒上下文切换次数

* cpu
  - us:用户cpu时间所占百分比
  - sy:系统cpu时间所占百分比
  - id:空闲cpu时间所占百分比
  - wt:io等待的CPU时间所占百分比