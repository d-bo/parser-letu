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

var httpClient = &http.Client{
    Timeout: time.Second * 2200,
}

func Chanel(glob_session *mgo.Session) {

    syslog.Syslog(syslog.LOG_INFO, "Letu step3 start")
    fmt.Println("Letu step3 start")

    var pr *Product
    var results []Link
    var f func(*html.Node, *Product, *BrandSingle)
    var f1 func(*html.Node, *Product, *BrandSingle)
    var f2 func(*html.Node, *Product, *BrandSingle)
    var f3 func(*html.Node, *Product, *BrandSingle)
    var f13 func(*html.Node, *Product)    // Navi href extract
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
                        //fmt.Println(churl)
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

    // Extract product page
    f1 = func(node *html.Node, pr *Product, br *BrandSingle) {
        if node.Type == html.ElementNode && node.Data == "a" {
            for _, a := range node.Attr {
                if a.Key == "href" {
                    if strings.Contains(a.Val, "push") {
                        contents := renderNode(node)
                        if strings.Contains(contents, "img") {
                            fmt.Println("PROD_URL: "+a.Val)
                            //chanelUrls = append(chanelUrls, churl)
                            pr.Url = "https://www.letu.ru"+a.Val
                            resp, err := httpClient.Get(pr.Url)
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

                            f2(doc, pr, br)
                            //fmt.Println(pr)
                        }
                    }
                }
            }
        }

        // Iterate inner nodes recursive
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f1(c, pr, br)
        }
    }

    // Products html table
    f2 = func(node *html.Node, pr *Product, br *BrandSingle) {

        var source string
        var match = false

        if node.Type == html.ElementNode && node.Data == "div" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "breadcrumbs") {
                        f13(node, pr)
                    }
                }
            }
        }
        if node.Type == html.ElementNode && node.Data == "img" {
            for _, a := range node.Attr {
                if a.Key == "src" {
                    source = a.Val
                }
                if a.Key == "itemprop" {
                    if strings.Contains(a.Val, "image") {
                        match = true
                    }
                }
            }
            if match {
                pr.Img = source
                match = false
            }
        }

        if node.Type == html.ElementNode && node.Data == "tr" {
            context := renderNode(node)
            if strings.Contains(context, "addToCart") && strings.Contains(context, "h2-like") {
                f3(node, pr, br)
                fmt.Println("\t", pr)

                // Insert or update
                if LetuDB == "" {
                    LetuDB = "parser"
                }

                c := glob_session.DB(LetuDB).C(LetuProducts)
                d := glob_session.DB(LetuDB).C(MakeTimeMonthlyPrefix(LetuPrice))
                e := glob_session.DB(LetuDB).C(LogCollection)
                glob_session.SetMode(mgo.Monotonic, true)

                // check double
                num, err := c.Find(bson.M{"articul": pr.Articul}).Count()
                if err != nil {
                    syslog.Critf("chanel check double zero err: %s", err)
                    fmt.Println("chanel count double zero err", err)
                }

                pr.Price = strings.Trim(pr.Price, " ")

                price := ProductPrice{
                    Price: pr.Listingprice,
                    Articul: pr.Articul,
                    Brand: br.Name,
                    Date: MakeTimePrefix(""),
                }

                fmt.Println(strings.Join(pr.Navi, ";"))

                if num < 1 {
                    fmt.Println("New:", pr.Articul)

                    /*
                    Name string
                    Articul string
                    Desc string
                    Img string
                    Match_articul string
                    */

                    new := ProductFinal{
                        Articul: pr.Articul,
                        Name: pr.Name,
                        Name_e: pr.Name_e,
                        Desc: pr.Desc,
                        Img: pr.Img,
                        Brand: br.Name,
                        Listingprice: pr.Listingprice,
                        Url: pr.Url,
                        Navi: strings.Join(pr.Navi, ";"),
                        LastUpdate: MakeTimePrefix(""),
                    }
                    // Insert 'letu_products_final'
                    err := c.Insert(new)
                    if err != nil {
                        syslog.Critf("Chanel insert final product error: %s", err)
                        fmt.Println("Chanel insert final product error", err)
                    } else {
                        // Success insert new prod
                        fmt.Println("New prod:", pr.Articul)
                    }
                    // Insert 'letu_price'
                    err = d.Insert(price)
                    if err != nil {
                        syslog.Critf("Chanel insert price error: %s", err)
                        fmt.Println("Chanel insert price error", err)
                    }
                    // Log new product
                    e.Insert(LogStruct{
                        Subject: "letu",
                        Action: "new_articul",
                        Val: pr.Articul,
                        Date: MakeTimePrefix(""),
                    })
                } else {
                    fmt.Println("DOUBLE:", pr.Articul)
                    // DOUBLE ??
                    // Update price column
                    change := mgo.Change{
                        Update: bson.M{
                            "$set": bson.M{
                                "listingprice": pr.Price,
                                "Name_e": pr.Name_e,
                                // as of fixed 24.10.17
                                "desc": pr.Desc,
                                "img": pr.Img,
                                "url": pr.Url,
                                "Navi": strings.Join(pr.Navi, ";"),
                                "LastUpdate": MakeTimePrefix(""),
                            },
                        },
                        ReturnNew: true,
                    }
                    doc := ProductFinal{}
                    c.Find(bson.M{"articul": pr.Articul}).Apply(change, &doc)
                    // insert 'letu_price' on double
                    err = d.Insert(price)
                    if err != nil {
                        syslog.Critf("Chanel insert on double price error: %s", err)
                        fmt.Println("Chanel insert on double price error", err)
                    }
                }

                pr.Navi = []string{}
            }
        }
        // Iterate inner nodes recursive
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f2(c, pr, br)
        }
    }

    f3 = func(node *html.Node, pr *Product, br *BrandSingle) {

        var source string
        var match = false

        if node.Type == html.ElementNode && node.Data == "div" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "h2-like") {
                        contents := renderNode(node)
                        contents = extractContext(contents)
                        pr.Name = contents
                        //fmt.Println(pr.Name)
                    }
                }
            }
        }
        if node.Type == html.ElementNode && node.Data == "p" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "description") {
                        contents := renderNode(node)
                        contents = extractContext(contents)
                        pr.Name_e = contents
                        //fmt.Println(pr.Desc)
                    }
                }
            }
        }
        if node.Type == html.ElementNode && node.Data == "div" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "productDecsrip") {
                        contents := renderNode(node)
                        contents = extractContext(contents)
                        pr.Desc = contents
                        //fmt.Println(pr.Desc)
                    }
                }
            }
        }
        if node.Type == html.ElementNode && node.Data == "p" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "price_no_discount") {
                        contents := renderNode(node)
                        contents = extractContext(contents)
                        pr.Listingprice = contents
                        //fmt.Println(pr.Listingprice)
                    }
                }
            }
        }
        if node.Type == html.ElementNode && node.Data == "input" {
            for _, a := range node.Attr {
                if a.Key == "productid" {
                    source = a.Val
                }
                if a.Key == "class" {
                    if strings.Contains(a.Val, "atg_behavior_addItemToCart") {
                        match = true
                    }
                }
                if match {
                    pr.Articul = source
                    match = false
                }
            }
        }

        // Iterate inner nodes recursive
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f3(c, pr, br)
        }
    }

    // Extract href tags
    f13 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "a" {
            for _, a := range node.Attr {
                if a.Key == "href" {
                    pre := renderNode(node)
                    pre = extractContext(pre)
                    pr.Navi = append(pr.Navi, pre)
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f13(c, pr)
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

    url_final := "https://www.letu.ru/chanel/"
    pr = &Product{Brand: "CHANEL"}
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
    //f1(doc)

    i := 0
    for i < len(chanelUrls) {
        resp, err := httpClient.Get(chanelUrls[i])
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
        fmt.Println("URL:"+chanelUrls[i])
        f1(doc, pr, br)
        i++
    }
    //fmt.Println(chanelUrls)

    syslog.Syslog(syslog.LOG_INFO, "Letu chanel step end")
    fmt.Println("Letu chanel step end")
}