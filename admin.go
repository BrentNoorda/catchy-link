package catchylink

import (
    "fmt"
    "time"
    "strings"
    "strconv"
    "net/http"
    "golang.org/x/net/context"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
    "google.golang.org/appengine/mail"
    "google.golang.org/appengine/datastore"
    "google.golang.org/appengine/urlfetch"
    "github.com/mailgun/mailgun-go"     // of this is missing do# goapp get github.com/mailgun/mailgun-go
)

func prepare_renew_email_body(linkRequest CatchyLinkRequest,renewUrl string) (body,htmlBody string) {

    var noUrlLink string

    body = "You have a memorable URL on catchy.link that will be expiring soon. The link:\n\n" +
           "   " + strings.Replace(myRootUrl,"//","// ",1) + "/" + linkRequest.CatchyUrl + "\n\n" +
           "redirects to:\n\n" +
           "   " + linkRequest.LongUrl + "\n\n\n" +
           "To RENEW this catchy.link, click on the following link (or copy and paste it to the address field in your browser):\n\n" +
           "   RENEW: " + renewUrl + "\n"

    // make url disguised so email reader doens't automatically make it a link
    noUrlLink = myRootUrl + "/" + linkRequest.CatchyUrl
    noUrlLink = strings.Replace(noUrlLink,"/","<font>/</font>",-1)
    noUrlLink = strings.Replace(noUrlLink,".","<font>.</font>",-1)
    noUrlLink = strings.Replace(noUrlLink,":","<font>:</font>",-1)

    htmlBody = "<table width=\"97%\" style=\"margin: auto;max-width:800px\" align=\"center\">\n" +
               "<tr><td width=\"100%\">\n" +
               "You have a memorable URL on catchy.link that will be expiring soon. The link:<br/><br/>\n" +
               " &nbsp; " + noUrlLink + "<br/><br/>\n" +
               "redirects to:<br/><br/>\n" +
               " &nbsp; <a href=\"" + linkRequest.LongUrl + "\">" + linkRequest.LongUrl + "</a><br/><br/>\n" +
               "To RENEW this catchy.link, click on the following button:<br/><br/>\n" +
               " &nbsp; <a href=\"" + renewUrl + "\"><button style=\"background-color:#dddddd;\"><font size=\"+1\">renew catchy.link</font></button></a><br/><br/><br/>\n" +
               "<font size=\"-2\">if that button fails, copy and paste this url into your browser: " + renewUrl + "</font>" +
               "</td></tr></table>"

     return
}

