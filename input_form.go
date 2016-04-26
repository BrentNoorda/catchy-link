package catchylink

import (
    "fmt"
    "html"
    "time"
    "strings"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
    "google.golang.org/appengine/mail"
    "google.golang.org/appengine/datastore"
)


func input_form_success(w http.ResponseWriter,linkRequest CatchyLinkRequest) {
    var page string
    page = strings.Replace(input_form_success_html,"{{longurl_a}}",strings.Replace(linkRequest.LongUrl,"\"","&quot;",1),1)
    page = strings.Replace(page,"{{longurl_t}}",html.EscapeString(linkRequest.LongUrl),1)
    page = strings.Replace(page,"{{shorturl_t}}",html.EscapeString(linkRequest.CatchyUrl),1)
    page = strings.Replace(page,"{{youremail}}",html.EscapeString(linkRequest.YourEmail),1)
    fmt.Fprint(w,page)
}

func input_form(w http.ResponseWriter) {
    fmt.Fprint(w,input_form_html)
}

func input_form_with_error_msg(w http.ResponseWriter,fieldname string,errormsg string,form *FormInput) {
    var page string
    page = strings.Replace(input_form_html,"{{"+fieldname+"-style}}","display:inline;",1)
    page = strings.Replace(page,"{{"+fieldname+"-errormsg}}",errormsg,1)

    if form != nil {
        page = strings.Replace(page,"{{longurl-value}}","value=\"" + html.EscapeString(form.LongUrl) + "\"",1)
        page = strings.Replace(page,"{{catchyurl-value}}","value=\"" + html.EscapeString(form.CatchyUrl) + "\"",1)
        page = strings.Replace(page,"{{youremail-value}}","value=\"" + html.EscapeString(form.YourEmail) + "\"",1)
    }

    fmt.Fprint(w,page)
}

func post_new_catchy_link(w http.ResponseWriter, r *http.Request) {
    var errormsg string
    ctx := appengine.NewContext(r)

    r.ParseForm()
    var form FormInput
    form.LongUrl = strings.TrimSpace(r.PostFormValue("longurl"))
    form.CatchyUrl = strings.TrimSpace(r.PostFormValue("catchyurl"))
    form.YourEmail = strings.TrimSpace(r.PostFormValue("youremail"))
    lowerCatchyUrl := strings.ToLower(form.CatchyUrl)

    // VALIDATE THE INPUT
    if errormsg = errormsg_if_blank(form.LongUrl,"Long URL"); errormsg!="" {
        input_form_with_error_msg(w,"longurl",errormsg,&form)
        return
    }
    if errormsg = errormsg_if_blank(form.CatchyUrl,"Catchy URL"); errormsg!="" {
        input_form_with_error_msg(w,"catchyurl",errormsg,&form)
        return
    }
    if errormsg = errormsg_if_blank(form.YourEmail,"Your Email"); errormsg!="" {
        input_form_with_error_msg(w,"youremail",errormsg,&form)
        return
    }
    if strings.ContainsAny(form.LongUrl," \t\r\n") {
        input_form_with_error_msg(w,"longurl","Long URL cannot contain space characters",&form)
        return
    }
    if strings.ContainsAny(form.CatchyUrl," \t\r\n") {
        input_form_with_error_msg(w,"catchyurl","Catchy URL cannot contain space characters",&form)
        return
    }
    if strings.ContainsAny(form.CatchyUrl,"+%") {
        input_form_with_error_msg(w,"catchyurl","Catchy URL cannot contain characters \"+\" or \"%\"",&form)
        return
    }
    if 250 < len(form.LongUrl) {
        input_form_with_error_msg(w,"longurl","Long URL is too long (keep it under 250)",&form)
        return
    }
    if 250 < len(lowerCatchyUrl) {
        input_form_with_error_msg(w,"catchyurl","Catchy URL is too long (keep it under 250)",&form)
        return
    }
    if 250 < len(form.YourEmail) {
        input_form_with_error_msg(w,"youremail","Your Email is too long (keep it under 250)",&form)
        return
    }

    // check that it's not one of our few disallowed files
    for _, each := range disallowed_roots {
        if strings.HasPrefix(lowerCatchyUrl,each) {
            input_form_with_error_msg(w,"catchyurl","Cathy URL cannot begin with \"" + each + "\"",&form)
            return
        }
    }

    // create CatchyLinkRequest and inform user about it
    expire := time.Now().Add( time.Duration(RequestTimeMin*60*1000*1000*1000) )
    linkRequest := CatchyLinkRequest {
        UniqueKey: random_string(55),
        LongUrl: form.LongUrl,
        CatchyUrl: form.CatchyUrl,
        YourEmail: form.YourEmail,
        Expire: expire.Unix(),
    }
    key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "linkrequest", nil), &linkRequest)
    if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
    }

    doitUrl := fmt.Sprintf("%s/~/doit/%d/%s",myRootUrl,key.IntID(),linkRequest.UniqueKey)
    cancelUrl := fmt.Sprintf("%s/~/cancel/%d/%s",myRootUrl,key.IntID(),linkRequest.UniqueKey)
    log.Infof(ctx,"\n\n\ndoitUrl = %s\n\ncancelUrl = %s\n\n  ",doitUrl,cancelUrl)

    // send email to user
    msg := &mail.Message{
        Sender:  sender_email_address,
        To:      []string{form.YourEmail},
        Subject: "Email from CatchyLink",
        Body:    "Email from catchylink yes it is",
    }
    if err := mail.Send(ctx, msg); err != nil {
        log.Errorf(ctx, "Couldn't send email: %v", err)
    }

    input_form_success(w,linkRequest)
}
