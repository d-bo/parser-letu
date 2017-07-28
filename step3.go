package main

/**
 * Step3: collect product data
 */

import (
    "os"
    "io"
    "log"
    "fmt"
    "time"
    "bytes"
    "strings"
    "net/http"
    "io/ioutil"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "golang.org/x/net/html"
    )

const LetuRootUrl string = "https://www.letu.ru"
const LetuCollectionPages = "letu_pages"
const LetuProducts = "letu_products"
const GestoriDB = "gestori_db"

var LetuDB string = os.Getenv("LETU_MONGO_DB")

// single link product page
type Link struct {
    Link string
}

// Product
type Product struct {
    Price string
    Price_discount string
    Name string
    Articul string
    Desc string
    Img string
    Match_articul string
}

type Gestori struct {
    _id string
    Name_e string
    Cod_good string
    Retail_price string
    Barcod string
}

// Link pool
var LinkPool []Link

var glob_session, glob_err = mgo.Dial("mongodb://localhost:27017/")

func makeTimePrefix(coll string) string {
    t := time.Now()
    ti := t.Format("02-01-2006")
    fin := ti + "_" + coll
    return fin
}

// Render node
func renderNode(node *html.Node) string {
    var buf bytes.Buffer
    w := io.Writer(&buf)
    err := html.Render(w, node)
    if err != nil {
        log.Fatal(err)
    }
    return buf.String()
}

// Get tag context
// TODO: prevent endless loop
func extractContext(s string) string {
    z := html.NewTokenizer(strings.NewReader(s))
    for {
        tt := z.Next()
        switch tt {
            case html.ErrorToken:
                fmt.Println(z.Err())
                continue
            case html.TextToken:
                text := string(z.Text())
                return text
        }
    }
}

func main() {

    var i int
    var pr *Product
    var results []Link
    var f func(*html.Node, *Product)
    var f1 func(*html.Node, *Product)
    var f2 func(*html.Node, *Product)
    var f3 func(*html.Node, *Product)
    var f4 func(*html.Node, *Product)
    var f5 func(*html.Node, *Product)
    var f6 func(*html.Node, *Product)



    f = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "table" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "atg_store_productSummary") {
                        f1(node, pr)
                    }
                }
            }
        }
        // iterate inner nodes recursive
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f(c, pr)
        }
    }
    // extract price
    f2 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "p" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "price_no_discount") {
                        pre := renderNode(node)
                        pre = extractContext(pre)
                        pr.Price = pre
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f2(c, pr)
        }
    }
    // extract articul
    f3 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "p" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "article") {
                        pre := renderNode(node)
                        pre = extractContext(pre)
                        pre = strings.Replace(pre, "Артикул", "", -1)
                        pre = strings.TrimSpace(pre)
                        pr.Articul = pre
                        fmt.Println("Articul: ", pre)
                        
                        // gestori match
                        /*
                        var gestres []Gestori
                        c := glob_session.DB(LetuDB).C(LetuProducts)
                        glob_session.SetMode(mgo.Monotonic, true)
                        err := c.Find(bson.M{"Artic": pre}).One(&gestres)
                        if err != nil {
                            panic(err)
                        } else {
                            fmt.Println("GESTORI MATCH: ", gestres)
                        }
                        */
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f3(c, pr)
        }
    }
    // extract articul
    f4 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "div" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "h2-like") {
                        pre := renderNode(node)
                        pre = extractContext(pre)
                        pre = strings.TrimSpace(pre)
                        pr.Name = pre
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f4(c, pr)
        }
    }
    // extract articul
    f5 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "p" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "description") {
                        pre := renderNode(node)
                        pre = extractContext(pre)
                        pre = strings.TrimSpace(pre)
                        pr.Desc = pre
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f5(c, pr)
        }
    }
    // extract articul
    f6 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "img" {
            for _, a := range node.Attr {
                //key := strings.TrimSpace(a.Val)
                if strings.Contains(a.Val, "src") {
                    //value := strings.TrimSpace(a.Val)
                }
                if a.Key == "itemprop" {
                    //if strings.Contains(a.Val, "jpg") {
                    if a.Val == "image" {
                        //match = true
                        /*
                        pre := renderNode(node)
                        pre = extractContext(pre)
                        pre = strings.TrimSpace(pre)
                        pr.Img = pre
                        fmt.Println(pr)
                        */
                    }
                }
            }
        }

        i = 0
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f6(c, pr)
        }
    }
    // found product container
    f1 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "tr" {
            fmt.Println("Product node found")
            f2(node, pr)	// price
            f3(node, pr)	// article
            f4(node, pr)	// name
            f5(node, pr)	// desc

            if pr.Price != "default" {
                coll := makeTimePrefix(LetuProducts)
                if LetuDB == "" {
                    LetuDB = "parser"
                }
                c := glob_session.DB(LetuDB).C(coll)
                glob_session.SetMode(mgo.Monotonic, true)
                err := c.Insert(pr)
                if err != nil {
                    fmt.Println("shit")
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f1(c, pr)
        }
    }

    start := time.Now()
    defer glob_session.Close()

    // get target pages from mongo
    coll := makeTimePrefix(LetuCollectionPages)
    if LetuDB == "" {
        LetuDB = "parser"
    }
    c := glob_session.DB(LetuDB).C(coll)
    glob_session.SetMode(mgo.Monotonic, true)
    err := c.Find(bson.M{}).All(&results)

    if glob_err != nil {
        log.Fatal(err)
    }

    fmt.Println(len(results))

    i = 0
    for _, v := range results {
        pr = &Product{Price: "default"}
        var httpClient = &http.Client{
            Timeout: time.Second * 10,
        }
        url_final := LetuRootUrl + v.Link
        resp, err := httpClient.Get(url_final)
        if err != nil {
            fmt.Println(err)
            continue
        }
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Println(err)
        }
        doc, err_p := html.Parse(strings.NewReader(string(body)))
        if err_p != nil {
            log.Println(err)
        }
        fmt.Println(url_final, doc, "\r\n")
        f(doc, pr)
        i++
    }

    if glob_err != nil {
        log.Fatal(glob_err)
    }

    elapsed := time.Since(start)
    fmt.Printf("Script took %s", elapsed)
}
