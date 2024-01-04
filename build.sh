#!/bin/bash

basedir=`readlink -f ${0}`

if [ -z ${basedir} ]; then
  exit
fi

basedir=`dirname ${basedir}`

if [ -z ${basedir} ]; then
  exit
fi

cd ${basedir}

if [ -z `which go` ]; then
  echo -e "\033[31;1mUnable to find compiler: golang\033[0m"
  exit
fi

platform=${1}
application=${2}

if [ -z ${platform} ]; then
  platform="amd64"
fi

if [ -z ${application} ]; then
  application="main"
fi

case ${platform} in
  386)
    export GOOS=linux
    export GOARCH=386
    export CGO_ENABLED=1
    ;;
  amd64)
    export GOOS=linux
    export GOARCH=amd64
    export CGO_ENABLED=1
    ;;
  win32)
    export GOOS=windows
    export GOARCH=386
    if [ ! -z `which x86_64-w64-mingw32-gcc` ]; then
      export CC=x86_64-w64-mingw32-gcc
      export CGO_ENABLED=1
    else
      export CGO_ENABLED=0
    fi
    export GOEXE=.exe
    ;;
  win64)
    export GOOS=windows
    export GOARCH=amd64
    if [ ! -z `which x86_64-w64-mingw32-gcc` ]; then
      export CC=x86_64-w64-mingw32-gcc
      export CGO_ENABLED=1
    else
      export CGO_ENABLED=0
    fi
    export GOEXE=.exe
    ;;
  mipsle)
    export GOOS=linux
    export GOARCH=mipsle
    export CGO_ENABLED=0
    export GOMIPS=softfloat
    ;;
  arm)
    export GOOS=linux
    export GOARCH=arm
    export CGO_ENABLED=0
    ;;
  all)
    ${0} 386 ${application} ${3}
    ${0} amd64 ${application} ${3}
    ${0} win32 ${application} ${3}
    ${0} win64 ${application} ${3}
    ${0} mipsle ${application} ${3}
    ${0} arm ${application} ${3}
    exit
    ;;
  *)
    echo -e "\033[31;1munsupported platform: ${platform}\033[0m"
    exit
    ;;
esac

for file in `find ${basedir} -name "*.go"`
do
  go fmt ${file} || exit
done

if [ -d ${basedir}/cmd ]; then
  if [ -f ${basedir}/cmd/main.go ]; then
    go mod tidy && go build -o ${basedir}/${application}${GOEXE} -ldflags "-w -s" ${basedir}/cmd/main.go && mv ${basedir}/${application}${GOEXE} ${basedir}/${application}_${GOOS}_${GOARCH}${GOEXE}
  elif [ -f ${basedir}/cmd/${application}/main.go ]; then
    go mod tidy && go build -o ${basedir}/${application}${GOEXE} -ldflags "-w -s" ${basedir}/cmd/${application}/main.go && mv ${basedir}/${application}${GOEXE} ${basedir}/${application}_${GOOS}_${GOARCH}${GOEXE}
  else
    echo "application folder was not found: ${application}"
    exit
  fi
else
  echo "cmd folder was not found"
  exit
fi

if [ 0 -eq ${?} ] && [ ! -z ${3} ]; then
  if [ ! -z `which upx` ]; then
    upx -9 ${basedir}/${application}_${GOOS}_${GOARCH}${GOEXE}
  fi
fi
