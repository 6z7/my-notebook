内存分配策略
0：表示内核将检查是否有足够的可用内存供应用进程使用；如果有足够的可用内存，内存申请允许；否则，内存申请失败，并把错误返回给应用进程。
1: 表示内核允许分配所有的物理内存，而不管当前的内存状态如何。
2: 不允许超过内存分配上限 (CommitLimit = 物理内存 * overcommit_ratio(默认50，即50%) + swap大小)