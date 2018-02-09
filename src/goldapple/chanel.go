package goldapple

/**
 * Step3: collect product data
 */

import (
    //"os"
    //"io"
    "fmt"
    "log"
    "time"
    "strings"
    "net/http"
    "io/ioutil"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "golang.org/x/net/html"
    "github.com/blackjack/syslog"
    )

const (
    LetuProducts = "letu_products_final"
    LetuPrice = "letu_price"
    GestoriDB = "gestori"
    LogFile = "Log"
    LogCollection = "Log"
)

// Breadcrumbs
var Navi []string

type LogStruct struct {
    Subject string
    Action string
    Val string
    Date string
}

type Counter struct {
    count_double int
    count_new int
    count_gestori_match int
}

var chanelUrls []string
var letu_root = "https://www.letu.ru"

type SingleUrl struct {
    Url string
}

func Chanel(glob_session *mgo.Session) {

    syslog.Syslog(syslog.LOG_INFO, "Letu step3 start")
    fmt.Println("Letu step3 start")

    var pr *Product
    var results []Link
    var f func(*html.Node, *Product, *BrandSingle)
    var f1 func(*html.Node)
    var churl string

    // Products html table
    f = func(node *html.Node, pr *Product, br *BrandSingle) {
        if node.Type == html.ElementNode && node.Data == "div" {
            for _, a := range node.Attr {
                if a.Key == "onclick" {
                    if strings.Contains(a.Val, "location=") && !strings.Contains(a.Val, "push") {
                        churl = strings.Replace(a.Val, "location=", "", -1)
                        churl = strings.Replace(churl, "'", "", -1)
                        churl = letu_root + churl
                        fmt.Println(churl)
                        chanelUrls = append(chanelUrls, churl)
                    }
                }
            }
        }
        // Iterate inner nodes recursive
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f(c, pr, br)
        }
    }

    // Products html table
    f1 = func(node *html.Node) {
        if node.Type == html.ElementNode && node.Data == "a" {
            for _, a := range node.Attr {
                if a.Key == "href" {
                    if strings.Contains(a.Val, "push") {
                        //fmt.Println(a.Val)
                        //f1()
                        chanelUrls = append(chanelUrls, churl)
                    }
                }
            }
        }
        // Iterate inner nodes recursive
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f1(c)
        }
    }

    // get target pages from mongo
    coll := MakeTimePrefix(LetuCollectionPages)
    if LetuDB == "" {
        LetuDB = "parser"
    }
    c := glob_session.DB(LetuDB).C(coll)
    glob_session.SetMode(mgo.Monotonic, true)
    err := c.Find(bson.M{}).All(&results)

    if err != nil {
        syslog.Critf("Step3 find error: %s", err)
        fmt.Println("Step3 find error", err)
    }

    var httpClient = &http.Client{
        Timeout: time.Second * 2200,
    }
    url_final := "https://www.letu.ru/chanel/"
    fmt.Println("URL:", url_final)
    pr = &Product{Price: "default", Url: url_final}
    resp, err := httpClient.Get(url_final)
    if err != nil {
        syslog.Critf("Chanel httpClient error: %s", err)
        fmt.Println("Chanel httpClient error", err)
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        syslog.Critf("Step3 ioutil.ReadAll error: %s", err)
        fmt.Println("Step3 ioutil.ReadAll error", err)
    }
    doc, err_p := html.Parse(strings.NewReader(string(body)))
    if err_p != nil {
        log.Println(err)
    }
    br := &BrandSingle{Name: "CHANEL"}

    f(doc, pr, br)
    f1(doc)

    i := 0
    for i < len(chanelUrls) {
        fmt.Println(chanelUrls[i])
        i++
    }
    //fmt.Println(chanelUrls)

    syslog.Syslog(syslog.LOG_INFO, "Letu chanel step end")
    fmt.Println("Letu chanel step end")
}