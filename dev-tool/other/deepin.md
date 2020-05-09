
deepin启动报错[Failed to start Light Display Manager](http://www.linuxboy.net/deepinjc/139490.html)

1. 通过启动盘进入终端
2. sudo mount  /dev/sda3  /mnt  (挂在系统的根分区) 
3. sudo chroot /mnt    (进入系统根分区)
4. sudo apt install dde  （重新按照dde,所有要有网）
5. 重启
