#!/bin/bash
v=$1

pushd ../msuwav
GOOS=windows GOARCH=amd64 go build
popd

pushd ../wavmsu
GOOS=windows GOARCH=amd64 go build
popd

mkdir msu-tools-$v
pushd msu-tools-$v
cp ../../msuwav/msuwav.exe .
cp ../../wavmsu/wavmsu.exe .
popd

zip msu-tools-$v.zip msu-tools-$v/*
