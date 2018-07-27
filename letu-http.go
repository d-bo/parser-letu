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
    C_PORT = "8804"
    C_TYPE = "tcp"
)

const ENV_PREF = "prod"
const BrandCollection = "letu_brands"

var DB string = os.Getenv("LETU_MONGO_DB")

func main() {

    r := gin.Default()

    // Start parser
    r.GET("/v1/check/:user", func(c *gin.Context) {
        name := c.Param("name")
    })

    // Parse single product page
    r.GET("/v1/single_product", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "r.GET(\"/single_product\", func(c *gin.Context)",
        })
    })

    // Parse single product page
    r.GET("/v1/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    // Parse single product page
    r.GET("/", func(c *gin.Context) {
        c.String(200, "GASec")
    })

    r.Run(":"+C_PORT)
}