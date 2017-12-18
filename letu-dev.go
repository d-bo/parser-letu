package main

// Go letu parser service
// TODO: single mongo connection, close on SIGTERM, SIGINT

import (
    "os"
    "fmt"
    "time"
    "goldapple"
    "gopkg.in/mgo.v2"
    "github.com/blackjack/syslog"
)

const (
    C_HOST = "0.0.0.0"
    C_PORT = "8800"
    C_TYPE = "tcp"
    LetuBrandCollection = "letu_brands"
)

var ENV_PREF = "dev"
var LetuDB string = os.Getenv("LETU_MONGO_DB")

// A time prefix before collection name
func makeTimePrefix(coll string) string {
    t := time.Now()
    ti := t.Format("02-01-2006")
    if coll == "" {
        return ti
    }
    fin := ti + "_" + coll
    return fin
}

// goroutine handler
func main() {

    // Inject variable to goldapple pkg
    goldapple.ENV_PREF = ENV_PREF

    syslog.Openlog("letu_parser"+ENV_PREF, syslog.LOG_PID, syslog.LOG_USER)
    syslog.Syslog(syslog.LOG_INFO, "Start LETU parser ... " + C_HOST + ":" + C_PORT)

    session, err := mgo.Dial("mongodb://localhost:27017/")
    if err != nil {
        //fmt.Println("Mongo connection error: ", err)
        //syslog.Critf("Rive error: %s", err)
        syslog.Err("Mongo connection refused. Not started ?")
        os.Exit(1)
    }

    coll := makeTimePrefix(LetuBrandCollection)
    if LetuDB == "" {
        LetuDB = "parser"
    }
    c := session.DB(LetuDB).C(coll)
    session.SetMode(mgo.Monotonic, true)
    num, err := c.Count()

    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("Today allready parsed count:", num)
    }

    //goldapple.Step1(session)
    //goldapple.Step2(session)
    goldapple.Step3(session)
}