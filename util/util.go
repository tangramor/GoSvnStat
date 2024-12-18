//工具类，封装若干工具方法
package util

import (
	"GoSvnStat/statStruct"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

//svn xml log path struct
type Path struct {
	Action   string `xml:"action,attr"`
	Kind     string `xml:"kind,attr"`
	Path     string `xml:",chardata"`
	PropMods string `xml:"prop-mods,attr"`
	TextMods string `xml:"text-mods,attr"`
}

//svn xml log logentry struct
type Logentry struct {
	Revision string `xml:"revision,attr"`
	Author   string `xml:"author"`
	Date     string `xml:"date"`
	Paths    []Path `xml:"paths>path"`
	Msg      string `xml:"msg"`
}

//svn xml log result struct
type SvnXmlLogs struct {
	Logentry []Logentry `xml:"logentry"`
}

//调用命令执行svn diff操作, 返回diff的结果
func CallSvnDiff(oldVer, newVer int, fileName string) (stdout string, err error) { /*{{{*/
	app := "svn"
	param1 := "diff"
	param2 := "--old"
	param3 := fileName + "@" + strconv.Itoa(oldVer)
	param4 := "--new"
	param5 := fileName + "@" + strconv.Itoa(newVer)

	cmd := exec.Command(app, param1, param2, param3, param4, param5)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", err
	} else {
		return out.String(), nil
	}
} /*}}}*/

//获取有diff的行数
func GetLineDiff(diffBuffer string) (appendLines, removeLines int, err error) { /*{{{*/
	//svndiff 结果头部有 --- +++ 标识,从-1开始计数跳过
	appendLines = -1
	removeLines = -1
	err = nil
	lines := strings.Split(diffBuffer, "\n")
	for _, line := range lines {
		if strings.Index(line, "+") == 0 {
			appendLines++
		}
		if strings.Index(line, "-") == 0 {
			removeLines++
		}
	}
	if appendLines == -1 || removeLines == -1 {
		appendLines = 0
		removeLines = 0
	}
	return
} /*}}}*/

//解析xml格式的svn log
func ParaseSvnXmlLog(svnXmlLogFile string) (svnXmlLogs SvnXmlLogs, err error) { /*{{{*/
	content, err := ioutil.ReadFile(svnXmlLogFile)
	if err != nil {
		log.Fatal(err)
	}
	err = xml.Unmarshal(content, &svnXmlLogs)
	if err != nil {
		log.Fatal(err)
	}
	return
} /*}}}*/

//获取svn根
func GetSvnRoot(workDir string) (svnRoot string, err error) { /*{{{*/
	pwd, _ := os.Getwd()
	if strings.HasPrefix(workDir, "/") {
		pwd = ""
	}
	app := "svn"
	param1 := "info"
	param2 := "--xml"
	param3 := pwd + "/" + workDir

	cmd := exec.Command(app, param1, param2, param3)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", err
	} else {
		re := regexp.MustCompile(`(?i)<root>(.*)</root>`)
		roots := re.FindStringSubmatch(out.String())
		if len(roots) > 1 {
			return roots[1], nil
		} else {
			log.Fatalf("cannot find the svn root by svn info")
			return "", nil
		}
	}
} /*}}}*/

//根据 svn revision 号获取提交日期
func GetSvnDateByRevision(revision string, svnUrl string) (date string, err error) { /*{{{*/
	app := "svn"
	param1 := "log"
	param2 := "--xml"
	param3 := "-r"

	cmd := exec.Command(app, param1, param2, param3, revision, svnUrl)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return "", err
	} else {
		re := regexp.MustCompile(`(?i)<date>(.*)</date>`)
		date := re.FindStringSubmatch(out.String())
		if len(date) > 1 {
			rdate, rerr := time.Parse(DATE_NANOSEC, date[1])
			CheckErr(rerr)
			return rdate.Format(DATE_DAY), nil
		} else {
			log.Fatalf("cannot find the svn date by svn revision")
			return "", nil
		}
	}
} /*}}}*/

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func createScriptFile(path string) {
	f, err := os.Create(path)
	if err != nil {
		return
	}

	_, err = fmt.Fprintf(f, `#!/bin/bash
startDate=$1
endDate=$2
svnUrl=$3
logName=$4
if [[ "$startDate" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]; then
    startDate="{"$startDate"}"
fi
if [[ "$endDate" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]; then
    endDate="{"$endDate"}"
fi
svn log -r $startDate":"$endDate --xml -v $svnUrl > $logName
`)

	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}

	err = f.Chmod(0755)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}

	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

