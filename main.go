package main

import (
	"GoSvnStat/statStruct"
	"GoSvnStat/util"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DEFAULT_SMALLEST_TIME_STRING = "1000-03-20T08:38:17.428370Z"
	DATE_DAY                     = "2006-01-02"
	DATE_HOUR                    = "2006-01-02 15"
	DATE_SECOND                  = "2006-01-02T15:04:05Z"
)

var year *string = flag.String("y", "", "svn log for a year, like 2006; priority 5")
var quarter *string = flag.String("q", "", "svn log for a quarter, like 2006Q1; priority 4")
var month *string = flag.String("m", "", "svn log for a month, like 2006-01; priority 3")
var week *string = flag.String("w", "", "svn log for a week like 2006W20; priority 2")
var startDate *string = flag.String("s", "", "svn log start date, like 2006-01-02, or reversion number; priority 1")
var endDate *string = flag.String("e", "", "svn log end date, like 2006-01-03, or reversion number, or HEAD; priority 1")

var svnDir *string = flag.String("d", "", "code working directory")
var svnUrl *string = flag.String("url", "", "svn repository URL")
var logNamePrefix *string = flag.String("n", "", "svn log file name prefix")
var reGenerate *string = flag.String("reg", "n", "force to regenerate log file, y or n")
var chartTemplate *string = flag.String("t", "", "hightcharts Template file")
var chartData statStruct.ChartData

func main() {
	flag.Parse()

	//判断 svn URL 是否指定
	if *svnUrl == "" {
		log.Fatal("-url cannot be empty, -url svn repository URL")
		return
	}

	pwd, _ := os.Getwd()

	//判断有没有指定画图的模版文件
	if *chartTemplate == "" {
		log.Println("-t is empty, -t hightchartsTemplate file path")
		*chartTemplate = pwd + "/gostatsvn.html"
	}

	//判断有没有指定 svnWorkDir
	if *svnDir == "" {
		log.Println("-d is empty, -d svnWorkDir")
		*svnDir = pwd
	}

	//判断有没有指定重新生成日志文件
	if *reGenerate == "y" {
		log.Println("-reg is y, will re-generate the log file")
	}

	var dontignore = true //按优先级忽略

	//按开始、结束日期统计，第一优先，即有这个参数就忽略其它时间参数
	if *startDate != "" && *endDate != "" {
		dontignore = false
	}

	//按某年星期统计，第二优先
	//星期以周一为起始日
	if *week != "" && dontignore {
		dontignore = false

		yw := strings.Split(*week, "W")
		if yw[1] == "" {
			log.Fatal("Week format wrong, it should be like 2006W02, exit.")
			return
		}
		y, err_y := strconv.Atoi(yw[0]) //转换年份为整数值
		w, err_w := strconv.Atoi(yw[1]) //转换周数为整数值
		if err_w == nil && err_y == nil {
			s, e, err := util.GetWeekStartEnd(y, w)
			if err == nil {
				*startDate = s
				*endDate = e
			}
		}
	}

	//按某年月度统计，第三优先
	if *month != "" && dontignore {
		dontignore = false

		ym := strings.Split(*month, "-")
		if ym[1] == "" {
			log.Fatal("Month format wrong, it should be like 2006-02, exit.")
			return
		}
		y, err_y := strconv.Atoi(ym[0]) //转换年份为整数值
		m, err_m := strconv.Atoi(ym[1]) //转换月份为整数值
		if err_m == nil && err_y == nil {
			s, e, err := util.GetMonthStartEnd(y, m)
			if err == nil {
				*startDate = s
				*endDate = e
			}
		}
	}

	//按某年季度统计，第四优先
	if *quarter != "" {
		dontignore = false

		yq := strings.Split(*quarter, "Q")
		if yq[1] == "" {
			log.Fatal("Quarter format wrong, it should be like 2006Q2, exit.")
			return
		}
		y, err_y := strconv.Atoi(yq[0]) //转换年份为整数值
		q, err_q := strconv.Atoi(yq[1]) //转换季度为整数值

		if err_q == nil && err_y == nil {
			s, e, err := util.GetQuarterStartEnd(y, q)
			if err == nil {
				*startDate = s
				*endDate = e
			}
		}
	}

	//按年统计
	if *year != "" && dontignore {
		*startDate = *year + "-01-01"
		*endDate = *year + "-12-31"
	}

	//获取天数
	days := util.GetDurationDays(*startDate, *endDate)
	log.Printf("Total %d days during the stats", days)

	//生成 svn 日志文件
	svnXmlFile, err := util.GetSvnLogFile(*startDate, *endDate, *svnUrl, *logNamePrefix, *reGenerate)

	if err != nil {
		log.Println(err)
		log.Fatal("Failed to generate svn xml log file, exit.")
		return
	}

	//判断文件是否存在
	if _, err := os.Stat(svnXmlFile); os.IsNotExist(err) {
		log.Fatalf("svn log file '%s' not exists.", svnXmlFile)
		return
	}

	log.Printf("svn log file is %s \n", svnXmlFile)

	authorTimeStats, AuthorStats := util.GenerateStat(*svnDir, svnXmlFile, days)

	//输出结果
	ConsoleOutPutTable(AuthorStats)
	//fmt.Printf("%v\n", authorTimeStats)
	minTimestamp, maxTimestamp := getMinMaxTimestamp(authorTimeStats)
	fmt.Printf("%d\t%d\n", minTimestamp, maxTimestamp)
	dayAuthorStats := StatLogByDay(authorTimeStats)
	fmt.Printf("%v\n", dayAuthorStats)
	dayAuthorStatsOutput := StatLogByFullDay(dayAuthorStats, minTimestamp, maxTimestamp)
	xaxis := util.GetXAxis(minTimestamp, maxTimestamp)
	series := util.GetSeries(dayAuthorStatsOutput)
	chartData.XAxis = xaxis
	chartData.Series = series
	fmt.Printf("%s\n%s\n", xaxis, series)
	DrawCharts()
	//输出按小时统计结果
	//ConsoleOutPutHourTable(authorTimeStats)
	//输出按周统计结果
	//ConsoleOutPutWeekTable(authorTimeStats)

}

