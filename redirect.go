package catchylink

import (
    "fmt"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
)

func redirect_to_url(w http.ResponseWriter, r *http.Request,url string) {
    var redirect_code int

    if r.ProtoAtLeast(1,1) {
        redirect_code = http.StatusTemporaryRedirect
    } else {
        redirect_code = http.StatusFound
    }
    http.Redirect(w,r,url,redirect_code)
}


func redirect_AxonActionPotential(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)
    log.Infof(ctx,"----------------- redirect_AxonActionPotential -----------------")
    redirect_to_url(w,r,"http://www.brent-noorda.com/medical/physiologylab/axon/index.html");
}

func redirect_handler(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)
    log.Infof(ctx,"%s","Catchylink3, world!Path:\"" + r.URL.Path + "\"  RawPath:\"" + r.URL.RawPath + "\"  RawQuery:\"" + r.URL.RawQuery + "\"")
    if r.URL.Path == "/" {
        if r.Method == "POST" {
            post_new_catchy_link(w,r)
        } else {
            input_form(w)
        }
    } else {
        fmt.Fprint(w, "Catchylink3, world!<br/>Path:" + r.URL.Path + "<br/>RawPath:" + r.URL.RawPath + "<br/>RawQuery:" + r.URL.RawQuery)
    }
}
