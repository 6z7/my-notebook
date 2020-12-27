`/proc/sys/fs/file-max`

file-max：所有进程允许打开的最大fd数量

file-max一般为内存大小（KB）的10%
grep -r MemTotal /proc/meminfo | awk '{printf("%d",$2/10)}'

临时修改  echo 1000000 > /proc/sys/fs/file-max
永久修改  /etc/sysctl.conf中设置 fs.file-max = 1000000