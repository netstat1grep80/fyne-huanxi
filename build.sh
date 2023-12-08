#!/bin/bash

# 判断是否有足够的参数
if [ $# -lt 1 ]; then
    echo "Usage: $0 <parameter>"
    exit 1
fi

# 获取第一个参数的值
first_param=$1

# 使用 if 语句判断第一个参数的值
if [ "$first_param" == "linux" ]; then
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app.linux
    echo "Linux 编译完成"
elif [ "$first_param" == "windows" ]; then
    	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOARCH=amd64 fyne package -os windows -icon favicon.ico
	echo "Windows 编译完成"
else
    CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o huanxi
    echo "MacOS 编译完成"
fi

