#!/bin/bash

basedir=$PWD
echo basedir is $basedir

export GOPATH=$basedir
echo GOPATH is $GOPATH

# add $basedir/bin to PATH 
already=$( echo $PATH | sed -e 's/:/\n/g' | grep "$basedir/bin")
if [ -z "$already" ]; then
    export PATH=$PATH:$basedir/bin
fi

# add $basedir/src/yamlsort to PATH 
already=$( echo $PATH | sed -e 's/:/\n/g' | grep "$basedir/src/yamlsort")
if [ -z "$already" ]; then
    export PATH=$PATH:$basedir/src/yamlsort
fi

if [ $# -eq 0 ]; then
    mode=build
else
    mode=$1
fi

# check dep command
type dep
RC=$?
if [ $RC -ne 0 ]; then
    # install dep command
    go get -u github.com/golang/dep/cmd/dep
fi

if [ x"$mode"x = x"depbuild"x ]; then
    pushd $GOPATH/src/yamlsort
        dep init
        go vet
        go install
    popd
fi


if [ x"$mode"x = x"glidebuild"x ]; then
    pushd $GOPATH/src/yamlsort
        glide update
        go vet
        go install
    popd
fi


if [ x"$mode"x = x"build"x ]; then
    pushd $GOPATH/src/yamlsort
        go vet
        go install
    popd
fi

