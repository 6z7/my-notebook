# 查看某个内核参数
sysctl fs.file-max

# 临时修改某个内核参数
sysctl -w fs.file-max=100000  # 设置文件打开数
sysctl -w net.ipv4.ip_forward=1  # 开启IP转发

修改sysctl.conf文件使其生效  sysctl -p

# 修改端口范围
net.ipv4.ip_local_port_range = 1024 65535

# 1024以下端口都是超级管理员用户（如root）才可以使用，普通用户只能使用大于1024的端口值


vm.overcommit_memory = 1