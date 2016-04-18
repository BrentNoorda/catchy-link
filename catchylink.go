package catchylink

import (
    "os"
    "fmt"
    "strings"
    "io/ioutil"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
)

var index_html string

func init() {

    // read index.html only once, so we don't read it again and again and again
    bytes, err := ioutil.ReadFile("web/index.html")
    if err != nil {
        fmt.Fprintf(os.Stderr,"YIKES!!!! Cannot read web/index.html");
    } else {
        index_html = string(bytes)
    }

    http.HandleFunc("/", handler)
}

func post_new_catchy_link(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)
    r.ParseForm()
    log.Infof(ctx,"POST FORM longurl=\"%s\"",r.PostFormValue("longurl"))
    log.Infof(ctx,"POST FORM catchyurl=\"%s\"",strings.Join(r.Form["catchyurl"],""))
    log.Infof(ctx,"POST FORM youremail=\"%s\"",r.Form["youremail"][0])
    for k, v := range r.Form {
        log.Infof(ctx,"key:%s", k)
        log.Infof(ctx,"val:%s", strings.Join(v, ""))
    }
    homepage(w,r)
}

func homepage(w http.ResponseWriter, r *http.Request) {
    if len(index_html) <= 0 {
        ctx := appengine.NewContext(r)
        log.Errorf(ctx,"Could not read homepage")
    }
    fmt.Fprint(w,index_html)
}

func handler(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)
    log.Infof(ctx,"%s","Catchylink3, world!<br/>Path:\"" + r.URL.Path + "\"  RawPath:\"" + r.URL.RawPath + "\"  RawQuery:\"" + r.URL.RawQuery + "\"")
    if r.URL.Path == "/" {
        if r.Method == "POST" {
            post_new_catchy_link(w,r)
        } else {
            homepage(w,r)
        }
    } else {
        fmt.Fprint(w, "Catchylink3, world!<br/>Path:" + r.URL.Path + "<br/>RawPath:" + r.URL.RawPath + "<br/>RawQuery:" + r.URL.RawQuery)
    }
}