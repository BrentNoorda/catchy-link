package catchylink

import (
    "fmt"
    "html"
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

func add_scheme_if_missing(inUrl string) (outUrl, errormsg string) {
    outUrl = inUrl
    if strings.Index(outUrl,"://") < 1 {
        outUrl = "http://" + outUrl
        errormsg = "\"http://\"" + " added to Long URL. If that is OK then resubmit this form."
    }
    return
}

func violates_special_email_root_rule(ctx context.Context,lCatchyUrl,lEmail string) bool { // catchylinks starting with email must belong to user
    // this should be easy with a regular expression - but I don't feel like taking the time to work out the correct regexp
    var url_parts []string
    var email_parts []string

    url_parts = strings.Split(lCatchyUrl,"/")
    email_parts = strings.Split(url_parts[0],"@")

    if len(email_parts) == 2  &&
       email_parts[0] != ""   &&
       email_parts[1] != "" {
        // has something@something
        var domain_parts []string
        domain_parts = strings.Split(email_parts[1],".")
        if len(domain_parts) >= 2  &&
           domain_parts[0] != ""   &&
           domain_parts[1] != "" {
            // has something@something.something

            if url_parts[0] != lEmail {
                return true
            }
        }
    }
    return false
}

func input_form_success(w http.ResponseWriter,linkRequest CatchyLinkRequest,sender_email_address string) {
    var page string
    page = strings.Replace(input_form_success_html(),"{{longurl_a}}",strings.Replace(linkRequest.LongUrl,"\"","&quot;",1),1)
    page = strings.Replace(page,"{{longurl_t}}",html.EscapeString(linkRequest.LongUrl),1)
    page = strings.Replace(page,"{{shorturl_t}}",html.EscapeString(linkRequest.CatchyUrl),1)
    page = strings.Replace(page,"{{youremail}}",html.EscapeString(linkRequest.Email),1)
    page = strings.Replace(page,"{{myemail}}",sender_email_address,1)
    fmt.Fprint(w,page)
}

func prepare_request_email_body(linkRequest CatchyLinkRequest, doitUrl string) (body,htmlBody string) {

    var noUrlLink string

    body = "You have requested a memorable URL to redirect:\n\n" +
           "   " + strings.Replace(myRootUrl,"//","// ",1) + "/" + linkRequest.CatchyUrl + "\n\n" +
           "to\n\n" +
           "   " + linkRequest.LongUrl + "\n\n\n" +
           "To VERIFY this url request, click on the following link (or copy and paste it to the address field in your browser):\n\n" +
           "   VERIFY: " + doitUrl + "\n"

    // make url disguised so email reader doens't automatically make it a link
    noUrlLink = myRootUrl + "/" + linkRequest.CatchyUrl
    noUrlLink = strings.Replace(noUrlLink,"/","<font>/</font>",-1)
    noUrlLink = strings.Replace(noUrlLink,".","<font>.</font>",-1)
    noUrlLink = strings.Replace(noUrlLink,":","<font>:</font>",-1)

    htmlBody = "<table width=\"97%\" style=\"margin: auto;max-width:800px\" align=\"center\">\n" +
               "<tr><td width=\"100%\">\n" +
               "You have requested a memorable URL to redirect:<br/><br/>\n" +
               " &nbsp; " + noUrlLink + "<br/><br/>\n" +
               "to<br/><br/>\n" +
               " &nbsp; <a href=\"" + linkRequest.LongUrl + "\">" + linkRequest.LongUrl + "</a><br/><br/>\n" +
               "To VERIFY this url request, click on the following button:<br/><br/>\n" +
               " &nbsp; <a href=\"" + doitUrl + "\"><button style=\"background-color:#dddddd;\"><font size=\"+1\">create catchy.link</font></button></a><br/><br/><br/>\n" +
               "<font size=\"-2\">if that button fails, copy and paste this url into your browser: " + doitUrl + "</font>" +
               "</td></tr></table>"

     return
}

func input_form(w http.ResponseWriter) {
    fmt.Fprint(w,input_form_html())
}

func input_form_with_message(w http.ResponseWriter,fieldname string,errormsg string,extramsg string,form *FormInput) {
    var page string

    page = strings.Replace(input_form_html(),"{{"+fieldname+"-style}}","display:inline;",1)
    page = strings.Replace(page,"{{"+fieldname+"-table-style}}","display:table-row;",1)
    page = strings.Replace(page,"{{"+fieldname+"-errormsg}}",errormsg,1)

    page = strings.Replace(page,"<!--etc-->",extramsg,1)

    if form != nil {
        page = strings.Replace(page,"selected=\"selected\"","",1)
        page = strings.Replace(page,"{{selected-" + form.Duration + "}}","selected=\"selected\"",1)

        page = strings.Replace(page,"{{longurl-value}}","value=\"" + html.EscapeString(form.LongUrl) + "\"",1)
        page = strings.Replace(page,"{{catchyurl-value}}","value=\"" + html.EscapeString(form.CatchyUrl) + "\"",1)
        page = strings.Replace(page,"{{youremail-value}}","value=\"" + html.EscapeString(form.Email) + "\"",1)
    }

    fmt.Fprint(w,page)
}

func does_this_catchy_url_belong_to_someone_else(ctx context.Context, lCatchyUrl, lemail string, now time.Time) bool {
    var err error
    var key *datastore.Key
    var redirect CatchyLinkRedirect

    log.Infof(ctx,"------------------- does_this_catchy_url_belong_to_someone_else --------")
    key = datastore.NewKey(ctx,"redirect",lCatchyUrl,0,nil)
    if err = datastore.Get(ctx, key, &redirect); err != nil {
        // there is no existing record
        log.Infof(ctx,"FALSE: because err != nil; err = %v for entity key %v",err,key)
        return false
    } else if ( redirect.Expire <= now.Unix() ) {
        // the existing record has expired
        log.Infof(ctx,"FALSE: because existing record has expired")
        return false
    } else if ( strings.ToLower(redirect.Email) == lemail ) {
        // existing record belongs to this user
        log.Infof(ctx,"FALSE: because existing record belongs to this email address")
        return false
    } else {
        // record is valid, has not timed out, and belongs to another user
        log.Infof(ctx,"TRUE: because existing valid record belongs to someone else")
        return true
    }
}

func post_new_catchy_link(w http.ResponseWriter, r *http.Request) {
    var errormsg string
    var err error
    var lEmail string
    ctx := appengine.NewContext(r)

    r.ParseForm()
    var form FormInput
    form.LongUrl = strings.TrimSpace(r.PostFormValue("longurl"))
    form.CatchyUrl = strings.TrimSpace(r.PostFormValue("catchyurl"))
    form.Email = strings.TrimSpace(r.PostFormValue("youremail"))
    form.Duration = r.PostFormValue("duration")
    if form.Duration != "1" && form.Duration != "7" && form.Duration != "31" && form.Duration != "365" {
        form.Duration = "7"
    }

    // remove / from the end of the CatchyUrl (they cause problems)
    form.CatchyUrl = strings.TrimRight(form.CatchyUrl,"/ \n\r\t")

    if r.PostFormValue("from404") != "" {
        // this comes directly from a 404 page where someone asked to try again, so just return
        // page
        input_form_with_message(w,"","","",&form)
        return
    }

    lCatchyUrl := strings.ToLower(form.CatchyUrl)

    // VALIDATE THE INPUT
    if errormsg = errormsg_if_blank(form.LongUrl,"Long URL"); errormsg!="" {
        input_form_with_message(w,"longurl",errormsg,"",&form)
        return
    }
    if errormsg = errormsg_if_blank(form.CatchyUrl,"Catchy URL"); errormsg!="" {
        input_form_with_message(w,"catchyurl",errormsg,"",&form)
        return
    }
    if errormsg = errormsg_if_blank(form.Email,"Your Email"); errormsg!="" {
        input_form_with_message(w,"youremail",errormsg,"",&form)
        return
    }
    if strings.ContainsAny(form.LongUrl," \t\r\n") {
        input_form_with_message(w,"longurl","Long URL cannot contain space characters","",&form)
        return
    }
    if strings.ContainsAny(form.CatchyUrl," \t\r\n") {
        input_form_with_message(w,"catchyurl","Catchy URL cannot contain space characters","",&form)
        return
    }
    if strings.ContainsAny(form.CatchyUrl,"#") {
        input_form_with_message(w,"catchyurl","Catchy URL cannot contain the character \"#\"","",&form)
        return
    }
    if strings.ContainsAny(form.CatchyUrl,"+%") {
        input_form_with_message(w,"catchyurl","Catchy URL cannot contain characters \"+\" or \"%\"","",&form)
        return
    }
    if 1000 < len(form.LongUrl) {
        input_form_with_message(w,"longurl","Long URL is too long (keep it under 1000)","",&form)
        return
    }
    if 250 < len(lCatchyUrl) {
        input_form_with_message(w,"catchyurl","Catchy URL is too long (keep it under 250)","",&form)
        return
    }
    if 150 < len(form.Email) {
        input_form_with_message(w,"youremail","Your Email is too long (keep it under 150)","",&form)
        return
    }

    // check that it's not one of our few disallowed files
    for _, each := range disallowed_roots {
        if strings.HasPrefix(lCatchyUrl,each) {
            input_form_with_message(w,"catchyurl","Catchy URL cannot begin with \"" + each + "\"","",&form)
            return
        }
    }

    // a few special checks for special situations
    if form.LongUrl, errormsg = add_scheme_if_missing(form.LongUrl); errormsg != "" {
        input_form_with_message(w,"longurl",errormsg,"",&form)
        return
    }

    lEmail = strings.ToLower(form.Email)
    if violates_special_email_root_rule(ctx,lCatchyUrl,lEmail) {
        input_form_with_message(w,"catchyurl",
                                "This catchy.link appears to begin with an email address, but it does not match your email address. There is a special " +
                                "rule that any catchy.link beginning with an email address \"belongs\" to the person with that email address, and so " +
                                "may be submitted only by the person with that email address.",
                                "",&form)
        return
    }


    now := time.Now()

    // check that this record doesn't already exist in the DB
    if does_this_catchy_url_belong_to_someone_else(ctx,lCatchyUrl,lEmail,now) {
        input_form_with_message(w,"catchyurl","This catchy.link was already taken by someone else. Sorry.","",&form)
        return
    }

    // create CatchyLinkRequest and inform user about it
    expire := now.Add( time.Duration(RequestTimeMin*60*(1000*1000*1000)) )
    duration,_ := strconv.Atoi(form.Duration)
    linkRequest := CatchyLinkRequest {
        UniqueKey: random_string(55),
        LongUrl: form.LongUrl,
        CatchyUrl: form.CatchyUrl,
        Email: form.Email,
        Expire: expire.Unix(),
        Duration: int16(duration),
        OptF: 0,
    }
    var key *datastore.Key
    key, err = datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "linkrequest", nil), &linkRequest)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    doitUrl := fmt.Sprintf("%s/~/doit/%d/%s",myRootUrl,key.IntID(),linkRequest.UniqueKey)
    //cancelUrl := fmt.Sprintf("%s/~/cancel/%d/%s",myRootUrl,key.IntID(),linkRequest.UniqueKey)
    body,htmlBody := prepare_request_email_body(linkRequest,doitUrl)
    subject := "Verify URL on Catchy.Link"

    log.Infof(ctx,"-------------------------------------------------------------")
    log.Infof(ctx,"To: %s",form.Email)
    log.Infof(ctx,"Subject: %s\n",subject)
    log.Infof(ctx,"%s",body)
    log.Infof(ctx,"-------------------------------------------------------------")

    var sender_email_address string
    if Mailgun != nil {
        // send email message through mailgun
        httpc := urlfetch.Client(ctx)

        mg := mailgun.NewMailgun(
            Mailgun.domain_name,
            Mailgun.secret_key,
            Mailgun.public_key,
        )
        mg.SetClient(httpc)

        sender_email_address = Mailgun.from
        message := mg.NewMessage(
             /* From */ Mailgun.from,
             /* Subject */ subject,
             /* Body */ body,
             /* To */ form.Email,
        )
        message.SetHtml(htmlBody)

        _, _, err = mg.Send(message)
        if err != nil {
            log.Errorf(ctx, "Could not send email from Mailgun: %v", err)
            input_form_with_message(w,"youremail","Unspecified error sending email to that address. Sorry.","",&form)
            return
        }

    } else {
        // send email message through app engine
        sender_email_address = sender_email_address_if_no_mailgun
        msg := &mail.Message{
            Sender:  sender_email_address_if_no_mailgun,
            To:      []string{form.Email},
            Subject: subject,
            Body:    body,
            HTMLBody:htmlBody,
        }
        if err = mail.Send(ctx, msg); err != nil {
            log.Errorf(ctx, "Could not send email: %v", err)
            input_form_with_message(w,"youremail","Unspecified error sending email to that address. Sorry.","",&form)
            return
        }
    }

    input_form_success(w,linkRequest,sender_email_address)
}
