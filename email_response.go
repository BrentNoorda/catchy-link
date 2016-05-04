package catchylink

import (
    "fmt"
    "html"
    "time"
    "strings"
    "strconv"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
    "google.golang.org/appengine/memcache"
    "google.golang.org/appengine/datastore"
)

func email_doit_success(w http.ResponseWriter,linkRequest CatchyLinkRequest) {
    var page string
    page = strings.Replace(email_doit_success_html,"{{shorturl_a}}",strings.Replace(linkRequest.CatchyUrl,"\"","&quot;",1),1)
    page = strings.Replace(page,"{{shorturl_t}}",html.EscapeString(linkRequest.CatchyUrl),1)
    page = strings.Replace(page,"{{longurl_a}}",strings.Replace(linkRequest.LongUrl,"\"","&quot;",1),1)
    page = strings.Replace(page,"{{longurl_t}}",html.EscapeString(linkRequest.LongUrl),1)
    page = strings.Replace(page,"{{duration}}",duration_to_string(linkRequest.Duration),1)
    if linkRequest.Duration == 1 {
        page = strings.Replace(page,"{{reminder}}","style=\"display:none;\"",1)
    }
    fmt.Fprint(w,page)
}

func email_response_handler(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)
    parts := strings.Split(r.URL.Path,"/")
    if len(parts) < 5 {
        log.Errorf(ctx,"email_reponse_handler weird URL \"%s\"",r.URL.Path)
        input_form_with_error_msg(w,"globalerror","Unrecognized URL",nil)
    } else {
        command := parts[2]
        dbid, err := strconv.ParseInt(parts[3],10,64)
        uniqueKey := parts[4]
        if err != nil {
            log.Errorf(ctx,"email_reponse_handler weird URL \"%s\"\nerror: %v",r.URL.Path,err)
            input_form_with_error_msg(w,"globalerror","Unrecognized URL",nil)
        } else {
            if command == "doit"  ||  command == "cancel" {

                // check if this record exists and is still valid
                k := datastore.NewKey(ctx, "linkrequest", "", dbid, nil)
                log.Infof(ctx,"\n\n\nk = %v\n\n\n ",k)

                e := new(CatchyLinkRequest)
                if err := datastore.Get(ctx, k, e); err != nil {
                    log.Errorf(ctx,"email_reponse_handler datastore.Get failed. URL.Path:%s, err:%v",r.URL.Path,err)
                    input_form_with_error_msg(w,"globalerror","That URL request is not in our system. Maybe it has timed out.",nil)
                } else if e.UniqueKey != uniqueKey {
                    log.Errorf(ctx,"email_reponse_handler uniqueKey does not match.")
                    input_form_with_error_msg(w,"globalerror","That URL request is not in our system. Maybe it has timed out.",nil)
                } else if e.Expire <= time.Now().Unix() {
                    log.Errorf(ctx,"email_reponse_handler expire has elapsed.",r.URL.Path)
                    input_form_with_error_msg(w,"globalerror","That URL request is not in our system. Maybe it has timed out.",nil)
                } else {

                    // create the redirect record (unless someone else created it first)
                    var key *datastore.Key
                    var redirect CatchyLinkRedirect
                    var lCatchyUrl string

                    lCatchyUrl = strings.ToLower(e.CatchyUrl)
                    key = datastore.NewKey(ctx,"redirect",lCatchyUrl,0,nil)

                    err = datastore.Get(ctx, key, &redirect)
                    if err == nil  &&  strings.ToLower(redirect.Email) != strings.ToLower(e.Email) {
                        input_form_with_error_msg(w,"globalerror","Looks like someone else already took that catchy.link. That was quick. Sorry.",nil)
                    } else {
                        redirect.LongUrl = e.LongUrl
                        redirect.CatchyUrl = e.CatchyUrl
                        redirect.Email = e.Email
                        redirect.Duration = e.Duration
                        redirect.Expire = time.Now().Unix() + (int64(redirect.Duration) * 60 * 60 * 24)
                        redirect.Warned = 0

                        _, err = datastore.Put(ctx,key,&redirect)

                        // delete the key from memcache (in case it's there) because it is probably no longer valid
                        memcache.Delete(ctx,lCatchyUrl)

                        if err != nil {
                            log.Errorf(ctx,"Error %v putting catchyurl record %v",err,redirect)
                            input_form_with_error_msg(w,"globalerror","Unknown error creating record. Sorry.",nil)
                        } else {
                            email_doit_success(w,*e)
                        }
                    }
                }
            } else {
                log.Errorf(ctx,"email_reponse_handler weird URL \"%s\"",r.URL.Path)
                input_form_with_error_msg(w,"globalerror","Unrecognized URL",nil)
            }
        }
    }
}
