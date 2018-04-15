@echo off

set BINARYNAME=cirsim
set OUTPUTPATH=%GOPATH%\bin
set PROJECTPATH=%GOPATH%\src\github.com\felipeek\cirsim

go build -o %OUTPUTPATH%\%BINARYNAME%.exe -v %PROJECTPATH%\%BINARYNAME%\main.go