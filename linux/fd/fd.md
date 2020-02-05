    所有进程打开的文件描述符数不能超过/proc/sys/fs/ file-max
    单个进程打开的文件描述符数不能超过user limit中  nofile的soft limit
    nofile的soft limit不能超过其hard limit
    nofile的hard limit不能超过/proc/sys/fs/nr_open
   
    最大文件描述符
    系统最大打开文件描述符数：/proc/sys/fs/file-max

    永久设置：修改/etc/sysctl.conf文件，增加fs.file-max = 1000000

    进程最大打开文件描述符数

    使用ulimit -n查看当前设置。使用ulimit -n 1000000进行临时性设置。
    要想永久生效，你可以修改/etc/security/limits.conf文件 
    root      hard    nofile      1000000
    root      soft    nofile      1000000
    hard limit不能大于/proc/sys/fs/nr_open

    当前系统使用的打开文件描述符数
    cat /proc/sys/fs/file-nr
    其中第一个数表示当前系统已分配使用的打开文件描述符数，第二个数为分配后已释放的（目前已不再使用），第三个数等于file-max