//获取svn日志
func GetSvnLogFile(startDate string, endDate string, svnUrl string, namePrefix string, reGenerate bool) (svnLogFile string, err error) { /*{{{*/
	pwd, _ := os.Getwd()

	now := time.Now()

	if startDate == "" {
		startDate = now.AddDate(0, 0, -1).Format(DATE_DAY)
		log.Printf("Start Date is %s\n", startDate)
	}

	if endDate == "" {
		endDate = now.Format(DATE_DAY)
		log.Printf("End Date is %s\n", endDate)
	}

	log_folder := pwd + "/svn_logs/"
	if !fileExists(log_folder) {
		os.MkdirAll(log_folder, os.FileMode(0755))
	}

	log_name := namePrefix + "_svnlog_" + startDate + "_" + endDate + ".log"
	log_fullpath := log_folder + log_name

	log.Printf("Log filename is %s\n", log_name)

	//不强制重新生成日志，结束日期不是今天（因为今天可能还有新提交）且日志文件存在，则不重新生成
	if !reGenerate && endDate != now.Format(DATE_DAY) && fileExists(log_fullpath) {
		log.Printf("Log file %s already exists", log_fullpath)
		return log_fullpath, nil
	}

	if strings.Contains(startDate, "-") {
		startDate = startDate + DAY_START_SECOND
	}

	if strings.Contains(endDate, "-") {
		endDate = endDate + DAY_END_SECOND
	}

	app := "./GenerateSvnLog.sh"
	param1 := startDate
	param2 := endDate
	param3 := svnUrl
	param4 := log_fullpath

	//判断脚本文件是否存在
	if !fileExists(app) {
		log.Printf("script file '%s' not exists, will create it.", app)
		createScriptFile(app)
	}

	cmd := exec.Command(app, param1, param2, param3, param4)
	fmt.Printf("%v\n", cmd.Args)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return "", err
	} else {
		return param4, nil
	}
} /*}}}*/

//封装的checkErr
func CheckErr(err error) (err2 error) { /*{{{*/
	if err != nil {
		log.Panic(err)
		return err
	} else {
		return nil
	}
} /*}}}*/

