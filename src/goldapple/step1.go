package goldapple

/**
 * Step1: extract brand urls
 */

import (
    "os"
    "io"
    "fmt"
    "log"
    "time"
    "bytes"
    "strings"
    "net/http"
    "io/ioutil"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "golang.org/x/net/html"
    "github.com/blackjack/syslog"
    )

const LetuRootUrl string = "https://www.letu.ru/"
const LetuBrandUrl string = "https://www.letu.ru/browse/brandsDisplay.jsp"
const LetuCollection = "letu_brands"
const AllBrands = "all_brands"

var ENV_PREF string

var LetuDB string = os.Getenv("LETU_MONGO_DB")

// http response body struct
type Page struct {
    Body []byte
}

// Letu brand page: name & link
type Brand struct {
    Name string
    Link string
}

// all_brands collection
type AllBrand struct {
    Val string
    Source string
}

// Brand pool
var BrandPool []Brand

// Get url http response
func loadPage(url string) (*Page) {
    var httpClient = &http.Client{
        Timeout: time.Second * 2200,
    }
    resp, err := httpClient.Get(url)
    if err != nil {
        syslog.Critf("Step1 load page error: %s", err)
        fmt.Println("Step1 load page error", err)
        return &Page{Body: []byte{0}}
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
                syslog.Critf("Step1 extractContext() error: %s", z.Err())
                fmt.Println("Step1 extractContext() error", z.Err())
				return "Step1 extractContext() error"
			case html.TextToken:
				text := string(z.Text())
				return text
		}
	}
}

// Insert document to mongo brands collection
func mongoInsertBrand(b *Brand, glob_session *mgo.Session) bool {
    coll := MakeTimePrefix(LetuCollection)
    coll_all := AllBrands
    if LetuDB == "" {
        LetuDB = "parser"
    }
    c := glob_session.DB(LetuDB).C(coll)
    c_all := glob_session.DB(LetuDB).C(coll_all)
    glob_session.SetMode(mgo.Monotonic, true)

    // GLOBAL BRANDS DOUBLE
    // check `all_brands`` double
    // not so necessary but quite space economy
    allb := AllBrand{b.Name, "letu"}
    num, err := c_all.Find(bson.M{"val": b.Name, "source": "letu"}).Count()
    if num < 1 {
        c_all.Insert(allb)
    } else {
        // double
    }
    if err != nil {
        syslog.Critf("Step1 insert brand error: %s", err)
        fmt.Println("Step1 insert brand error", err)
    }

    // TODAY BRANDS DOUBLE
    // check today brands double
    num, err = c.Find(bson.M{"name": b.Name}).Count()
    if num < 1 {
        err = c.Insert(b)
    } else {
        // really double
    }
    if err != nil {
        syslog.Critf("Step1 insert error: %s", err)
        fmt.Println("Step1 insert error", err)
    }

    if err != nil {
        return true
    } else {
        return false
    }
}

func Step1(glob_session *mgo.Session) {

    syslog.Syslog(syslog.LOG_INFO, "Letu step1 start")
    fmt.Println("Letu step1 start")

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
                mongoInsertBrand(&b, glob_session)
            }
        }

        // iterate inner nodes recursive
        for c := node.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)
    syslog.Syslog(syslog.LOG_INFO, "Letu step1 end")
    fmt.Println("Letu step1 end")
}