//console输出结果
func ConsoleOutPutTable(AuthorStats map[string]statStruct.AuthorStat) { /*{{{*/
	fmt.Printf(" ==User== \t==Commits== ==AveragePerDay== ==Lines== ==Added== ==Modified== ==Deleted==\n")
	for author, val := range AuthorStats {
		fmt.Printf("%10s\t%5d\t%10d\t%10d\t%7d\t%10d\t%7d\n", author, val.CommitCount, val.AverageCommitsPerDay, val.AppendLines+val.RemoveLines, val.AddedFiles, val.ModifiedFiles, val.DeletedFiles)
	}
} /*}}}*/

//返回按天格式化好的数据
func StatLogByDay(authorTimeStats statStruct.AuthorTimeStats) (dayAuthorStats statStruct.AuthorTimeStats) { /*{{{*/
	dayAuthorStats = make(map[string]statStruct.AuthorTimeStat)
	for author, detail := range authorTimeStats {
		dayAuthorStat := make(map[string]statStruct.AuthorStat)
		_, ok := dayAuthorStats[author]
		if !ok { //初始化
			dayAuthorStats[author] = dayAuthorStat
		}
		for timeString, stats := range detail {
			//todo 找到正常转化时间的方法
			timeTime, err := time.Parse(time.RFC3339, timeString)
			util.CheckErr(err)
			timeFormat := timeTime.Format(DATE_DAY)
			//fmt.Printf("%v\t%v\n", timeString, timeTime)
			if err == nil {
				oldDayAuthorStat, ok := dayAuthorStat[timeFormat]
				var authorStat statStruct.AuthorStat
				if ok {
					authorStat.CommitCount = oldDayAuthorStat.CommitCount + stats.CommitCount
					authorStat.AppendLines = oldDayAuthorStat.AppendLines + stats.AppendLines
					authorStat.RemoveLines = oldDayAuthorStat.RemoveLines + stats.RemoveLines
					authorStat.AddedFiles = oldDayAuthorStat.AddedFiles + stats.AddedFiles
					authorStat.ModifiedFiles = oldDayAuthorStat.ModifiedFiles + stats.ModifiedFiles
					authorStat.DeletedFiles = oldDayAuthorStat.DeletedFiles + stats.DeletedFiles
				} else {
					authorStat.CommitCount = stats.CommitCount
					authorStat.AppendLines = stats.AppendLines
					authorStat.RemoveLines = stats.RemoveLines
					authorStat.AddedFiles = stats.AddedFiles
					authorStat.ModifiedFiles = stats.ModifiedFiles
					authorStat.DeletedFiles = stats.DeletedFiles
				}
				dayAuthorStat[timeFormat] = authorStat
			}
		}
		dayAuthorStats[author] = dayAuthorStat
	}
	return
} /*}}}*/

