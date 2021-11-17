//工具类，封装若干工具方法
package util

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
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

	//不强制重新生成日志，结束日期不是今天且日志文件存在，则不重新生成
	if regenerate == "n" && endDate != now.Format("2006-01-02") && fileExists(log_fullpath) {
		log.Printf("Log file %s already exists", log_fullpath)
		return log_fullpath, nil
	}

	log.Printf("Log filename is %s\n", log_name)

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
