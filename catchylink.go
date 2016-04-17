package catchylink

import (
    "fmt"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
)

func init() {
    http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)
    log.Infof(ctx,"Catchylink3, world!<br/>Path:\"" + r.URL.Path + "\"  RawPath:\"" + r.URL.RawPath + "\"  RawQuery:\"" + r.URL.RawQuery + "\"")
    if r.URL.Path == "/" {
        fmt.Fprintf(w, "Return index.html")
    } else {
        fmt.Fprint(w, "Catchylink3, world!<br/>Path:" + r.URL.Path + "<br/>RawPath:" + r.URL.RawPath + "<br/>RawQuery:" + r.URL.RawQuery)
    }
}