//获取当年第几周的起始日期和结束日期
func GetWeekStartEnd(year int, week int) (startDate string, endDate string, err error) {
	if week < 53 && week > 0 {
		firstday, err := time.Parse(DATE_DAY, strconv.Itoa(year)+"-01-01")
		if err == nil {
			firstMon := firstday
			//获取当年第一天是周几
			fdy, w1 := firstday.ISOWeek()
			if fdy != year && w1 != 1 { //年份不相同，需要找到当年第一个周一
				offset := int(time.Sunday-firstday.Weekday()) + 1
				if offset < 0 {
					offset += 7
				}

				firstMon = time.Date(firstday.Year(), firstday.Month(), firstday.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
			}
			weekMonday := time.Date(firstMon.Year(), firstMon.Month(), firstMon.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, (week-1)*7)
			startDate := weekMonday.Format(DATE_DAY)
			weekSunday := weekMonday.AddDate(0, 0, 6)
			endDate := weekSunday.Format(DATE_DAY)

			return startDate, endDate, nil
		}
		return "", "", err
	} else {
		return "", "", errors.New("week wrong, it should be between 1 and 52")
	}
}

//获取当年某月的起始日期和结束日期
func GetMonthStartEnd(year int, month int) (startDate string, endDate string, err error) {
	m_str := strconv.Itoa(month)
	if month < 10 {
		m_str = "0" + m_str
	}
	firstday, err := time.Parse(DATE_DAY, strconv.Itoa(year)+"-"+m_str+"-01")
	if err == nil {
		startDate := firstday.Format(DATE_DAY)

		lastday := firstday.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
		endDate = lastday.Format(DATE_DAY)

		return startDate, endDate, nil
	}
	return "", "", err
}

//获取当年某季度的起始日期和结束日期
func GetQuarterStartEnd(year int, quarter int) (startDate string, endDate string, err error) {
	q_arr := []string{"1", "2", "3", "4"}
	index := sort.SearchStrings(q_arr, strconv.Itoa(quarter))

	y := strconv.Itoa(year)
	s := ""
	e := ""

	if index < len(q_arr) && q_arr[index] == strconv.Itoa(quarter) {
		switch quarter {
		case 1:
			s = y + "-01-01"
			e = y + "-03-31"
		case 2:
			s = y + "-04-01"
			e = y + "-06-30"
		case 3:
			s = y + "-07-01"
			e = y + "-09-30"
		case 4:
			s = y + "-10-01"
			e = y + "-12-31"
		}
		return s, e, nil
	}
	return "", "", errors.New("quarter should be 1 or 2 or 3 or 4")
}

//获取起始日期和结束日期之间的天数，包括结束日期当天
func GetDurationDays(startDate string, endDate string) (days int) {
	s, _ := time.Parse(DATE_DAY, startDate)
	e, _ := time.Parse(DATE_DAY, endDate)

	return int(e.Sub(s).Hours()/24) + 1
}

//根据输入参数生成统计数据
// extraField / extraValue 用于给 csv 日志添加额外的字段值，比如该项目 id
func GenerateStat(startDate string, endDate string, svnUrl string, svnDir string, logNamePrefix string, reGenerate bool, csvExport bool, extraField string, extraValue string, author string) (ats statStruct.AuthorTimeStats, as map[string]statStruct.AuthorStat) {
	//生成 svn 日志文件
	svnXmlFile, err := GetSvnLogFile(startDate, endDate, svnUrl, logNamePrefix, reGenerate)

	//获取天数
	if !strings.Contains(startDate, "-") {
		startDate, _ = GetSvnDateByRevision(startDate, svnUrl)
	}
	days := GetDurationDays(startDate, endDate)
	log.Printf("Total %d days during the stats", days)

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

	//获取svn root目录
	svnRoot, _ := GetSvnRoot(svnDir)

	svnXmlLogs, err := ParaseSvnXmlLog(svnXmlFile)
	//	fmt.Printf("%v", svnXmlLogs)
	CheckErr(err)

	authorTimeStats := make(statStruct.AuthorTimeStats)

	AuthorStats := make(map[string]statStruct.AuthorStat)

	//导出 CSV 格式的 log
	if csvExport {
		ExportLogToCsv(svnXmlLogs, startDate, endDate, svnUrl, logNamePrefix, reGenerate, extraField, extraValue, author)
	}

	sd, _ := time.Parse(DATE_SECOND, startDate+DAY_START_SECOND)
	ed, _ := time.Parse(DATE_SECOND, endDate+DAY_END_SECOND)

	filterAuthor := len(author) > 0

	for _, svnXmlLog := range svnXmlLogs.Logentry {
		
		if filterAuthor && svnXmlLog.Author != author {
			continue
		}

		//综合统计
		Author, ok_as := AuthorStats[svnXmlLog.Author]

		Author.StartDate = sd.Format(DATE_MYSQL)
		Author.EndDate = ed.Format(DATE_MYSQL)

		//记录人和日期的详细log，用于细分统计
		authorTimeStat, ok_tss := authorTimeStats[svnXmlLog.Author]
		saveTime, err := time.Parse(DATE_SECOND, svnXmlLog.Date)
		CheckErr(err)
		saveTimeStr := saveTime.Format(DATE_SECOND)

		if !ok_tss { //对象不存在
			authorTimeStat = make(statStruct.AuthorTimeStat)
		}

		//取基于时间的用户操作信息
		AuthorTS, ok_ts := authorTimeStat[saveTimeStr]

		AuthorTS.StartDate = sd.Format(DATE_MYSQL)
		AuthorTS.EndDate = ed.Format(DATE_MYSQL)

		Author.CommitCount += 1
		AuthorTS.CommitCount += 1

		for _, path := range svnXmlLog.Paths {
			// Action = M，修改
			if path.Action == "M" && path.Kind == "file" {
				newRev, _ := strconv.Atoi(svnXmlLog.Revision)

				//修改计数器
				Author.ModifiedFiles += 1
				AuthorTS.ModifiedFiles += 1

				//当前在 svn 工作目录，则可以统计 diff 行数
				if svnRoot != "" {
					log.Printf("svn diff on r%d ,\n", newRev)

					stdout, err := CallSvnDiff(newRev-1, newRev, svnRoot+path.Path)
					if err == nil {
						//fmt.Println("stdout ",stdout)
					} else {
						fmt.Println("err ", err.Error())
					}
					appendLines, removeLines, err := GetLineDiff(stdout)
					log.Printf("\t%s on r%d +%d -%d,\n", path.Path, newRev, appendLines, removeLines)
					if err == nil {
						if ok_as {
							Author.AppendLines += appendLines
							Author.RemoveLines += removeLines
						} else {
							Author.AppendLines = appendLines
							Author.RemoveLines = removeLines
						}
						AuthorStats[svnXmlLog.Author] = Author

						if ok_ts {
							AuthorTS.AppendLines += appendLines
							AuthorTS.RemoveLines += removeLines
						} else {
							AuthorTS.AppendLines = appendLines
							AuthorTS.RemoveLines = removeLines
						}
						authorTimeStat[saveTimeStr] = AuthorTS
					}
				}
			} else if path.Action == "A" && path.Kind == "file" { // Action = A，添加
				//修改计数器
				Author.AddedFiles += 1
				AuthorTS.AddedFiles += 1

				AuthorStats[svnXmlLog.Author] = Author
				authorTimeStat[saveTimeStr] = AuthorTS

			} else if path.Action == "D" && path.Kind == "file" { // Action = D，删除
				//修改计数器
				Author.DeletedFiles += 1
				AuthorTS.DeletedFiles += 1

				AuthorStats[svnXmlLog.Author] = Author
				authorTimeStat[saveTimeStr] = AuthorTS
			}
		}

		authorTimeStats[svnXmlLog.Author] = authorTimeStat
	}

	for name, author := range AuthorStats {
		//保留4位小数
		author.AverageCommitsPerDay, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(author.CommitCount)/float64(days)), 64)

		AuthorStats[name] = author
	}

	return authorTimeStats, AuthorStats
}

