#!/bin/bash

if [ $# -ne 4 ]; then

    echo "Enter startDate (like 2006-01-02), endDate (like 2006-01-03), svnUrl, logName:"
    read startDate endDate svnUrl logName

    if [[ "$startDate" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]; then
        startDate="{"$startDate"}"
    fi

    if [[ "$endDate" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]; then
        endDate="{"$endDate"}"
    fi

    svn log -r $startDate":"$endDate --xml -v $svnUrl > $logName

else

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

fi