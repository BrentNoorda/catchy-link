package catchylink

import (
    "strings"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
    "google.golang.org/appengine/datastore"
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
        log.Infof(ctx, "vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
        log.Infof(ctx, "\nPath: %s\nRawPath: %s\nRawQuery: %s",r.URL.RawQuery,r.URL.Path,r.URL.RawPath,r.URL.RawQuery)
        log.Infof(ctx, "\nRequestURI: %s\n",r.RequestURI)
        log.Infof(ctx, "^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")

        // strip any slashes or spaces from beginning or end of this raw query string
        var lCatchyUrl string
        lCatchyUrl = strings.ToLower(strings.TrimRight(strings.TrimLeft(r.RequestURI,"/ \n\r\t"),"/ \n\r\t"))

        if lCatchyUrl == "" {
            input_form(w)
        } else {

            // find this catchyurl in the database
            var key *datastore.Key
            var redirect CatchyLinkRedirect
            key = datastore.NewKey(ctx,"redirect",lCatchyUrl,0,nil)
            log.Infof(ctx,"key from %s = %v",lCatchyUrl,key)
            if datastore.Get(ctx, key, &redirect) != nil {
                // there is no existing record
                input_form_with_error_msg(w,"globalerror","Unrecognized catchy.link URL",nil)
            } else {
                redirect_to_url(w,r,redirect.LongUrl)
            }
        }
    }
}
