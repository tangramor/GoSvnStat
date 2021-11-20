# GoSvnStat

A svn stat tool written by Go.  
用GO写的用来统计每个人的代码提交数的工具。


## Build 编译

配置好 golang 开发环境，执行 `go build`。下面是在 MacOS 上的 oh-my-zsh 下的操作过程。

```
➜  GoSvnStat git:(master) go build
go: cannot find main module, but found .git/config in /Users/wangjh/workspace/Goland/GoSvnStat
	to create a module there, run:
	go mod init

➜  GoSvnStat git:(master) go mod init GoSvnStat
go: creating new go.mod: module GoSvnStat
go: to add module requirements and sums:
	go mod tidy

➜  GoSvnStat git:(master) ✗ go env -w GO111MODULE="on"

➜  GoSvnStat git:(master) ✗ go mod tidy

➜  GoSvnStat git:(master) ✗ go build
```


## Usage 用法：

配置运行环境，把客户端证书放到合适的目录，然后修改 `~/.subversion/servers`

```
[groups]
dev = svn.dev.50dg.com

[dev]
ssl-client-cert-file = <这里写.p12证书文件的全路径>
store-plaintext-passwords = no

```

运行编译好的 GoSvnStat，以下是可使用的参数。

* -url 参数指定 svn 仓库 URL，**必需参数**
* -d 参数指定 svn 的开发路径
* -t 参数指定画图的模版文件路径，模版文件是项目根目录下的 gostatsvn.html 文件
* -all 参数统计到当前的全部日志信息，，该参数与其它时间、版本参数互斥。
* -y 参数指定要统计的年份，该参数与其它时间、版本参数互斥。格式为 `2017`。如果指定了此参数，将一并生成**当年**按季度、按月份和按星期的统计数据
* -q 参数指定要统计的季度，该参数与其它时间、版本参数互斥。格式为 `2017Q2`，其中 Q 后跟第几季度数字。
* -m 参数指定要统计的月份，该参数与其它时间、版本参数互斥。格式为 `2017-02`。
* -w 参数指定要统计的星期，该参数与其它时间、版本参数互斥。格式为 `2017W5`，其中 W 后跟第几周数字。
* -s 参数限定版本开始日期，未指定的话就是一天前，格式为 `2017-03-06`，也可以直接使用版本数字
* -e 参数限定版本结束日期，未指定的话就是今天，格式为 `2017-03-06`，也可以直接使用版本数字或 `HEAD`
* -n 参数指定 svn 的 xml 格式日志和统计数据文件名称前缀，未指定的话就是 Temp
* -reg 参数确定是否要强制重新生成日志文件和统计文件，`y` 或 `n`，缺省为 `n`
* -csvlog 参数确定是否要导出生成 csv 格式日志文件，`y` 或 `n`，缺省为 `n`。但选择 -y 参数时会缺省自动生成指定年度的 csv 日志文件

生成的 svn 日志会放置在当前目录下的 `svn_logs` 子目录；

如果有 csv 导出，则会放置到当前目录下的 `svn_csv_logs` 子目录，包含 `_commits.csv` 和 `_paths.csv` 两个文件，以 `revision` 关联；

统计数据会放置在当前目录下的 `svn_stats` 子目录。

```
# svn 命令行日志生成方法： svn log -r {2021-11-01}:{2021-11-16} -v --xml https://svn.dev.50dg.com/icesvn/ice_server > ice_server_svnlog_202111.xml

./GoSvnStat -all y -csvlog y -n ice_server -url https://svn.dev.50dg.com/icesvn/ice_server

./GoSvnStat -s 2021-01-01 -e 2021-11-16 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svn.dev.50dg.com/icesvn/ice_server

./GoSvnStat -y 2017 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svn.dev.50dg.com/icesvn/ice_server

./GoSvnStat -q 2017Q3 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svn.dev.50dg.com/icesvn/ice_server

./GoSvnStat -m 2018-02 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svn.dev.50dg.com/icesvn/ice_server

./GoSvnStat -w 2018W11 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svn.dev.50dg.com/icesvn/ice_server
```

