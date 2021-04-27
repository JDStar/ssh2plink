#! /bin/bash
mkdir -p build
env GOOS=windows GOARCH=amd64 go build -o build/ssh2plink_64bit.exe
env GOOS=windows GOARCH=386 go build -o build/ssh2plink_32bit.exe
