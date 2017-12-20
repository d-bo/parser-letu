package goldapple

/**
 * Step2: Extract product url
 */

import (
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

var LetuCollectionPages = "letu_pages"

type BrandPass struct {
    Name string
}

// single link product page
type Link struct {
    Link string
    Brand string
}

// Link pool
var LinkPool []Link

// Brand pool
var BrandPoolResult []Brand

func Step2(glob_session *mgo.Session) {

    syslog.Syslog(syslog.LOG_INFO, "Letu step2 start")
    fmt.Println("Letu step2 start")

    // html parser itself
    var f func(*html.Node, *BrandPass)
    f = func(node *html.Node, br *BrandPass) {
        match := false
        var value string
        if node.Type == html.ElementNode && node.Data == "a" {
            for _, a := range node.Attr {
                if a.Key == "href" {
                    value = a.Val
                }
                if a.Key == "class" {
                    if strings.Contains(a.Val, "atg_store_basicButton") {
                        match = true
                    }
                }
            }
            if match && !strings.Contains(value, "javascript") {
                pre := renderNode(node)
                pre = extractContext(pre)
                b := Link{value, br.Name}
                LinkPool = append(
                    LinkPool,
                    b,
                )

                coll := makeTimePrefix(LetuCollectionPages)
                if LetuDB == "" {
                    LetuDB = "parser"
                }
                c := glob_session.DB(LetuDB).C(coll)
                glob_session.SetMode(mgo.Monotonic, true)
                err := c.Insert(b)

                if err != nil {
                    syslog.Critf("Step2 link insert error: %s", err)
                    fmt.Println("Step2 link insert error", err)
                }
            }
        }

        // iterate inner nodes recursive
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f(c, br)
        }
    }

    var results []Brand

    // get target pages from mongo
    if LetuDB == "" {
        LetuDB = "parser"
    }
    coll := makeTimePrefix(LetuCollection)
    c := glob_session.DB(LetuDB).C(coll)
    glob_session.SetMode(mgo.Monotonic, true)
    err := c.Find(bson.M{}).All(&results)

    if err != nil {
        syslog.Critf("Step2 find error: %s", err)
    }

    // uncomment this if you want to start from target brand
    //match_flag := 0

    for _, v := range results {
        url_final := LetuRootUrl+ v.Link + "&Nrpp=6000"
        fmt.Println(url_final)

        if !strings.Contains(url_final, "q_brandId") {
            continue
        } else {
            // uncomment this if you want to start from target brand
            /*
            if match_flag == 0 {
                if strings.Contains(v.Link, "/browse/brandProducts.jsp?q_brandId=192001&N=4146502249") {
                    match_flag = 1
                } else {
                    continue
                }
            }
            */
        }

        var httpClient = &http.Client{
            Timeout: time.Second * 2200,
        }

        resp, err := httpClient.Get(url_final)
        if err != nil {
            syslog.Critf("Step2 httpClient get error: %s", err)
            fmt.Println("Step2 httpClient get error", err)
            continue
        }

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            syslog.Critf("Step2 insert error: %s", err)
            fmt.Println("Step2 insert error", err)
        }

        doc, err_p := html.Parse(strings.NewReader(string(body)))
        if err_p != nil {
            log.Println(err)
        }

        br := &BrandPass{Name: v.Name}
        f(doc, br)
    }

    syslog.Syslog(syslog.LOG_INFO, "Letu step2 end")
    fmt.Println("Letu step2 end")
}