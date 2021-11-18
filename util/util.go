//工具类，封装若干工具方法
package util

import (
	"GoSvnStat/statStruct"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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
	Action string `xml:"action,attr"`
	Kind   string `xml:"kind,attr"`
	Path   string `xml:",chardata"`
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

//获取svn日志
func GetSvnLogFile(startDate string, endDate string, svnUrl string, namePrefix string, regenerate string) (svnLogFile string, err error) { /*{{{*/
	pwd, _ := os.Getwd()

	now := time.Now()

	if startDate == "" {
		startDate = now.AddDate(0, 0, -1).Format("2006-01-02")
		log.Printf("Start Date is %s\n", startDate)
	}

	if endDate == "" {
		endDate = now.Format("2006-01-02")
		log.Printf("End Date is %s\n", endDate)
	}

	if namePrefix == "" {
		namePrefix = "Temp"
	}

	log_folder := pwd + "/svn_logs/"
	if !fileExists(log_folder) {
		os.MkdirAll(log_folder, os.FileMode(0755))
	}

	log_name := namePrefix + "_svnlog_" + startDate + "_" + endDate + ".log"
	log_fullpath := log_folder + log_name

	log.Printf("Log filename is %s\n", log_name)

	//不强制重新生成日志，结束日期不是今天且日志文件存在，则不重新生成
	if regenerate == "n" && endDate != now.Format("2006-01-02") && fileExists(log_fullpath) {
		log.Printf("Log file %s already exists", log_fullpath)
		return log_fullpath, nil
	}

	app := "./GenerateSvnLog.sh"
	param1 := startDate
	param2 := endDate
	param3 := svnUrl
	param4 := log_fullpath

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
		firstday, err := time.Parse("2006-01-02", strconv.Itoa(year)+"-01-01")
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
			startDate := weekMonday.Format("2006-01-02")
			weekSunday := weekMonday.AddDate(0, 0, 6)
			endDate := weekSunday.Format("2006-01-02")

			return startDate, endDate, nil
		}
		return "", "", err
	} else {
		return "", "", errors.New("week wrong, it should be between 1 and 52")
	}
}

//获取当年某月的起始日期和结束日期
func GetMonthStartEnd(year int, month int) (startDate string, endDate string, err error) {
	firstday, err := time.Parse("2006-01-02", strconv.Itoa(year)+"-"+strconv.Itoa(month)+"-01")
	if err == nil {
		startDate := firstday.Format("2006-01-02")

		lastday := firstday.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
		endDate = lastday.Format("2006-01-02")

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
	s, _ := time.Parse("2006-01-02", startDate)
	e, _ := time.Parse("2006-01-02", endDate)

	return int(e.Sub(s).Hours()/24) + 1
}

func GenerateStat(startDate string, endDate string, svnUrl string, svnDir string, logNamePrefix string, reGenerate string) (ats statStruct.AuthorTimeStats, as map[string]statStruct.AuthorStat) {
	//获取天数
	days := GetDurationDays(startDate, endDate)
	log.Printf("Total %d days during the stats", days)

	//生成 svn 日志文件
	svnXmlFile, err := GetSvnLogFile(startDate, endDate, svnUrl, logNamePrefix, reGenerate)

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

	for _, svnXmlLog := range svnXmlLogs.Logentry {
		//综合统计
		Author, ok_as := AuthorStats[svnXmlLog.Author]

		//记录人和日期的详细log，用于细分统计
		authorTimeStat, ok_tss := authorTimeStats[svnXmlLog.Author]
		saveTime, err := time.Parse("2006-01-02T15:04:05Z", svnXmlLog.Date)
		CheckErr(err)
		saveTimeStr := saveTime.Format(DATE_SECOND)

		if !ok_tss { //对象不存在
			authorTimeStat = make(statStruct.AuthorTimeStat)
		}

		//取基于时间的用户操作信息
		AuthorTS, ok_ts := authorTimeStat[saveTimeStr]

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
		author.AverageCommitsPerDay = int(math.Ceil(float64(author.CommitCount) / float64(days)))
		AuthorStats[name] = author
	}

	return authorTimeStats, AuthorStats
}
