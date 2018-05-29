#!/bin/sh
BINARYNAME=cirsim
OUTPUTPATH=$GOPATH/bin
PROJECTPATH=$GOPATH/src/github.com/felipeek/cirsim

go build -o $OUTPUTPATH/$BINARYNAME -v $PROJECTPATH/$BINARYNAME/main.go