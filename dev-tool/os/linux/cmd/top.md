## top

`top`命令用来查看进程的资源占用情况

输出结果:
```
top - 15:30:28 up  8:55,  1 user,  load average: 1.04, 1.34, 1.14
Tasks: 222 total,   1 running, 189 sleeping,   0 stopped,   0 zombie
%Cpu(s):  6.0 us,  3.2 sy,  0.0 ni, 90.5 id,  0.0 wa,  0.0 hi,  0.3 si,  0.0 st
KiB Mem :  8030888 total,  2904644 free,  2762808 used,  2363436 buff/cache
KiB Swap:  4194300 total,  3933436 free,   260864 used.  4324264 avail Mem 

  PID USER      PR  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND                                                                        
15624 senki     20   0 3353808 331924  66836 S  12.6  4.1  45:40.12 WeChat.exe                                                                 
 2335 root      20   0  572536 129964  97656 S   7.0  1.6  15:58.49 Xorg
 ```

 ## 第一行
 
 系统的运行时间和平均负载信息，同`uptime`的输出

 ## 第二行

 进程信息, 总共几个进程、几个正在运行、几个在睡眠、几个已经停止、几个僵死进程

 ## 第三行

 CPU状态信息

 * us:用户空间消耗CPU时间所占百分比
 * sy:内核空间消耗CPU时间所占百分比
 * ni:调整过优先级的进程的CPU时间的百分比
 * id:空闲CPU时间所占百分比
 * wa:io等待的CPU时间所占百分比
 * hi:硬件中断时间所占百分比
 * si:软件中断所占百分比
 * st:当Linux系统是在虚拟机中运行时，等待CPU资源的时间占比

 ## 第四行

 内存状态

 总共多少kb、空闲多少kb、使用多少kb、buff和cache占用多少kb

 ## 第五行

 swap交换区信息

 总共多少kb、空闲多少kb、使用多少kb、剩余多少kb

## 第六行

空行

## 第七行

各进程的状态

* PID:进程ID
* USER:进程所有者
* PR:进程优先级，越小优先级越高
* NI：进程nice值，负数表示高优先级，正值表示低优先级(PR(new)=PR(old)+nice)
* VIRT:进程使用的虚拟内存大小，单位kb。VIRT=SWAP+RES
* RES:进程使用的且未被换出的物理内存大小，单位kb
* SHR:共享内存大小，单位kb
* S:进程状态。D=不可中断的睡眠状态 R=运行 S=睡眠 T=跟踪/停止 Z=僵尸进程
* %CPU:CPU时间占用百分比
* %MEM:进程使用的物理内存百分比
* TIM+:进程使用的CPU时间总计
* COMMAND:进程的名称