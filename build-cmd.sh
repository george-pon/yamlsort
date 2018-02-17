#!/bin/bash

basedir=$PWD
echo basedir is $basedir

export GOPATH=$basedir
echo GOPATH is $GOPATH

already=$( echo $PATH | sed -e 's/:/\n/g' | grep "$basedir/bin")
if [ -z "$already" ]; then
    export PATH=$PATH:$basedir/bin
fi

already=$( echo $PATH | sed -e 's/:/\n/g' | grep "$basedir/src/yamlsort")
if [ -z "$already" ]; then
    export PATH=$PATH:$basedir/src/yamlsort
fi

if [ $# -eq 0 ]; then
    mode=build
else
    mode=$1
fi


if [ x"$mode"x = x"jenkinsbuild"x ]; then
    pushd $GOPATH/src/yamlsort
        glide update
        go vet
        go build
    popd
fi


if [ x"$mode"x = x"build"x ]; then
    pushd $GOPATH/src/yamlsort
        go vet
        go build
    popd
fi

