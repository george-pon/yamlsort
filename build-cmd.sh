#!/bin/bash

VERSION=$(git describe)

basedir=$PWD
echo basedir is $basedir

export GOPATH=$basedir
echo GOPATH is $GOPATH

# add $basedir/bin to PATH 
if ! echo $PATH | sed -e 's/:/\n/g' | grep "$basedir/bin" > /dev/null ; then
    echo "append  $basedir/bin  to PATH"
    export PATH=$PATH:$basedir/bin
fi

# add $basedir/src/yamlsort to PATH 
if ! echo $PATH | sed -e 's/:/\n/g' | grep "$basedir/src/yamlsort" > /dev/null ; then
    echo "append  $basedir/src/yamlsort  to PATH"
    export PATH=$PATH:$basedir/src/yamlsort
fi

if [ $# -eq 0 ]; then
    mode=modbuild
else
    mode=$1
fi

if [ x"$mode"x = x"modbuild"x ]; then
    # for golang 1.11
    export GO111MODULE=on
    pushd $GOPATH/src/yamlsort
        while true
        do
            if [ ! -r $basedir/src/yamlsort/go.mod ]; then
                go mod init
                RC=$? ; if [ $RC -ne 0 ]; then break ; fi
            fi

            go vet
            RC=$? ; if [ $RC -ne 0 ]; then break ; fi

            go install -ldflags "-X main.version=$(git describe)"
            RC=$? ; if [ $RC -ne 0 ]; then break ; fi

            GOOS=windows GOARCH=amd64 go build   -ldflags "-X main.version=$(git describe)"    -o ../../bin/windows_amd64_yamlsort.exe
            RC=$? ; if [ $RC -ne 0 ]; then break ; fi

            GOOS=linux GOARCH=amd64 go build   -ldflags "-X main.version=$(git describe)"    -o ../../bin/linux_amd64_yamlsort
            RC=$? ; if [ $RC -ne 0 ]; then break ; fi

            GOOS=freebsd GOARCH=amd64 go build   -ldflags "-X main.version=$(git describe)"    -o ../../bin/freebsd_amd64_yamlsort
            RC=$? ; if [ $RC -ne 0 ]; then break ; fi

            pushd ../../bin
            zip -u windows_amd64_yamlsort_${VERSION}.zip windows_amd64_yamlsort.exe
            zip -u linux_amd64_yamlsort_${VERSION}.zip linux_amd64_yamlsort
            zip -u freebsd_amd64_yamlsort_${VERSION}.zip freebsd_amd64_yamlsort
            popd

            break
        done
    popd
fi

if [ x"$mode"x = x"depbuild"x ]; then
    # check dep command
    if ! type dep 2>/dev/null ; then
        # install dep command
        go get -u github.com/golang/dep/cmd/dep
    fi
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

echo "SUCCESS."
