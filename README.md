## Install

```bash
$ cd parser_letu
$ export GOPATH=`pwd`
$ go get golang.org/x/net/html 
$ go get gopkg.in/mgo.v2
$ time ./start-letu.sh
```

## Cron

```bash
crontab -e
```

```bash
# everyday at 11:00
# 00 11 * * 1 source <PROJECT_PATH>/start-letu.sh <PROJECT_PATH>

00 11 * * 1 source /home/administrator/parser_letu/start-letu.sh /home/administrator/parser_letu
```

