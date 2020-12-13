# DNS

A(Address)记录: 用来指定主机名（或域名）对应的IP地址记录

AAAA记录：主机名（或域名）指向一个IPv6地址

CNAME记录：将域名指向一个域名，实现与被指向域名相同的访问效果

NS记录： 域名解析服务器记录，如果要将子域名指定某个域名服务器来解析，需要设置NS记录

SRV记录：

MX记录：

显性URL转发记录： 将域名指向一个http(s)协议地址，访问域名时，自动跳转至目标地址。例如：将www.example.cn显性转发到www.itbilu.com后，访问www.example.cn时，地址栏显示的地址为：www.itbilu.com。

隐性UR转发记录： 将域名指向一个http(s)协议地址，访问域名时，自动跳转至目标地址，隐性转发会隐藏真实的目标地址。例如：将www.example.cn显性转发到www.example.com后，访问www.example.cn时，地址栏显示的地址仍然是：www.example.cn。