func StatLogByFullDay(dayAuthorStats statStruct.AuthorTimeStats, minTimestamp int64, maxTimestamp int64) (dayAuthorStatsOutput statStruct.AuthorTimeStats) { /*{{{*/
	//得到时间的开始和结束日期
	minTime := time.Unix(minTimestamp, 0)
	minDay := minTime.Format(DATE_DAY)
	minTime, _ = time.Parse(DATE_DAY, minDay)
	minDayTimestamp := minTime.Unix()
	maxTime := time.Unix(maxTimestamp, 0)
	maxDay := maxTime.Format(DATE_DAY)
	maxTime, _ = time.Parse(DATE_DAY, maxDay)
	maxDayTimestamp := maxTime.Unix()
	dayAuthorStatsOutput = make(statStruct.AuthorTimeStats)
	//遍历所有author
	for author, dayAuthorStat := range dayAuthorStats {
		fmt.Printf("====user: %s=====\n", author)
		minDayAuthor := minDay
		minTimeAuthor := minTime
		minDayTimestampAuthor := minDayTimestamp
		dayAuthorStatOutput := make(statStruct.AuthorTimeStat)
		//输出每个用户每天的信息
		for {
			authorStat, ok := dayAuthorStat[minDayAuthor]
			if ok {
				fmt.Printf("%s\t%d\t%d\t%d\t%d\t%d\n", minDayAuthor, authorStat.CommitCount, authorStat.AppendLines+authorStat.RemoveLines, authorStat.AddedFiles, authorStat.ModifiedFiles, authorStat.DeletedFiles)
				dayAuthorStatOutput[minDayAuthor] = authorStat
			} else {
				fmt.Printf("%s\t%d\t%d\t%d\t%d\t%d\n", minDayAuthor, 0, 0, 0, 0, 0)
				authorStat.CommitCount = 0
				authorStat.AppendLines = 0
				authorStat.RemoveLines = 0
				authorStat.AddedFiles = 0
				authorStat.ModifiedFiles = 0
				authorStat.DeletedFiles = 0
				dayAuthorStatOutput[minDayAuthor] = authorStat
			}
			minDayTimestampAuthor += 86400
			minTimeAuthor = time.Unix(minDayTimestampAuthor, 0)
			minDayAuthor = minTimeAuthor.Format(DATE_DAY)
			if minDayTimestampAuthor > maxDayTimestamp {
				break
			}
		}
		dayAuthorStatsOutput[author] = dayAuthorStatOutput
	}
	fmt.Printf("%v\n", dayAuthorStatsOutput)
	return
} /*}}}*/

//console 按天输出结果，空余的天按0补齐
//获取时间的最大值和最小值
func getMinMaxTimestamp(authorTimeStats statStruct.AuthorTimeStats) (minTimestamp int64, maxTimestamp int64) { /*{{{*/
	minTimestamp = 0
	maxTimestamp = 0
	//先取得时间的最大值和最小值
	for _, detail := range authorTimeStats {
		//fmt.Printf("%s\t%v\n", author, detail)
		for timeString, _ := range detail {
			timeTime, err := time.Parse(DATE_SECOND, timeString)
			if err == nil {
				if minTimestamp == 0 || minTimestamp > timeTime.Unix() {
					minTimestamp = timeTime.Unix()
				}
				if maxTimestamp < timeTime.Unix() {
					maxTimestamp = timeTime.Unix()
				}
			}
		}
		//fmt.Printf("%d\t%d\n", minTimestamp, maxTimestamp)
	}
	return
} /*}}}*/

//console按小时输出结果
//todo 此处有bug,1.没有全部按小时归并，还是按每天每小时归并的。2.显示的小时不是按24小时制
func ConsoleOutPutHourTable(authorTimeStats statStruct.AuthorTimeStats) { /*{{{*/
	defaultSmallestTime, _ := time.Parse("2006-01-02T15:04:05Z", DEFAULT_SMALLEST_TIME_STRING)
	fmt.Printf(" ==user== \t==hour==\t==commits== ==lines== ==Added== ==Modified== ==Deleted==\n")
	//先取到时间的区间值
	for authorName, Author := range authorTimeStats {
		var minTime time.Time
		var maxTime time.Time
		for sTime, _ := range Author {
			fmtTime, err := time.Parse(DATE_HOUR, sTime)
			util.CheckErr(err)
			if minTime.Before(defaultSmallestTime) || minTime.After(fmtTime) {
				minTime = fmtTime
			}
			if maxTime.Before(defaultSmallestTime) || maxTime.Before(fmtTime) {
				maxTime = fmtTime
			}
		}
		//Todo 用户按时合并,去重
		//输出单个用户的数据
		for sTime, Sval := range Author {
			fmtTime, err := time.Parse(DATE_HOUR, sTime)
			util.CheckErr(err)
			fmt.Printf("%10s\t%5d\t%5d\t%12d\t%10d\t%10d\t%10d\n", authorName, fmtTime.Hour(), Sval.CommitCount, Sval.AppendLines+Sval.RemoveLines, Sval.AddedFiles, Sval.ModifiedFiles, Sval.DeletedFiles)
		}
	}
} /*}}}*/

