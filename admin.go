package catchylink

import (
    "fmt"
    "time"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
    "google.golang.org/appengine/datastore"
)

func admin_handler(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)
    log.Infof(ctx,"%s","!!!!admin_handler<br/>Path:\"" + r.URL.Path + "\"  RawPath:\"" + r.URL.RawPath + "\"  RawQuery:\"" + r.URL.RawQuery + "\"")

    if r.URL.Path == "/-/cleanup_old_link_requests" {
        query := datastore.NewQuery("linkrequest").Filter("Expire <",time.Now().Unix()-30).KeysOnly() // 30 second back so don't delete here while checking there
        keys, err := query.GetAll(ctx, nil)
        if err != nil {
            log.Errorf(ctx, "query error: %v", err)
        } else {
            err := datastore.DeleteMulti(ctx,keys)
            if err != nil {
                log.Errorf(ctx, "DeleteMulti error: %v, keys = %v", err,keys)
            }
        }
    }
}

func robots_txt_handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "user-agent: *\r\nAllow: /$\r\nDisallow: /\r\n")
}

func favicon_ico_handler(w http.ResponseWriter, r *http.Request) {
    redirect_to_url(w,r,"https://googledrive.com/host/0B4rxOB63nnDMdE10cnlDWGxDSUU")
}