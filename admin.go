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
    var query *datastore.Query
    var err error
    var keys []*datastore.Key
    var key *datastore.Key

    ctx := appengine.NewContext(r)
    log.Infof(ctx,"%s","!!!!admin_handler<br/>Path:\"" + r.URL.Path + "\"  RawPath:\"" + r.URL.RawPath + "\"  RawQuery:\"" + r.URL.RawQuery + "\"")

    // remove link requests that have timed out
    if r.URL.Path == "/-/cleanup_old_link_requests" {
        query = datastore.NewQuery("linkrequest").Filter("Expire <",time.Now().Unix()-30).KeysOnly() // 30 second back so don't delete here while checking there
        keys, err = query.GetAll(ctx, nil)
        if err != nil {
            log.Errorf(ctx, "query error: %v", err)
        } else {
            err := datastore.DeleteMulti(ctx,keys)
            if err != nil {
                log.Errorf(ctx, "DeleteMulti error: %v, keys = %v", err,keys)
            }
        }
    }

    // send emails to everyone who is going to expire in expiration_warning_days or less, and has not yet
    // received a warning email. Also extend their expiration time so they have at least expiration_warning_days
    // after receiving the email
    var now time.Time
    var expiring_soon_cutoff int64

    now = time.Now()
    expiring_soon_cutoff = now.Unix() + ((expiration_warning_days+666/*TODO remove 666*/) * (60*60*24))
    query = datastore.NewQuery("redirect").Filter("Expire <",expiring_soon_cutoff).Filter("Warned =",false)
    for q_iter := query.Run(ctx); ; {
        var redirect CatchyLinkRedirect
        key, err = q_iter.Next(&redirect)
        if ( err == datastore.Done ) {
            break
        } else if err != nil {
            log.Errorf(ctx,"Error querying redirect for timeouts = %v",err)
            break
        } else {

            // send an email for this record, giving the user a chance to renew, and then change the record to know
            // that such an email has already been sent (and set expire up a tad)


            log.Infof(ctx,"key = %v\nredirect = %v",key,redirect)
        }
    }
}

func robots_txt_handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "user-agent: *\r\nAllow: /$\r\nDisallow: /\r\n")
}

func favicon_ico_handler(w http.ResponseWriter, r *http.Request) {
    redirect_to_url(w,r,"https://googledrive.com/host/0B4rxOB63nnDMdE10cnlDWGxDSUU")
}