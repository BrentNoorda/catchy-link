package catchylink

import (
    "strings"
    "strconv"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
    "google.golang.org/appengine/datastore"
)

func email_response_handler(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)
    parts := strings.Split(r.URL.Path,"/")
    if len(parts) < 5 {
        log.Errorf(ctx,"email_reponse_handler weird URL \"%s\"",r.URL.Path)
        homepage_with_error_msg(w,"globalerror","Unrecognized URL",nil)
    } else {
        command := parts[2]
        dbid, err := strconv.ParseInt(parts[3],10,64)
        uniqueKey := parts[4]
        if err != nil {
            log.Errorf(ctx,"email_reponse_handler weird URL \"%s\"\nerror: %v",r.URL.Path,err)
            homepage_with_error_msg(w,"globalerror","Unrecognized URL",nil)
        } else {
            if command == "doit"  ||  command == "cancel" {

                // check if this record exists and is still valid
                k := datastore.NewKey(ctx, "linkrequest", "", dbid, nil)
                log.Infof(ctx,"\n\n\nk = %v\n\n\n ",k)

                e := new(CatchyLinkRequest)
                if err := datastore.Get(ctx, k, e); err != nil {
                    log.Errorf(ctx,"email_reponse_handler datastore.Get failed. URL.Path:%s, err:%v",r.URL.Path,err)
                    homepage_with_error_msg(w,"globalerror","That URL request is not in our system. Maybe it has timed out.",nil)
                } else if e.UniqueKey != uniqueKey {
                    log.Errorf(ctx,"email_reponse_handler datastore.Get failed. URL.Path:%s, unique key did not match.",r.URL.Path)
                    homepage_with_error_msg(w,"globalerror","That URL request is not in our system. Maybe it has timed out.",nil)
                } else {
                    homepage(w)
                }
            } else {
                log.Errorf(ctx,"email_reponse_handler weird URL \"%s\"",r.URL.Path)
                homepage_with_error_msg(w,"globalerror","Unrecognized URL",nil)
            }
        }
    }
}
