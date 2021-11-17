#!/bin/bash

if [ $# -ne 4 ]; then

    echo "Enter startDate (like 2006-01-02), endDate (like 2006-01-03), svnUrl, logName:"
    read startDate endDate svnUrl logName

    svn log -r "{"$startDate"}:{"$endDate"}" --xml -v $svnUrl > $logName

else

    startDate=$1
    endDate=$2
    svnUrl=$3
    logName=$4

    svn log -r "{"$startDate"}:{"$endDate"}" --xml -v $svnUrl > $logName

fi