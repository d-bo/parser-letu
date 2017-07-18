#!/bin/bash

# mongodb prefix
PRFX="letu_"
# go step1-3 will catch this
export LETU_MONGO_DB=$PRFX`date +%F`

echo -e "\e[44m"$LETU_MONGO_DB"\e[0m"
start=`date +%s`
go run step1.go
go run step2.go
go run step3.go
end=`date +%s`

runtime=$((end-start))
echo -e "\e[42mLETU FINISHED\e[0m"
echo "Script exec time: $runtime"
