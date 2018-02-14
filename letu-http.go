package main

import (
    "os"
    "fmt"
    "goldapple"
    "gopkg.in/mgo.v2"
    "github.com/gin-gonic/gin"
    "github.com/blackjack/syslog"
)

const (
    C_HOST = "0.0.0.0"
    C_PORT = "7802"
    C_TYPE = "tcp"
)

const ENV_PREF = "prod"
const BrandCollection = "letu_brands"

var DB string = os.Getenv("LETU_MONGO_DB")

func main() {

    r := gin.Default()

    // Start parser
    r.GET("/v1/start", func(c *gin.Context) {

        session, glob_err := mgo.Dial("mongodb://apidev:apidev@localhost:27017/parser")

        if glob_err != nil {
            syslog.Critf("Error: %s", glob_err)
        }

        syslog.Openlog("letu_parser", syslog.LOG_PID, syslog.LOG_USER)
        syslog.Syslog(syslog.LOG_INFO, "Start LETU parser ... ")

        coll := goldapple.MakeTimePrefix(BrandCollection)
        if DB == "" {
            DB = "parser"
        }
        collections := session.DB(DB).C(coll)
        session.SetMode(mgo.Monotonic, true)
        num, err := collections.Count()

        if err != nil {
            syslog.Critf("Error: %s", err)
        }

        if num > 1 {
            syslog.Err("LETU brands allready parsed today")
            fmt.Println("LETU brands allready parsed today")
            c.JSON(200, gin.H{
                "message": "LETU brands allready parsed today",
            })
        } else {
            c.JSON(200, gin.H{
                "message": "r.GET(\"/start\", func(c *gin.Context)",
            })
            goldapple.Chanel(session)
            goldapple.Step1(session)
            goldapple.Step2(session)
            goldapple.Step3(session)
        }

        session.Close()
    })

    // Parse single product page
    r.GET("/v1/single_product", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "r.GET(\"/single_product\", func(c *gin.Context)",
        })
    })

    // Parse single product page
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    r.Run(":"+C_PORT)
}