func admin_handler(w http.ResponseWriter, r *http.Request) {
    var query *datastore.Query
    var err error

    ctx := appengine.NewContext(r)

    if r.URL.Path != "/-/cleanup_old_db_stuff" {
        log.Errorf(ctx, "admin_handler r.URL.Path=\"%s\" != \"%s\"",r.URL.Path,"/-/cleanup_old_db_stuff")
        return
    }

    log.Infof(ctx,"vvvvvvvvvvvvvvvvvvvvvvvvvvv ADMIN_HANDLER vvvvvvvvvvvvvvvvvvvvvvvvvvv")
    log.Infof(ctx,"%s","!!!!admin_handler<br/>Path:\"" + r.URL.Path + "\"  RawPath:\"" + r.URL.RawPath + "\"  RawQuery:\"" + r.URL.RawQuery + "\"")

    /***************************************************************************************************************/
    // remove link requests that have timed out
    var expired_cutoff int64 = time.Now().Unix() - 30    // 30 seconds back so don't delete here while checking there
    query = datastore.NewQuery("linkrequest").Order("Expire")
    request_expire_count := 0
    for q_iter := query.Run(ctx); ; {
        var request CatchyLinkRequest
        var request_key *datastore.Key
        request_key, err = q_iter.Next(&request)
        if err == datastore.Done {
            break
        } else if err != nil {
            log.Errorf(ctx,"admin_handler request q_iter.Next() err = %v",err)
            break
        } else if expired_cutoff <= request.Expire {
            // have reached the last of the expired requests
            break
        } else {
            request_expire_count += 1
            if err = datastore.Delete(ctx,request_key); err != nil {
                log.Errorf(ctx,"admin_handler request Delete request=%v err=%v",request,err)
            }
        }
    }
    log.Infof(ctx,"%d linkrequest records expired and deleted",request_expire_count)

    /***************************************************************************************************************/
    // delete all redirect records that have expired -
    // for those that are about to expire, send emails to everyone who is going to expire in expiration_warning_days or less,
    // and has not yet received a warning email. Also extend their expiration time so they have at least expiration_warning_days
    // after receiving the email
    var now time.Time = time.Now()
    expired_cutoff = now.Unix() - 30 // 30 seconds back so don't delete here while checking there
    var expiring_soon_cutoff int64 = now.Unix() + (expiration_warning_days * seconds_per_day)

    //log.Infof(ctx,"expiring_soon_cutoff = %d",expiring_soon_cutoff)

    query = datastore.NewQuery("redirect").Order("Expire")
    expiration_warning_count := 0
    expiration_warning_retry_count := 0
    redirect_expire_count := 0
    for q_iter := query.Run(ctx); ; {
        var redirect CatchyLinkRedirect
        var redirect_key *datastore.Key
        redirect_key, err = q_iter.Next(&redirect)
        //log.Infof(ctx,"redirect_key = %v, err = %v",redirect_key,err)
        if err == datastore.Done {
            break
        } else if err != nil {
            log.Errorf(ctx,"admin_handler redirect q_iter.Next() err = %v",err)
            break
        } else if expiring_soon_cutoff <= redirect.Expire {
            // have reached the last of the records we want to examine
            break
        } else {

            //log.Infof(ctx,"NEXT redirect = %v",redirect)

            if (redirect.Expire < expired_cutoff) && ((redirect.Duration <=1) || (redirect.Warn >= max_email_warning_retries)) {
                // all hope is lost for this record - delete it (don't worry much about memcache, it will reach same conclusion soon
                redirect_expire_count += 1
                if err = datastore.Delete(ctx,redirect_key); err != nil {
                    log.Errorf(ctx,"admin_handler redirect Delete redirect=%v err=%v",redirect,err)
                }
            } else {
                // this record hasn't expired yet, but maybe it will soon, so might want to send user a chance to renew
                if (redirect.Duration > 1) && (redirect.Warn < max_email_warning_retries) { // ignore 1-day only or if too many already sent

                    expiration_warning_count += 1
                    var request_key *datastore.Key
                    var retry_tomorrow bool = false

                    // send an email for this record, giving the user a chance to renew, and then change the record to know
                    // that such an email has already been sent (and set expire up a tad)

                    // create CatchyLinkRequest for user to update
                    expire := now.Add( time.Duration(max_email_warning_retries*seconds_per_day*(1000*1000*1000)) )
                    linkRequest := CatchyLinkRequest {
                        UniqueKey: random_string(55),
                        LongUrl: redirect.LongUrl,
                        CatchyUrl: redirect.CatchyUrl,
                        Email: redirect.Email,
                        Expire: expire.Unix(),
                        Duration: redirect.Duration,
                        OptF: redirect.OptF,
                    }
                    request_key, err = datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "linkrequest", nil), &linkRequest)
                    if err != nil {
                        log.Errorf(ctx,"Error putting new CatchyLinkRequest in cron = %v",err)
                        retry_tomorrow = true
                    } else {

                        // send an email that expiration is happening soon
                        renewUrl := fmt.Sprintf("%s/~/renew/%d/%s",myRootUrl,request_key.IntID(),linkRequest.UniqueKey)
                        //cancelUrl := fmt.Sprintf("%s/~/cancel/%d/%s",myRootUrl,reqeust_key.IntID(),linkRequest.UniqueKey)
                        body,htmlBody := prepare_renew_email_body(linkRequest,renewUrl)
                        subject := "Renew URL on Catchy.Link"

                        //log.Infof(ctx,"-------------------------------------------------------------")
                        //log.Infof(ctx,"To: %s",redirect.Email)
                        //log.Infof(ctx,"Subject: %s\n",subject)
                        //log.Infof(ctx,"%s",body)
                        //log.Infof(ctx,"-------------------------------------------------------------")

                        if Mailgun != nil {
                            // send email message through mailgun
                            httpc := urlfetch.Client(ctx)

                            mg := mailgun.NewMailgun(
                                Mailgun.domain_name,
                                Mailgun.secret_key,
                                Mailgun.public_key,
                            )
                            mg.SetClient(httpc)

                            message := mg.NewMessage(
                                 /* From */ Mailgun.from,
                                 /* Subject */ subject,
                                 /* Body */ body,
                                 /* To */ redirect.Email,
                            )
                            message.SetHtml(htmlBody)

                            _, _, err = mg.Send(message)
                            if err != nil {
                                log.Errorf(ctx, "Could not send renew email from Mailgun: %v", err)
                                retry_tomorrow = true
                            }

                        } else {
                            // send email message through app engine
                            msg := &mail.Message{
                                Sender:  sender_email_address_if_no_mailgun,
                                To:      []string{redirect.Email},
                                Subject: subject,
                                Body:    body,
                                HTMLBody:htmlBody,
                            }
                            if err = mail.Send(ctx, msg); err != nil {
                                log.Errorf(ctx, "Could not send renew email: %v", err)
                                retry_tomorrow = true
                            }
                        }
                    }

                    // update redirect field to reflect having warned the user (whether failed or not)
                    if retry_tomorrow {
                        redirect.Warn += 1
                        expiration_warning_retry_count += 1
                    } else {
                        redirect.Warn = 127     // appears to have worked - don't try again
                    }
                    redirect.Expire = now.Unix() + (expiration_warning_days * seconds_per_day)
                    if _,err = datastore.Put(ctx,redirect_key,&redirect); err != nil {
                        log.Errorf(ctx,"error on datastore.Put redirect in cron; err = %v; redirect = %v; redirect_key = %v",err,redirect,redirect_key)
                    }

                }
            }
        }
    }
    log.Infof(ctx,"%d redirect records expired and deleted",redirect_expire_count)
    log.Infof(ctx,"expiration warnings sent on %d accounts (will retry on %d of those)",expiration_warning_count,expiration_warning_retry_count)

    log.Infof(ctx,"^^^^^^^^^^^^^^^^^^^^^^^^^^^ ADMIN_HANDLER ^^^^^^^^^^^^^^^^^^^^^^^^^^^")
}

