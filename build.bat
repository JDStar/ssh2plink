@echo off
if not exist build mkdir build
set GOOS=windows
set GOARCH=amd64
go build -o build/ssh2plink_64bit.exe
echo created: build/ssh2plink_64bit.exe

set GOOS=windows
set GOARCH=386
go build -o build/ssh2plink_32bit.exe
echo created: build/ssh2plink_64bit.exe