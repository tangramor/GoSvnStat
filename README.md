# GoSvnStat

A svn stat tool written by Go. Based on https://github.com/DigDeeply/GoStatsvn and added more features.

用GO写的用来统计每个人的代码提交数的工具。基于 https://github.com/DigDeeply/GoStatsvn 修改实现。



## What it can do 这个工具能做什么

This tool can generate svn stats based on svn logs, calculate how many commits were made by a developer and average commit times per day, and how many files were added/modified/deleted during a period. If the tool was executed in the source code folder, it can also count the lines made by the developer.

此工具可以基于 svn log，来统计一段时间内开发人员进行的提交次数、平均每日提交次数、提交的文件增/删/改数量。如果是在开发源码目录下执行此工具，还可以统计开发者的提交行数。

Stats Example output / 统计结果输出样例 （CSV）

```csv
Author,Commits,AverageCommitsPerDay,Lines,FilesAdded,FilesModified,FilesDeleted,StartDate,EndDate,ProjectId,SvnUrlId
developer1,1,0.0110,0,339,0,0,2017-04-01 00:00:00,2017-06-30 23:59:59,3,6
developer1,51,0.5543,0,3212,444,27,2017-07-01 00:00:00,2017-09-30 23:59:59,3,6
developer23,46,0.5000,0,56315,223,28,2017-07-01 00:00:00,2017-09-30 23:59:59,3,6
developer34,1,0.0109,0,1,1,0,2017-10-01 00:00:00,2017-12-31 23:59:59,3,6
```

The above example was generated with `-csvextf="ProjectId,SvnUrlId" -csvextv="3,6"` options, because it will be imported into database with association of `Project` and `SvnUrl` tables.

上面的例子是此工具在执行时添加了 `-csvextf="ProjectId,SvnUrlId" -csvextv="3,6"` 参数生成的，目的是为了方便导入数据库时，这两个参数用于和 `Project` 、 `SvnUrl` 两个表通过 ID 做关联。

It can also generate JSON formate stats, which can be used for web frontend chart.

此工具也可以生成 JSON 格式的统计结果，可以直接用作 Web 图表展示的数据源。




## Build 编译

Configure golang develop environment and execute `go build`.

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

➜  GoSvnStat git:(master) ✗ go env -w GOPROXY=https://goproxy.cn,direct

➜  GoSvnStat git:(master) ✗ go env -w GO111MODULE="on"

➜  GoSvnStat git:(master) ✗ go mod tidy

➜  GoSvnStat git:(master) ✗ CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' .
```



## Usage 用法：

Configure execution evironment, setup svn client tool and config svn server/user. For example, if the svn server needs cert file, you need put the file to a path and config  `~/.subversion/servers` file with corresponding path.

配置运行环境，如果 svn 服务器是使用证书登录的，把客户端证书放到合适的目录，然后修改 `~/.subversion/servers`

```
[groups]
dev = svnserver.com

[dev]
ssl-client-cert-file = <这里写.p12证书文件的全路径>
store-plaintext-passwords = no

```

运行编译好的 GoSvnStat，以下是可使用的参数。

* -h 参数打印帮助信息
* -url 参数指定 svn 仓库 URL，**必需参数**
* -u 参数指定 svn 用户名，只输出此用户的日志和统计信息
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
* -logextf 参数确定是否要在导出的 csv 格式日志文件里附加字段，例如 projectid，需要 -csvlog=y，为空值即不添加
* -logextv 参数为要在导出的 csv 格式日志文件里附加字段的值，需要 -csvlog=y，若 -logextf 为空值此参数无效果
* -csv 参数确定是否要生成 csv 格式统计文件，`y` 或 `n`，缺省为 `y`
* -csvextf 参数确定是否要在导出的 csv 格式统计文件里附加字段，例如 "projectid,svnurlid"，需要 -csv=y，为空值即不添加
* -csvextv 参数为要在导出的 csv 格式统计文件里附加字段的值，例如 "21,3"，需要 -csv=y，若 -csvextf 为空值此参数无效果
* -json 参数确定是否要生成 json 格式统计文件，`y` 或 `n`，缺省为 `n`

```
Usage: GoSvnStat [-htyqmwsedn] [-all] [-url=svn_repo_url] [-reg=y] [-csvlog=y] [-logextf=projectid] [-logextv=1] [-json=y] [-csv=y] [-csvextf=projectid] [-csvextv=1] 

