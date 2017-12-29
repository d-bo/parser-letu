package goldapple

/**
 * Step3: collect product data
 */

import (
    "os"
    "io"
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

type BrandSingle struct {
    Name string
}

// Product
type Product struct {
    Price string
    Price_discount string
    Name string
    Articul string
    Desc string
    Img string
    Gestori string `json:"gestori,omitempty" bson:"gestori,omitempty"`
    Brand string
    Listingprice string
    Discountprice string
    Oldprice string
    Volume string
    Url string
}

// ProductFinal
type ProductFinal struct {
    Name string
    Articul string
    Desc string
    Img string
    Gestori string `json:"gestori,omitempty" bson:"gestori,omitempty"`
    Brand string
    Listingprice string
    Discountprice string
    Oldprice string
    Volume string
    Url string
}

// struct for ILDE_price
type ProductPrice struct {
    Price string
    Price_discount string
    Oldprice string
    Articul string
    Brand string
}

type Gestori struct {
    _id string
    Name_e string
    Cod_good string
    Retail_price string
    Barcod string
}

func Log(msg []byte) {
    f, err := os.OpenFile("log/"+makeTimePrefix(LogFile), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0775)
    defer f.Close()
    if err != nil {
        syslog.Critf("Step3 openfile error: %s", err)
        fmt.Println("Step3 openfile error", err)
    }
    bytemsg := []byte(msg)
    n, err := f.Write(bytemsg)
    if err == nil && n < len(bytemsg) {
        syslog.Critf("Step3 logwrite error: %s", io.ErrShortWrite)
    }
}

func makeTimeMonthlyPrefix(coll string) string {
    t := time.Now()
    ti := t.Format("01-2006")
    fin := ti + "_" + coll
    return fin
}

func Step3(glob_session *mgo.Session) {

    syslog.Syslog(syslog.LOG_INFO, "Letu step3 start")
    fmt.Println("Letu step3 start")

    var i int
    var pr *Product
    var results []Link
    var f func(*html.Node, *Product, *BrandSingle)
    var f1 func(*html.Node, *Product, *BrandSingle)
    var f2 func(*html.Node, *Product)
    var f3 func(*html.Node, *Product)
    var f4 func(*html.Node, *Product)
    var f5 func(*html.Node, *Product)
    var f6 func(*html.Node, *Product)
    var f7 func(*html.Node, *Product)
    var f8 func(*html.Node, *Product)
    var f9 func(*html.Node, *Product)
    var f10 func(*html.Node, *Product)
    var f11 func(*html.Node, *Product)



    // Products html table
    f = func(node *html.Node, pr *Product, br *BrandSingle) {
        if node.Type == html.ElementNode && node.Data == "table" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "atg_store_productSummary") {
                        f1(node, pr, br)
                    }
                }
            }
        }
        // iterate inner nodes recursive
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f(c, pr, br)
        }
    }

    // Found product container
    f1 = func(node *html.Node, pr *Product, br *BrandSingle) {
        if node.Type == html.ElementNode && node.Data == "tr" {
            f2(node, pr)    // price
            f3(node, pr)    // article
            f4(node, pr)    // name
            f5(node, pr)    // desc
            f8(node, pr)    // old price
            f9(node, pr)    // new price
            f10(node, pr)   // new price double check

            if pr.Price != "default" {
                if LetuDB == "" {
                    LetuDB = "parser"
                }

                c := glob_session.DB(LetuDB).C(LetuProducts)
                d := glob_session.DB(LetuDB).C(makeTimeMonthlyPrefix(LetuPrice))
                e := glob_session.DB(LetuDB).C(LogCollection)
                glob_session.SetMode(mgo.Monotonic, true)

                // check double
                num, err := c.Find(bson.M{"articul": pr.Articul}).Count()
                if err != nil {
                    syslog.Critf("Step3 check double zero err: %s", err)
                    fmt.Println("Step3 count double zero err", err)
                }

                price := ProductPrice{
                    Price: pr.Price,
                    Price_discount: pr.Price_discount,
                    Articul: pr.Articul,
                    Brand: br.Name,
                }

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
                        Desc: pr.Desc,
                        Img: pr.Img,
                        Gestori: pr.Gestori,
                        Brand: br.Name,
                        Listingprice: pr.Price,
                        Volume: pr.Volume,
                        Url: pr.Url,
                    }
                    // Insert 'letu_products_final'
                    err := c.Insert(new)
                    if err != nil {
                        syslog.Critf("Step3 insert final product error: %s", err)
                        fmt.Println("Step3 insert final product error", err)
                    } else {
                        // Success insert new prod
                        fmt.Println("New prod:", pr.Articul)
                    }
                    // Insert 'letu_price'
                    err = d.Insert(price)
                    if err != nil {
                        syslog.Critf("Step3 insert price error: %s", err)
                        fmt.Println("Step3 insert price error", err)
                    }
                    // Log new product
                    e.Insert(LogStruct{
                        Subject: "letu",
                        Action: "new_articul",
                        Val: pr.Articul,
                        Date: makeTimePrefix(""),
                    })
                } else {
                    fmt.Println("DOUBLE:", pr.Articul)
                    // DOUBLE ??
                    // Update price column
                    change := mgo.Change{
                        Update: bson.M{
                            "$set": bson.M{
                                "listingprice": pr.Price,
                                // as of fixed 24.10.17
                                "desc": pr.Desc,
                                "volume": pr.Volume,
                                "img": pr.Img,
                                "url": pr.Url,
                            },
                        },
                        ReturnNew: true,
                    }
                    doc := ProductFinal{}
                    c.Find(bson.M{"articul": pr.Articul}).Apply(change, &doc)
                    // insert 'letu_price' on double
                    err = d.Insert(price)
                    if err != nil {
                        syslog.Critf("Step3 insert on double price error: %s", err)
                        fmt.Println("Step3 insert on double price error", err)
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f1(c, pr, br)
        }
    }

    // Extract no discount price
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

    // Extract articul
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
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f3(c, pr)
        }
    }

    // Extract name
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

    // Extract volume
    f5 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "p" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "description") {
                        pre := renderNode(node)
                        pre = extractContext(pre)
                        pre = strings.TrimSpace(pre)
                        pr.Volume = pre
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f5(c, pr)
        }
    }

    // Extract image
    f6 = func(node *html.Node, pr *Product) {
        src := ""
        match := false
        _ = match
        _ = src
        if node.Type == html.ElementNode && node.Data == "img" {
            for _, a := range node.Attr {
                if a.Key == "itemprop" {
                    if strings.Contains(a.Val, "image") {
                        match = true
                    }
                }
                if a.Key == "src" {
                    if strings.Contains(a.Val, "jpg") {
                        src = a.Val
                    }
                }
            }
            if match == true {
                pr.Img = src
            }
            match = false
        }
        i = 0
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f6(c, pr)
        }
    }

    // Extract description
    f7 = func(node *html.Node, pr *Product) {
        dsc := ""
        match := false
        _ = match
        _ = dsc
        if node.Type == html.ElementNode && node.Data == "div" {
            for _, a := range node.Attr {
                if a.Key == "itemprop" {
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
            f7(c, pr)
        }
    }

    // Extract old price
    // If exist
    f8 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "p" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "old_price") {
                        pre := renderNode(node)
                        pre = extractContext(pre)
                        pre = strings.Replace(pre, "&nbsp;", "", -1)
                        // Dbg
                        if ENV_PREF == "dev" {
                            fmt.Println("Found old_price", pre)
                        }
                        pr.Oldprice = pre
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f8(c, pr)
        }
    }

    // Extract new_price
    // If exist
    f9 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "p" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "new_price") {
                        pre := renderNode(node)
                        pre = extractContext(pre)
                        pre = strings.Replace(pre, "&nbsp;", "", -1)
                        pre = strings.Replace(pre, "\n", "", -1)
                        pre = strings.Replace(pre, "<span class=\"star_for_discounted_price\">*</span>", "", -1)
                        pre = strings.TrimLeft(pre, " ")
                        // Dbg
                        if ENV_PREF == "dev" {
                            fmt.Println("Found new_price", pre)
                        }
                        pr.Listingprice = pre
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f9(c, pr)
        }
    }

    // Extract current_price
    // Extract new_price
    // If exist
    f10 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "strong" {
            for _, a := range node.Attr {
                if a.Key == "itemprop" {
                    if strings.Contains(a.Val, "price") {
                        pre := renderNode(node)
                        pre = extractContext(pre)
                        // Dbg
                        if ENV_PREF == "dev" {
                            fmt.Println("Found <strong itemprop=price", pre)
                        }
                        // Overwrite
                        pr.Listingprice = pre
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f10(c, pr)
        }
    }

    f11 = func(node *html.Node, pr *Product) {
        if node.Type == html.ElementNode && node.Data == "span" {
            for _, a := range node.Attr {
                if a.Key == "class" {
                    if strings.Contains(a.Val, "star_for_discounted_price") {
                        pre := renderNode(node)
                        pre = extractContext(pre)
                        pre = strings.Replace(pre, "(", "", -1)
                        pre = strings.Replace(pre, ")", "", -1)
                        pre = strings.Replace(pre, "*", "", -1)
                        // Overwrite
                        pr.Discountprice = pre
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f11(c, pr)
        }
    }

    // get target pages from mongo
    coll := makeTimePrefix(LetuCollectionPages)
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

    i = 0
    for _, v := range results {
        var httpClient = &http.Client{
            Timeout: time.Second * 2200,
        }
        url_final := LetuRootUrl + v.Link
        fmt.Println("URL:",url_final)
        pr = &Product{Price: "default", Url: url_final}
        resp, err := httpClient.Get(url_final)
        if err != nil {
            syslog.Critf("Step3 httpClient error: %s", err)
            fmt.Println("Step3 httpClient error", err)
            continue
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
        br := &BrandSingle{Name: v.Brand}

        // Find product image
        // Just before as all the text context
        f6(doc, pr)
        f7(doc, pr)
        f(doc, pr, br)
        i++
    }

    syslog.Syslog(syslog.LOG_INFO, "Letu step3 end")
    fmt.Println("Letu step3 end")
}
