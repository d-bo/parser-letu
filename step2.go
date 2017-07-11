package main

import (
    "io"
    "log"
    "fmt"
    "time"
    "bytes"
    "strings"
    "net/http"
    //"io/ioutil"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "golang.org/x/net/html"
    )

const LetuRootUrl string = "https://www.letu.ru/"
const LetuBrandUrl string = "https://www.letu.ru/browse/brandsDisplay.jsp"
const LetuDB = "import17"
const LetuCollection = "letu_brands"
const LetuCollectionPages = "letu_pages"

var glob_session, glob_err = mgo.Dial("mongodb://localhost:27017/")

// http response body struct
type Page struct {
    Body []byte
}

// Letu brand page: name & link
type Brand struct {
    Name string
    Link string
}

// single link product page
type Link struct {
    Link string
}

// Link pool
var LinkPool []Link

// Brand pool
var BrandPool []Brand
var BrandPoolResult []Brand

// Get url http response
func loadPage(url string) {
    var httpClient = &http.Client{
        Timeout: time.Second * 100,
    }
    resp, err := httpClient.Get(url)
    if err != nil {
        fmt.Println(err)
    }
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

// Insert document to mongo brands collection
func mongoInsertBrand(b *Brand) bool {
    c := glob_session.DB(LetuDB).C(LetuCollection)
    glob_session.SetMode(mgo.Monotonic, true)
    err := c.Insert(b)
    if err != nil {
        return true
    } else {
        return false
    }
}

// Step 2: Iterate over letoile brand pages
func main() {
    defer glob_session.Close()

    // html parser itself
    var f func(*html.Node)
    f = func(node *html.Node) {
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
                b := Link{value}
                LinkPool = append(
                    LinkPool,
                    b,
                )

                c := glob_session.DB(LetuDB).C(LetuCollectionPages)
                glob_session.SetMode(mgo.Monotonic, true)
                err := c.Insert(b)

                if err != nil {
                    fmt.Println(err)
                }
            }
        }

        // iterate inner nodes recursive
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }

    var results []Brand

    // get target pages from mongo
    c := glob_session.DB(LetuDB).C(LetuCollection)
    glob_session.SetMode(mgo.Monotonic, true)
    err := c.Find(bson.M{}).All(&results)

    if glob_err != nil {
        log.Fatal(err)
    }

    fmt.Println(len(results))

    for _, v := range results {
        url_final := LetuRootUrl+ v.Link + "&Nrpp=6000"
        //body := loadPage(url_final)

        var httpClient = &http.Client{
            Timeout: time.Second * 10,
        }
        resp, err := httpClient.Get(url)
        if err != nil {
            fmt.Println(err)
        }
        if body {
            doc, err := html.Parse(strings.NewReader(string(body.Body)))
            if err != nil {
                log.Fatal(err)
            }
            fmt.Println(url_final,doc, "\n")
            f(doc)
        }
    }

    if glob_err != nil {
        log.Fatal(glob_err)
    }

    fmt.Println(BrandPoolResult)
}
