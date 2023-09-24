### 功能
- [x] redis 导入 和导出json  
- [x] redis集群   
- [x] 支持tls  
- [x] 支持定时任务调用
- [x] 支持数据类型
 - string list hash set zset bitmaps hyperloglogs  geospatial  bitfield stream
- [ ] 忽略有ttl的key
### 命令格式
```sh
./redis-back load/dump  data.json redis://用户名:密码@任意节点地址:端口
# 连接到tls节点
./redis-back load/dump  data.json rediss://用户名:密码@任意节点地址:端口
# 跳过证书验证
./redis-back load/dump data.json rediss://用户名:密码@任意节点地址:端口 skip
# 使用ca证书验证
./redis-back load/dump data.json rediss://用户名:密码@任意节点地址:端口 /path/ca.crt domian
./redis-back load/dump data.json rediss://用户名:密码@任意节点地址:端口 /path/client.crt#/path/client.key domian
```
解释
- load/dump 导入json 或者 导出json
- data.json 导入或者导出的json文件路径，
  -  在导出的时候路径格式可以使用 data{当前时间格式}.json  注意需要是golang默认支持的日期格式:例如` dump date-{2006-01-02-15-04-05}.json rediss://default:XXXX@10.1.1.20:6101 skip` 会写入到文件 `  date-2023-09-24-12-16-39.json` 不兼容 YYYY-MM-dd HH:mm:ss的写法哈
- redis://用户名:密码@任意节点地址:10601  
  - 默认用户名是 default ，如果无密码 密码留空即可
  - rediss 代表 tls
- skip 跳过tls证书验证
- /path/ca.crt  使用ca证书验证tls
- /path/client.crt#/path/client.key  使用证书文件验证tls  
- domian 指定证书内包含的域名信息 因为部分情况下连接地址和域名信息不一样所以这里强制使用这个参数
- 在使用tls协议的时候，默认认为节点之间也是tls通讯的。本工具会查询主节点的地址的和端口并使用tls去连接，如果节点之间是tcp明文通讯。请把 rediss://前缀  修改为  rediss-tcp://
### 常见问题
#### 导入的覆盖规则
同key名称 会用json内的信息替换掉。
#### tls
如果提示 tls: failed to verify certificate: x509: certificate is not valid for any names, but wanted to match XXX。 这是证书不包含你当前的域名。这是证书的问题，你可以用 跳过证书验证的方式来处理
#### stream
stream导出后会丢失客户订阅消息，请酌情使用。
### 集群导入导出原理
连接到单个节点上之后，会首先查询一次Cluster信息如果查询成功则按照Cluster集群处理。如果失败就按照单个节点处理。  
导出的时候会逐个轮询到Cluster的每一个主节点上查询keys 根据类型来处理。  导入的时候就直接导入了。
### 导出到另外一个redis服务器方法
```sh
 dataFile=$(./redis-back_exe dump  data-{2006-01-02-15-04-05}.json rediss://default:XXX@10.1.1.201:6101 skip | tail -n 1)
 echo "准备把 $dataFile 导入到下面服务器.."
./redis-back_exe load   $dataFile redis://default:XXXX@10.1.1.60:6000
```
### 编译

```
mkdir redis-back&& cd redis-back
wget https://github.com/joyanhui/redis-back/archive/refs/heads/main.zip 
unzip main.zip 
go get github.com/redis/go-redis/v9@v9.0.5
go build -ldflags '-s -w -linkmode "external" -extldflags "-static"'     -o "$(basename "$(pwd)")" *.go
```