func ExportLogToCsv(svnXmlLogs SvnXmlLogs, startDate string, endDate string, svnUrl string, fileNamePrefix string, reGenerate bool, extraField string, extraValue string, author string) {
	pwd, _ := os.Getwd()

	filterAuthor := len(author) > 0

	log_folder := pwd + "/svn_csv_logs/" + fileNamePrefix + "/"
	if !fileExists(log_folder) {
		os.MkdirAll(log_folder, os.FileMode(0755))
	}

	commitLog := log_folder + fileNamePrefix + "_svnlog_" + startDate + "_" + endDate + "_commits.csv"
	pathsLog := log_folder + fileNamePrefix + "_svnlog_" + startDate + "_" + endDate + "_paths.csv"

	//文件已存在且未要求重新生成
	if !reGenerate && fileExists(commitLog) {
		log.Printf("SVN Log CSV file %s already exists", commitLog)
		return
	}

	log.Printf("SVN Log CSV filename is %s, detailed paths log filename is %s\n", commitLog, pathsLog)

	f_commit, err_c := os.OpenFile(commitLog, os.O_WRONLY|os.O_CREATE, 0666)
	f_paths, err_p := os.OpenFile(pathsLog, os.O_WRONLY|os.O_CREATE, 0666)
	if err_c != nil || err_p != nil {
		log.Fatalln("Error: ", err_c, err_p)
		return
	}

	writer_c := csv.NewWriter(f_commit)
	writer_p := csv.NewWriter(f_paths)

	var headerCommit = []string{"revision", "author", "date", "msg", "url"}
	var headerPaths = []string{"revision", "path", "action", "kind", "prop-mods", "text-mods"}
	if extraField != "" {
		if strings.Contains(extraField, ",") {
			fields := strings.Split(extraField, ",")
			headerCommit = append(headerCommit, fields...)
			headerPaths = append(headerPaths, fields...)
		} else {
			headerCommit = append(headerCommit, extraField)
			headerPaths = append(headerPaths, extraField)
		}
	}

	writer_c.Write(headerCommit)
	writer_p.Write(headerPaths)

	for _, svnXmlLog := range svnXmlLogs.Logentry {
		d, _ := time.Parse(DATE_NANOSEC, svnXmlLog.Date)

		if filterAuthor && svnXmlLog.Author != author {
			continue
		}

		var data_c = []string{svnXmlLog.Revision,
			svnXmlLog.Author,
			d.Format(DATE_MYSQL),
			strings.Replace(svnXmlLog.Msg, "\n", ". ", -1),
			svnUrl,
		}
		if extraField != "" {
			if strings.Contains(extraValue, ",") {
				values := strings.Split(extraValue, ",")
				data_c = append(data_c, values...)
			} else {
				data_c = append(data_c, extraValue)
			}
		}

		writer_c.Write(data_c)

		for _, path := range svnXmlLog.Paths {
			var data_p = []string{svnXmlLog.Revision,
				path.Path,
				path.Action,
				path.Kind,
				path.PropMods,
				path.TextMods,
			}
			if extraField != "" {
				if strings.Contains(extraValue, ",") {
					values := strings.Split(extraValue, ",")
					data_p = append(data_p, values...)
				} else {
					data_p = append(data_p, extraValue)
				}
			}

			writer_p.Write(data_p)
		}
	}

	// 将缓存中的内容写入到文件里
	writer_c.Flush()
	writer_p.Flush()

	if err := writer_c.Error(); err != nil {
		log.Fatalln("Commit File Error: ", err)
	}

	if err := writer_p.Error(); err != nil {
		log.Fatalln("Paths File Error: ", err)
	}
}