//console按周输出结果
func ConsoleOutPutWeekTable(authorTimeStats statStruct.AuthorTimeStats) { /*{{{*/
	weekAuthorStats := make(map[string]map[string]statStruct.AuthorStat)
	for authorName, Author := range authorTimeStats {
		weekAuthorStat := make(map[string]statStruct.AuthorStat)
		_, ok := weekAuthorStats[authorName]
		if ok {
		} else {
			weekAuthorStats[authorName] = weekAuthorStat
		}
		for sTime, sAuthor := range Author {
			fmtTime, err := time.Parse(DATE_HOUR, sTime)
			util.CheckErr(err)
			week := fmtTime.Weekday().String()
			oldAuthorStat, ok := weekAuthorStat[week]
			var authorStat statStruct.AuthorStat
			if ok {
				authorStat.CommitCount = oldAuthorStat.CommitCount + sAuthor.CommitCount
				authorStat.AppendLines = oldAuthorStat.AppendLines + sAuthor.AppendLines
				authorStat.RemoveLines = oldAuthorStat.RemoveLines + sAuthor.RemoveLines
				authorStat.AddedFiles = oldAuthorStat.AddedFiles + sAuthor.AddedFiles
				authorStat.ModifiedFiles = oldAuthorStat.ModifiedFiles + sAuthor.ModifiedFiles
				authorStat.DeletedFiles = oldAuthorStat.DeletedFiles + sAuthor.DeletedFiles
			} else {
				authorStat.CommitCount = sAuthor.CommitCount
				authorStat.AppendLines = sAuthor.AppendLines
				authorStat.RemoveLines = sAuthor.RemoveLines
				authorStat.AddedFiles = sAuthor.AddedFiles
				authorStat.ModifiedFiles = sAuthor.ModifiedFiles
				authorStat.DeletedFiles = sAuthor.DeletedFiles
			}
			weekAuthorStat[week] = authorStat
		}
		weekAuthorStats[authorName] = weekAuthorStat
	}
	fmt.Printf(" ==user== \t==week==\t==commits== ==lines== ==Added== ==Modified== ==Deleted==\n")
	allWeeks := []string{
		"Sunday ",
		"Monday",
		"Tuesday",
		"Wednesday",
		"Thursday",
		"Friday",
		"Saturday",
	}
	//输出
	for authorName, weekAuthorStat := range weekAuthorStats {
		for _, oneDay := range allWeeks {
			authorStat, ok := weekAuthorStat[oneDay]
			if ok {
				fmt.Printf("%10s\t%5s\t%10d\t%12d\t%10d\t%10d\t%10d\n", authorName, oneDay, authorStat.CommitCount, authorStat.AppendLines+authorStat.RemoveLines, authorStat.AddedFiles, authorStat.ModifiedFiles, authorStat.DeletedFiles)
			} else {
				fmt.Printf("%10s\t%5s\t%10d\t%12d\t%10d\t%10d\t%10d\n", authorName, oneDay, 0, 0, 0, 0, 0)
			}
		}
	}
} /*}}}*/

func showHandle(w http.ResponseWriter, r *http.Request) {
	//filename := r.FormValue("id")
	//imagePath := UPLOAD_DIR + "/" + filename

	//w.Header().Set("Content-Type", "text/html")
	//http.ServeFile(w, r, "src/gostatsvn.html")
	t, err := template.ParseFiles(*chartTemplate)
	if err != nil {
		log.Fatal("not find file: ", err.Error())
	} else {
		locals := make(map[string]interface{})
		xaxis := template.HTML(chartData.XAxis)
		series := template.HTML(chartData.Series)
		locals["xaxis"] = xaxis
		locals["series"] = series
		t.Execute(w, locals)
	}
}

func DrawCharts() {
	http.HandleFunc("/", showHandle)
	log.Println("listen on 8088")
	err := http.ListenAndServe(":8088", nil)
	if err != nil {
		log.Fatal("listen fatal: ", err.Error())
	}
}