Options:
  -all string
        svn log for all; priority 6 (default "n")
  -csv string
        generate csv stats files, y or n (default "y")
  -csvextf string
        append extra field to csv stat files, need -csv=y
  -csvextv string
        append extra field value to csv stat files, need -csv=y
  -csvlog string
        generate csv log files, y or n (default "n")
  -d string
        code working directory
  -e string
        svn log end date, like 2006-01-03, or reversion number, or HEAD; priority 1
  -h    print help information
  -json string
        generate json stats files, y or n (default "n")
  -logextf string
        append extra field to csv log files, need -csvlog=y
  -logextv string
        append extra field value to csv log files, need -csvlog=y
  -m string
        svn log for a month, like 2006-01; priority 3
  -n string
        svn log file name prefix
  -q string
        svn log for a quarter, like 2006Q1; priority 4
  -reg string
        force to regenerate log file, y or n (default "n")
  -s string
        svn log start date, like 2006-01-02, or reversion number; priority 1
  -t string
        hightcharts Template file
  -u string
        author name, svn logs/stats for this only author
  -url string
        svn repository URL
  -w string
        svn log for a week like 2006W20; priority 2
  -y string
        svn log for a year, like 2006; priority 5
```

The generated svn log files will be in `svn_logs` folder.

生成的 svn 日志会放置在当前目录下的 `svn_logs` 子目录；

If you choose to generate cdv format logs, they will be put under `svn_csv_logs` folder. There will be `_commits.csv` and `_paths.csv` 2 files associated by revision.

如果有 csv 导出，则会放置到当前目录下的 `svn_csv_logs` 子目录，包含 `_commits.csv` 和 `_paths.csv` 两个文件，以 `revision` 关联；

Stat output files will be put in `svn_stats`.

统计数据会放置在当前目录下的 `svn_stats` 子目录。

Some Usage Examples / 几个使用实例：

```
# svn 命令行日志生成方法： svn log -r {2021-11-01T00:00:00Z}:{2021-11-16T23:59:59Z} -v --xml https://svnserver.com/icesvn/ice_server > ice_server_svnlog_202111.xml

./GoSvnStat -y 2017 -n ice_server -csvlog=y -logextf=projectid -logextv=1 -csvextf="ProjectId,SvnUrlId" -csvextv="1,1" -url https://svnserver.com/icesvn/ice_server

./GoSvnStat -url https://svnserver.com/icesvn/ice_server -y 2017 -n ice_server -csvlog=y -logextf=projectid -logextv=1 -csv=n -json=y

./GoSvnStat -url https://svnserver.com/icesvn/ice_server -m 2017-12 -n ice_server -csvlog=y -logextf=projectid -logextv=1 -csv=n -json=y

./GoSvnStat -all y -csvlog y -n ice_server -url https://svnserver.com/icesvn/ice_server

./GoSvnStat -s 2021-01-01 -e 2021-11-16 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svnserver.com/icesvn/ice_server

./GoSvnStat -y 2017 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svnserver.com/icesvn/ice_server

./GoSvnStat -q 2017Q3 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svnserver.com/icesvn/ice_server

./GoSvnStat -m 2018-02 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svnserver.com/icesvn/ice_server

./GoSvnStat -w 2018W11 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svnserver.com/icesvn/ice_server

./GoSvnStat -u zhangsan -y 2017 -n ice_server -reg y -d /root/test -t /root/test/gostatsvn.html -url https://svnserver.com/icesvn/ice_server
```