//将统计结果数据保存到 JSON 文件
func SaveStatsToJson(namePrefix string, subFolder string, startDate string, endDate string, year int, typeName string, typeValue int, reGenerate bool, authorStatsArr [](map[string]statStruct.AuthorStat)) {
	pwd, _ := os.Getwd()

	if subFolder != "" {
		subFolder += "/"
	}

	log_folder := pwd + "/svn_stats/" + namePrefix + "/" + subFolder
	if !fileExists(log_folder) {
		os.MkdirAll(log_folder, os.FileMode(0755))
	}

	log_fullpath := log_folder + namePrefix + "_svnstats_"

	authorNameStats := []statStruct.AuthorNameStat{}
	for _, authorStats := range authorStatsArr {
		for author, authorstat := range authorStats {
			authorNameStat := statStruct.AuthorNameStat{
				Author: author,
				Stat:   authorstat,
			}
			authorNameStats = append(authorNameStats, authorNameStat)
		}
	}

	switch typeName {
	case YEAR_STATS:
		filename := log_fullpath + "year_" + strconv.Itoa(year)
		SaveYearStatsToJsonFile(year, authorNameStats, filename+".json", reGenerate)
	case QUARTER_STATS:
		filename := log_fullpath + "quarter_" + strconv.Itoa(year) + "Q" + strconv.Itoa(typeValue)
		SaveQuarterStatsToJsonFile(year, typeValue, authorNameStats, filename+".json", reGenerate)
	case MONTH_STATS:
		filename := log_fullpath + "month_" + strconv.Itoa(year) + "M" + strconv.Itoa(typeValue)
		SaveMonthStatsToJsonFile(year, typeValue, authorNameStats, filename+".json", reGenerate)
	case WEEK_STATS:
		filename := log_fullpath + "week_" + strconv.Itoa(year) + "W" + strconv.Itoa(typeValue)
		SaveWeekStatsToJsonFile(year, typeValue, authorNameStats, filename+".json", reGenerate)
	default:
		filename := log_fullpath + startDate + "_" + endDate
		SaveCustomStatsToJsonFile(startDate, endDate, authorNameStats, filename+".json", reGenerate)
	}

}

//将统计结果数据保存到 CSV 文件
func SaveStatsToCSV(namePrefix string, subFolder string, startDate string, endDate string, reGenerate bool, authorStatsArr [](map[string]statStruct.AuthorStat), extraField string, extraValue string) {
	pwd, _ := os.Getwd()

	if subFolder != "" {
		subFolder += "/"
	}

	log_folder := pwd + "/svn_stats/" + namePrefix + "/" + subFolder
	if !fileExists(log_folder) {
		os.MkdirAll(log_folder, os.FileMode(0755))
	}

	log_fullpath := log_folder + namePrefix + "_svnstats_"

	authorNameStats := []statStruct.AuthorNameStat{}
	for _, authorStats := range authorStatsArr {
		for author, authorstat := range authorStats {
			authorNameStat := statStruct.AuthorNameStat{
				Author: author,
				Stat:   authorstat,
			}
			authorNameStats = append(authorNameStats, authorNameStat)
		}
	}

	filename := log_fullpath + startDate + "_" + endDate

	SaveStatsToCsvFile(authorNameStats, filename+".csv", reGenerate, extraField, extraValue)

}

func SaveYearStatsToJsonFile(year int, authorStats []statStruct.AuthorNameStat, filepath string, reGenerate bool) {
	//文件已存在且未要求重新生成
	if !reGenerate && fileExists(filepath) {
		log.Printf("Stats file %s already exists", filepath)
		return
	}

	log.Printf("Stats filename is %s\n", filepath)

	yearStats := statStruct.YearStats{}
	yearStats.Year = year

	yearStats.Stats = authorStats

	file, _ := json.MarshalIndent(yearStats, "", " ")

	_ = ioutil.WriteFile(filepath, file, 0644)
}

