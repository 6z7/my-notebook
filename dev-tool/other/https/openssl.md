# openssl

```
// 生成12位的随机数 采用base64编码
openssl rand -base64 12

// 支持的密码算法
openssl list --cipher-algorithms


// 生成RSA密钥对
openssl genrsa -out mykey.pem 2048

// 校验密码对文件是否正确
openssl rsa -in mykey.pem -check -noout

// 从秘钥对中分离出公钥
openssl rsa  -in aa.pem -pubout -out mypubkey.pem

// 显示秘钥文件中的公钥信息
openssl rsa -in  mypubkey.pem -pubin

// 通过秘钥文件加密plain.txt文件
openssl rsautl -encrypt -inkey mykey.pem -in plain.txt -out cipher.txt

// 通过公钥加密plain.txt文件
openssl rsautl -encrypt -pubin  -inkey mypubkey.pem -in plain.txt -out cipher.txt

// 解密
openssl rsautl -decrypt -inkey mykey.pem -in cipher.txt -out cipher2.txt


// 生成DH密钥协商算法所需的参数
openssl dhparam -out dhparam.pem -2 1024
// 查看生成的参数
openssl dhparam -in dhparam.pem -noout -C
// 基于参数文件生成秘钥对文件
openssl genpkey -paramfile dhparam.pem -out dhkey.pem
// 查看生成的秘钥对文件
openssl pkey -in dhkey.pem -text -noout



// 查看系统有多少椭圆曲线，通信双方需要选择一条都支持的命名曲线
openssl ecparam -list_curves

// 生成一个参数文件，通过 -name指定命名曲线
openssl ecparam -name secp256k1-out secp256k1.pem

// 默认的情况下，查看参数文件只会显示曲线的名称
openssl ecparam -in secp256k1.pem -text -noout

// 显示参数文件里的具体参数
openssl ecparam -in secp256k1.pem -text -param_enc explicit -noout
```
