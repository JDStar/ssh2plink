@echo off
if not exist build mkdir build
set GOOS=windows
set GOARCH=amd64
set dest=build/ssh2plink_64bit.exe
go build -o %dest%
echo created: %dest%

set GOOS=windows
set GOARCH=386
set dest=build/ssh2plink_32bit.exe
go build -o %dest%
echo created: %dest%