func SaveQuarterStatsToJsonFile(year int, quarter int, authorStats []statStruct.AuthorNameStat, filepath string, reGenerate bool) {
	//文件已存在且未要求重新生成
	if !reGenerate && fileExists(filepath) {
		log.Printf("Stats file %s already exists", filepath)
		return
	}

	log.Printf("Stats filename is %s\n", filepath)

	quarterStats := statStruct.QuarterStats{}
	quarterStats.Year = year
	quarterStats.Quarter = quarter

	quarterStats.Stats = authorStats

	file, _ := json.MarshalIndent(quarterStats, "", " ")

	_ = ioutil.WriteFile(filepath, file, 0644)
}

func SaveMonthStatsToJsonFile(year int, month int, authorStats []statStruct.AuthorNameStat, filepath string, reGenerate bool) {
	//文件已存在且未要求重新生成
	if !reGenerate && fileExists(filepath) {
		log.Printf("Stats file %s already exists", filepath)
		return
	}

	log.Printf("Stats filename is %s\n", filepath)

	monthStats := statStruct.MonthStats{}

	monthStats.Year = year
	monthStats.Month = month

	monthStats.Stats = authorStats

	file, _ := json.MarshalIndent(monthStats, "", " ")

	_ = ioutil.WriteFile(filepath, file, 0644)
}

func SaveWeekStatsToJsonFile(year int, week int, authorStats []statStruct.AuthorNameStat, filepath string, reGenerate bool) {
	//文件已存在且未要求重新生成
	if !reGenerate && fileExists(filepath) {
		log.Printf("Stats file %s already exists", filepath)
		return
	}

	log.Printf("Stats filename is %s\n", filepath)

	weekStats := statStruct.WeekStats{}

	weekStats.Year = year
	weekStats.Week = week

	weekStats.Stats = authorStats

	file, _ := json.MarshalIndent(weekStats, "", " ")

	_ = ioutil.WriteFile(filepath, file, 0644)
}

func SaveCustomStatsToJsonFile(startDate string, endDate string, authorStats []statStruct.AuthorNameStat, filepath string, reGenerate bool) {
	//文件已存在且未要求重新生成
	if !reGenerate && fileExists(filepath) {
		log.Printf("Stats file %s already exists", filepath)
		return
	}

	log.Printf("Stats filename is %s\n", filepath)

	customStats := statStruct.CustomStats{}

	customStats.StartDate = startDate
	customStats.EndDate = endDate

	customStats.Stats = authorStats

	file, _ := json.MarshalIndent(customStats, "", " ")

	_ = ioutil.WriteFile(filepath, file, 0644)
}

func SaveStatsToCsvFile(authorStats []statStruct.AuthorNameStat, filepath string, reGenerate bool, extraField string, extraValue string) {
	//文件已存在且未要求重新生成
	if !reGenerate && fileExists(filepath) {
		log.Printf("Stats CSV file %s already exists", filepath)
		return
	}

	log.Printf("Stats CSV filename is %s\n", filepath)

	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln("Error: ", err)
		return
	}

	writer := csv.NewWriter(f)
	var header = []string{"Author", "Commits", "AverageCommitsPerDay", "Lines", "FilesAdded", "FilesModified", "FilesDeleted", "StartDate", "EndDate"}
	if extraField != "" {
		if strings.Contains(extraField, ",") {
			fields := strings.Split(extraField, ",")
			header = append(header, fields...)
		} else {
			header = append(header, extraField)
		}
	}
	writer.Write(header)

	for _, authorStat := range authorStats {
		var data = []string{authorStat.Author,
			strconv.Itoa(authorStat.Stat.CommitCount),
			strconv.FormatFloat(authorStat.Stat.AverageCommitsPerDay, 'f', 4, 64),
			strconv.Itoa(authorStat.Stat.AppendLines + authorStat.Stat.RemoveLines),
			strconv.Itoa(authorStat.Stat.AddedFiles),
			strconv.Itoa(authorStat.Stat.ModifiedFiles),
			strconv.Itoa(authorStat.Stat.DeletedFiles),
			authorStat.Stat.StartDate,
			authorStat.Stat.EndDate,
		}
		if extraField != "" {
			if strings.Contains(extraValue, ",") {
				values := strings.Split(extraValue, ",")
				data = append(data, values...)
			} else {
				data = append(data, extraValue)
			}
		}
		writer.Write(data)
	}

	// 将缓存中的内容写入到文件里
	writer.Flush()

	if err = writer.Error(); err != nil {
		log.Fatalln("Error: ", err)
	}
}
