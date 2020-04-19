下面的环境变量控制着Go运行时的行为，它们的用法和含义在以后的版本可能发生变化。

### GOGC

设置初始的垃圾收集百分比。当新分配的数据与上一次收集后剩余的活动数据的比率达到此百分比时，将触发垃圾收集。默认GOGC=100。设置GOGC=off可以禁用垃圾收集。使用`runtime/debug`包中的`SetGCPercent`可以在运行时修改这个配置

### GODEBUG

该变量用于控制运行时的调试变量。可以设置为逗号分割的多个`name=val`对。可用配置有:

* allocfreetrace

    设置`allocfreetrace=1`会导致对每个对象的分配和释放进行概要分析和栈跟踪。

* clobberfree

    设置`clobberfree=1`垃圾回收器在释放对象时用错误内容破坏对象的内存内容。

* cgocheck

* efence

* gccheckmark

* gcpacertrace

* gcshrinkstackoff

* gcstoptheworld

* gctrace

    设置gctrace=1后在每次gc后会打印日志  
    日志格式: gc # @#s #%: #+#+# ms clock, #+#/#/#+# ms cpu, #->#-># MB, # MB goal, # P  
    gc # ：第几次gc  
    @#s  ：程序运行的时间 单位秒   
    #%   ：gc花费的时间占运行时间(自启动以来)的百分比  
    #+...+# ms clock ： gc各阶段的时间
    #+...+# ms cpu ：  gc各阶段占用cpu时间
    #->#-># MB ：gc前堆的大小，gc后堆的大小，存活堆的大小 
    #MB goal ：下次触发gc的堆大小阈值
    \# P  ：使用的处理器数量

`gc 281 @4.558s 4%: 0+1.0+0 ms clock, 0+1.0/0/0+0 ms cpu, 4->4->0 MB, 5 MB goal, 8 P`

281：第281次执行  
@4.558s：程序运行了4.558秒  
4%：gc花费的时间占运行时间的4%  
0+1.0+0 ms clock：  
0+1.0/0/0+0 ms cpu：  
4->4->0 MB：gc前堆4MB、gc后4MB、存活  
5 MB goal：下次回收的阈值是5MB  
8 P：使用了8个P