func _create_dummy_redirect(ctx context.Context,i int) {
    var duration, days_ago int
    var redirect CatchyLinkRedirect
    var key *datastore.Key

    // variety of durations
    if i % 4 == 0 {
        duration = 1
    } else if i % 4 == 1 {
        duration = 7
    } else if i % 4 == 2 {
        duration = 31
    } else if i % 4 == 3 {
        duration = 365
    }

    // variety of days ago that the redirections were made
    days_ago = i % 6

    redirect = CatchyLinkRedirect {
        LongUrl: "https://google.com/" + strconv.Itoa(i),
        CatchyUrl: "GoOgLe/" + strconv.Itoa(i),
        Email: local_debugging_email,
        Expire: time.Now().Unix() + (int64(duration - days_ago) * seconds_per_day),
        Duration: int16(duration),
        OptF: 0,
        Warn: 0,
    }
    key = datastore.NewKey(ctx,"redirect",strings.ToLower(redirect.CatchyUrl),0,nil)
    datastore.Put(ctx,key,&redirect)
}

func _create_dummy_request(i int) {



}


func build_local_debug_db(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)

    for i := 0; i < 100; i++  {
        _create_dummy_redirect(ctx,i)
    }
    for i := 0; i < 30; i++  {
        _create_dummy_request(i)
    }

    input_form_with_message(w,"","","<br/><br/>LOCAL DEBUG DB BUILT",nil)
}

func robots_txt_handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "user-agent: *\r\nAllow: /$\r\nDisallow: /\r\n")
}

func favicon_ico_handler(w http.ResponseWriter, r *http.Request) {
    redirect_to_url(w,r,"https://googledrive.com/host/0B4rxOB63nnDMdE10cnlDWGxDSUU")
}