# GoSvnStat

A svn stat tool written by Go.  
用GO写的用来统计每个人的代码提交数的工具。


## Build 编译

配置好 golang 开发环境，执行 `go build`


## Usage 用法：

配置运行环境，把客户端证书放到合适的目录，然后修改 `~/.subversion/servers`

```
[groups]
dev = svn.dev.50dg.com

[dev]
ssl-client-cert-file = <这里写.p12证书文件的全路径>
store-plaintext-passwords = no

```

运行编译好的 GoSvnStat

* -url 参数指定 svn 仓库 URL，必需参数
* -d 参数指定 svn 的开发路径
* -t 参数指定画图的模版文件路径，模版文件是项目根目录下的 gostatsvn.html 文件
* -s 参数限定版本开始日期，未指定的话就是一天前
* -e 参数限定版本结束日期，未指定的话就是今天
* -n 参数指定 svn 的 xml 格式日志名称前缀，未指定的话就是 Temp
* -reg 参数确定是否要强制重新生成日志文件，y 或 n

```
# svn log -r {2021-11-01}:{2021-11-16} -v --xml https://svn.dev.50dg.com/icesvn/ice_server > ice_server_svnlog_202111.xml

./GoSvnStat -s 2021-01-01 -e 2021-11-16 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svn.dev.50dg.com/icesvn/ice_server
```

