package catchylink

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
)

func init() {
    http.HandleFunc("/", handler)
}

func homepage(w http.ResponseWriter, r *http.Request) {
    text, err := ioutil.ReadFile("web/index.html")
    if err != nil {
        ctx := appengine.NewContext(r)
        log.Errorf(ctx,"Could not read homepage")
    }
    fmt.Fprint(w,string(text))
}

func handler(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)
    log.Infof(ctx,"Catchylink3, world!<br/>Path:\"" + r.URL.Path + "\"  RawPath:\"" + r.URL.RawPath + "\"  RawQuery:\"" + r.URL.RawQuery + "\"")
    if r.URL.Path == "/" {
        homepage(w,r)
    } else {
        fmt.Fprint(w, "Catchylink3, world!<br/>Path:" + r.URL.Path + "<br/>RawPath:" + r.URL.RawPath + "<br/>RawQuery:" + r.URL.RawQuery)
    }
}