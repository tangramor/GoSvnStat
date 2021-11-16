#!/bin/bash

if [ $# -ne 4 ]; then

    echo "Enter startDate, endDate, svnUrl, logName:"
    read startDate endDate svnUrl logName

    svn log -r "{"$startDate"}:{"$endDate"}" --xml -v $svnUrl > $logName

else

    startDate=$1
    endDate=$2
    svnUrl=$3
    logName=$4

    svn log -r "{"$startDate"}:{"$endDate"}" --xml -v $svnUrl > $logName

fi