#!/bin/bash

cd $1
# go step1-3 will catch this
export ILDE_MONGO_DB="parser"
export GOPATH=$1
BASEDIR=$1

echo -e "\e[44mSTART LETU\e[0m"
start=`date +%s`
go run $BASEDIR"/step1.go"
go run $BASEDIR"/step2.go"
go run $BASEDIR"/step3.go"
end=`date +%s`

runtime=$((end-start))
echo -e "\e[42mLETU FINISHED\e[0m"
echo "Script exec time: $runtime"
