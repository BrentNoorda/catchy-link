package catchylink

import (
    "fmt"
    "time"
    "html"
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

func redirect_prompt(w http.ResponseWriter,url string,email string/*may be null*/) {
    var page string
    if email == "" {
        page = prompt_redirect_html()

    } else {
        page = prompt_redirect_with_email_html()
        page = strings.Replace(page,"{{youremail}}",email,1)
    }

    page = strings.Replace(page,"href=\"{{long_url}}\"","href=\"" + html.EscapeString(url) + "\"",1)
    page = strings.Replace(page,"{{long_url}}",url,1)
    fmt.Fprint(w,page)
}

func embed_to_url(w http.ResponseWriter,url string,title string) {
    var page string = embedded_iframe_html()
    page = strings.Replace(page,"{{url-goes-here}}",url,1)
    page = strings.Replace(page,"{{title-goes-here}}",title,1)
    fmt.Fprint(w,page)
}

func not_found_form(w http.ResponseWriter,catchyUrl string) {
    var page string
    w.WriteHeader(http.StatusNotFound)
    page = strings.Replace(notfound_404_form_html(),"{{catchyurl-value}}",html.EscapeString(catchyUrl),1)
    page = strings.Replace(page,"{{notfound-link}}",myRootUrl+"/"+catchyUrl,1)
    fmt.Fprint(w,page)
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
        //log.Infof(ctx, "\nPath: %s\nRawPath: %s\nRawQuery: %s",r.URL.Path,r.URL.RawPath,r.URL.RawQuery)
        //log.Infof(ctx, "\nRequestURI: %s\n",r.RequestURI)
        //log.Infof(ctx, "\nEscapedPath(): %s\n",r.URL.EscapedPath())

        // strip any slashes or spaces from beginning or end of this raw query string
        var catchyUrl, lCatchyUrl string
        catchyUrl = r.URL.Path
        if r.URL.RawQuery != "" {
            catchyUrl += "?" + r.URL.RawQuery
        }
        catchyUrl = strings.TrimRight(strings.TrimLeft(catchyUrl,"/ \n\r\t"),"/ \n\r\t")
        lCatchyUrl = strings.ToLower(catchyUrl)
        //log.Infof(ctx, "lCatchyUrl = \"%s\"\n",lCatchyUrl)
        //log.Infof(ctx, "^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")

        if lCatchyUrl == "" {
            input_form(w)
        } else {

            // TODO: look into memcache Expiration

            var err error
            var item *memcache.Item
            var splits []string
            var mode int16

            // look first in memcache
            if item, err = memcache.Get(ctx,lCatchyUrl); err == nil {
                //log.Infof(ctx,"url \"%s\" FOUND in memcache",lCatchyUrl);
                mode = get_link_mode(int16(item.Flags))
                var url string = string(item.Value)
                if url == "" {
                    // cached record the this was not found -- it's useful to cache notfound in case many people are looking for it
                    not_found_form(w,catchyUrl)
                } else {
                    if mode == mode_automatic {
                        redirect_to_url(w,r,url)
                    } else if mode == mode_embed {
                        // url and title are separated by a space character
                        splits = strings.SplitN(url," ",2)
                        embed_to_url(w,splits[0],splits[1])
                    } else if mode == mode_prompt {
                        redirect_prompt(w,url,"")
                    } else if mode == mode_prompt_email {
                        // url and email are separated by a space character
                        splits = strings.SplitN(url," ",2)
                        redirect_prompt(w,splits[0],splits[1])
                    } else {
                        log.Errorf(ctx,"OOOPS don't know how to embed yet for %X for url %s",mode,url)
                    }
                }
            } else {

                //log.Infof(ctx,"url \"%s\" not found in memcache; err = %v",lCatchyUrl,err);

                // find this catchyurl in the database
                var key *datastore.Key
                var redirect CatchyLinkRedirect
                key = datastore.NewKey(ctx,"redirect",lCatchyUrl,0,nil)
                //log.Infof(ctx,"key from %s = %v",lCatchyUrl,key)
                if datastore.Get(ctx, key, &redirect) != nil {
                    // there is no existing record, save that in the cache in case others try the same one
                    item = &memcache.Item{
                        Key: lCatchyUrl,
                        Value: []byte(""),
                    }
                    if err = memcache.Set(ctx, item); err != nil {
                        log.Errorf(ctx,"Error setting memcache null for \"%s\", err = %v",lCatchyUrl,err)
                    }

                    not_found_form(w,catchyUrl)
                } else {
                    // don't check here for expiration time, because the periodic cleaner will remove stuff at least once
                    // per day, and if something is returned for up to a day too long then who cares...

                    // store in memcache so we find it more quickly next time
                    var url string = redirect.LongUrl
                    mode = get_link_mode(redirect.OptF)
                    if mode == mode_embed {
                        url += " " + redirect.CatchyUrl
                    } else if mode == mode_prompt_email {
                        url += " " + redirect.Email
                    }
                    item = &memcache.Item{
                        Key: lCatchyUrl,
                        Value: []byte(url),
                        Expiration: time.Unix(redirect.Expire,0).Sub(time.Now()),
                        Flags: uint32(redirect.OptF),
                    }
                    if err = memcache.Set(ctx, item); err != nil {
                        log.Errorf(ctx,"Error setting memcache for \"%s\", err = %v",lCatchyUrl,err)
                    }

                    var mode int16 = get_link_mode(redirect.OptF)
                    if mode == mode_automatic {
                        redirect_to_url(w,r,redirect.LongUrl)
                    } else if mode == mode_embed {
                        embed_to_url(w,redirect.LongUrl,redirect.CatchyUrl)
                    } else if mode == mode_prompt {
                        redirect_prompt(w,redirect.LongUrl,"")
                    } else if mode == mode_prompt_email {
                        redirect_prompt(w,redirect.LongUrl,redirect.Email)
                    } else {
                        log.Errorf(ctx,"OOOPS don't know how to embed yet for %X for url %s",mode,url)
                    }
                }
            }

        }
    }
}
