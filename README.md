## Install

```bash
cd parser_letu
export GOPATH=`pwd`
go get golang.org/x/net/html
go get gopkg.in/mgo.v2
go get github.com/blackjack/syslog
# goldapple pkg
go build goldapple
go install goldapple
go run tcp-server.go
```

## Cron

```bash
crontab -e
```

```bash
# everyday at 11:00

00 11 * * * echo -n "start"|netcat 127.0.0.1 8800
```
