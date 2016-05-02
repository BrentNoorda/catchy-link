package catchylink

import (
    "time"
    "strings"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
    "google.golang.org/appengine/memcache"
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


func redirect_handler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/" {
        if r.Method == "POST" {
            post_new_catchy_link(w,r)
        } else {
            input_form(w)
        }
    } else {

        ctx := appengine.NewContext(r)

        //log.Infof(ctx, "vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
        //log.Infof(ctx, "\nPath: %s\nRawPath: %s\nRawQuery: %s",r.URL.RawQuery,r.URL.Path,r.URL.RawPath,r.URL.RawQuery)
        //log.Infof(ctx, "\nRequestURI: %s\n",r.RequestURI)
        //log.Infof(ctx, "^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")

        // strip any slashes or spaces from beginning or end of this raw query string
        var lCatchyUrl string
        lCatchyUrl = strings.ToLower(strings.TrimRight(strings.TrimLeft(r.RequestURI,"/ \n\r\t"),"/ \n\r\t"))

        if lCatchyUrl == "" {
            input_form(w)
        } else {

            // TODO: look into memcahce Expiration

            var err error
            var item *memcache.Item

            // look first in memcache
            if item, err = memcache.Get(ctx,lCatchyUrl); err == nil {
                //log.Infof(ctx,"url \"%s\" FOUND in memcache",lCatchyUrl);
                redirect_to_url(w,r,string(item.Value))

            } else {

                //log.Infof(ctx,"url \"%s\" not found in memcache; err = %v",lCatchyUrl,err);

                // find this catchyurl in the database
                var key *datastore.Key
                var redirect CatchyLinkRedirect
                key = datastore.NewKey(ctx,"redirect",lCatchyUrl,0,nil)
                //log.Infof(ctx,"key from %s = %v",lCatchyUrl,key)
                if datastore.Get(ctx, key, &redirect) != nil {
                    // there is no existing record
                    input_form_with_error_msg(w,"globalerror","Unrecognized catchy.link URL",nil)
                } else {
                    // don't check here for expiration time, because the periodic cleaner will remove stuff at least once
                    // per day, and if something is returned for up to a day too long then who cares...

                    // store in memcache so we find it more quickly next time
                    item = &memcache.Item{
                        Key: lCatchyUrl,
                        Value: []byte(redirect.LongUrl),
                        Expiration: time.Unix(redirect.Expire,0).Sub(time.Now()),
                    }
                    if err = memcache.Set(ctx, item); err != nil {
                        log.Errorf(ctx,"Error setting memcache for \"%s\", err = %v",lCatchyUrl,err)
                    }

                    redirect_to_url(w,r,redirect.LongUrl)
                }
            }

        }
    }
}
