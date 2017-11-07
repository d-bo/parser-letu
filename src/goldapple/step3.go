package goldapple

/**
 * Step3: collect product data
 */

import (
    "os"
    "io"
    "log"
    "fmt"
    "time"
    "strings"
    "net/http"
    "io/ioutil"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "golang.org/x/net/html"
    )

const LetuProducts = "letu_products_final"
const LetuPrice = "letu_price"
const GestoriDB = "gestori"
const LogFile = "Log"
const LogCollection = "Log"

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
    Volume string
    Url string
}

// struct for ILDE_price
type ProductPrice struct {
    Price string
    Price_discount string
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
        fmt.Println("step3: ", err)
    }
    bytemsg := []byte(msg)
    n, err := f.Write(bytemsg)
    if err == nil && n < len(bytemsg) {
        fmt.Println("step3: ", io.ErrShortWrite)
    }
}

func makeTimeMonthlyPrefix(coll string) string {
    t := time.Now()
    ti := t.Format("01-2006")
    fin := ti + "_" + coll
    return fin
}

func Step3(glob_session *mgo.Session) {

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
                        fmt.Println("step3: Articul: ", pre)
                        
                        // gestori match
                        var gestres Gestori
                        c := glob_session.DB(LetuDB).C(LetuProducts)
                        glob_session.SetMode(mgo.Monotonic, true)
                        err := c.Find(bson.M{"Artic": pre}).One(&gestres)
                        if err != nil {
                            fmt.Println("step3: GESTORI NOT FOUND")
                        } else {
                            fmt.Println("step3: GESTORI MATCH: ", gestres)
                            pr.Gestori = gestres.Cod_good
                            logstring := []byte("step3: Gestori match: "+pre+pr.Gestori+"\n")
                            Log(logstring)
                        }
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
                        pr.Volume = pre
                    }
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f5(c, pr)
        }
    }
    // found product container
    f1 = func(node *html.Node, pr *Product, br *BrandSingle) {
        if node.Type == html.ElementNode && node.Data == "tr" {
            f2(node, pr)	// price
            f3(node, pr)	// article
            f4(node, pr)	// name
            f5(node, pr)	// desc

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
                    fmt.Println("step3: ", err)
                }

                price := ProductPrice{
                    Price: pr.Price,
                    Price_discount: pr.Price_discount,
                    Articul: pr.Articul,
                    Brand: br.Name,
                }

                if num < 1 {

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
                    // insert 'letu_products_final'
                    err := c.Insert(new)
                    if err != nil {
                        fmt.Println("step3: ", err)
                    }
                    // insert 'letu_price'
                    err = d.Insert(price)
                    if err != nil {
                        fmt.Println("step3: ", err)
                    }
                    // log new product
                    e.Insert(LogStruct{
                        Subject: "letu",
                        Action: "new_articul",
                        Val: pr.Articul,
                        Date: makeTimePrefix(""),
                    })
                } else {
                    // if DOUBLE
                    // update price column
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
                        fmt.Println("step3: ", err)
                    }
                    fmt.Println("step3: Double articul")
                }
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f1(c, pr, br)
        }
    }
    // extract image
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
                        fmt.Println("MATCH")
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
                fmt.Println("IMAGE: ", src)
            }
            match = false
        }
        i = 0
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f6(c, pr)
        }
    }
    // extract image
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
                        fmt.Println("MATCH DESC", pr.Desc)
                    }
                }
            }
        }

        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f7(c, pr)
        }
    }

/*
                        pre := renderNode(node)
                        pre = extractContext(pre)
                        pre = strings.TrimSpace(pre)
                        pr.Desc = pre
                        fmt.Println(pr.Desc)
*/

    start := time.Now()

    // get target pages from mongo
    coll := makeTimePrefix(LetuCollectionPages)
    if LetuDB == "" {
        LetuDB = "parser"
    }
    c := glob_session.DB(LetuDB).C(coll)
    glob_session.SetMode(mgo.Monotonic, true)
    err := c.Find(bson.M{}).All(&results)

    if err != nil {
        fmt.Println(err)
    }

    i = 0
    for _, v := range results {
        var httpClient = &http.Client{
            Timeout: time.Second * 1200,
        }
        url_final := LetuRootUrl + v.Link
        pr = &Product{Price: "default", Url: url_final}
        resp, err := httpClient.Get(url_final)
        if err != nil {
            fmt.Println("step3: ", err)
            continue
        }
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Println("step3: ", err)
        }
        doc, err_p := html.Parse(strings.NewReader(string(body)))
        if err_p != nil {
            log.Println(err)
        }
        fmt.Println("step3: ", url_final, "\r\n")
        br := &BrandSingle{Name: v.Brand}
        // find product image
        // just before as all the text context
        f6(doc, pr)
        f7(doc, pr)
        f(doc, pr, br)
        fmt.Println(pr)
        i++
    }

    elapsed := time.Since(start)
    fmt.Printf("step3: Script took %s", elapsed)
}
