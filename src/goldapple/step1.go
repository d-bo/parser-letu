package goldapple

/**
 * Step1: extract brand urls
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
    "golang.org/x/net/html"
    )

const LetuRootUrl string = "https://www.letu.ru/"
const LetuBrandUrl string = "https://www.letu.ru/browse/brandsDisplay.jsp"
const LetuCollection = "letu_brands"

var LetuDB string = os.Getenv("LETU_MONGO_DB")
var glob_session_step1, glob_err_step1 = mgo.Dial("mongodb://localhost:27017/")

// http response body struct
type Page struct {
    Body []byte
}

// Letu brand page: name & link
type Brand struct {
    Name string
    Link string
}

// Brand pool
var BrandPool []Brand

// Get url http response
func loadPage(url string) (*Page) {
    var httpClient = &http.Client{
        Timeout: time.Second * 10,
    }
    resp, err := httpClient.Get(url)
    if err != nil && resp.StatusCode == 200 {
        panic(err)
    }
    body, err := ioutil.ReadAll(resp.Body)
    return &Page{Body: body}
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

func makeTimePrefix(coll string) string {
    t := time.Now()
    ti := t.Format("02-01-2006")
    if coll == "" {
        return ti
    }
    fin := ti + "_" + coll
    return fin
}

// Insert document to mongo brands collection
func mongoInsertBrand(b *Brand) bool {
    coll := makeTimePrefix(LetuCollection)
    if LetuDB == "" {
        LetuDB = "parser"
    }
    c := glob_session_step1.DB(LetuDB).C(coll)
    glob_session_step1.SetMode(mgo.Monotonic, true)
    err := c.Insert(b)
    if err != nil {
        return true
    } else {
        return false
    }
}

func Step1() {
    defer glob_session_step1.Close()
    body := loadPage(LetuBrandUrl)
    doc, err := html.Parse(strings.NewReader(string(body.Body)))

    if err != nil {
        log.Fatal(err)
    }

    // html parser itself
    var f func(*html.Node)
    f = func(node *html.Node) {
        match := false
        var value string
        if node.Type == html.ElementNode && node.Data == "option" {
            for _, a := range node.Attr {
                if a.Key == "value" {
                    value = a.Val
                }
                if a.Key == "class" {
                    if strings.Contains(a.Val, "chosen-brand") {
                        match = true
                    }
                }
            }
            if match {
                pre := renderNode(node)
                pre = extractContext(pre)
                b := Brand{pre, value}
                BrandPool = append(
                    BrandPool,
                    b,
                )
                mongoInsertBrand(&b)
            }
        }

        // iterate inner nodes recursive
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)
    fmt.Println(BrandPool)
}
