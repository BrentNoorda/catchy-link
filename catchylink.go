package catchylink

import (
    "os"
    "fmt"
    "html"
    "time"
    "strings"
    "strconv"
    "io/ioutil"
    "net/http"
    "math/rand"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
    "google.golang.org/appengine/mail"
    "google.golang.org/appengine/datastore"
)

const myRootUrl = "http://catchy.link"
const RequestTimeMin = 10       // requests will timeout in this many minutes
const sender_email_address = "emailer@catchy-link.appspotmail.com"

var disallowed_roots = [...]string {
    "index.",
    "favicon.ico",
    "robots.txt",
    "_/",
    "~/",
    "-/",
}

type CatchyLinkRequest struct {
    UniqueKey string
    LongUrl, CatchyUrl, YourEmail string
    Expire   int64
}

type FormInput struct {
    LongUrl, CatchyUrl, YourEmail string
}

func random_string(minLen int) string {
    ret := ""
    for len(ret) < minLen {
        ret += strconv.Itoa(int(rand.Uint32()))
    }
    return ret
}

var index_html string

func init() {

    rand.Seed(time.Now().UnixNano())

    // read index.html only once, so we don't read it again and again and again
    bytes, err := ioutil.ReadFile("web/index.html")
    if err != nil {
        fmt.Fprintf(os.Stderr,"YIKES!!!! Cannot read web/index.html");
    } else {
        index_html = string(bytes)
    }

    http.HandleFunc("/robots.txt", robots_txt_handler)
    http.HandleFunc("/-/", admin_handler)
    http.HandleFunc("/~/", email_response_handler)
    http.HandleFunc("/", handler)
}

func errormsg_if_blank(value string,fieldDescription string) string {
    if value == "" {
        return fieldDescription + " must not be blank"
    }
    return ""
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
        homepage_with_error_msg(w,"longurl",errormsg,&form)
        return
    }
    if errormsg = errormsg_if_blank(form.CatchyUrl,"Catchy URL"); errormsg!="" {
        homepage_with_error_msg(w,"catchyurl",errormsg,&form)
        return
    }
    if errormsg = errormsg_if_blank(form.YourEmail,"Your Email"); errormsg!="" {
        homepage_with_error_msg(w,"youremail",errormsg,&form)
        return
    }
    if strings.ContainsAny(form.LongUrl," \t\r\n") {
        homepage_with_error_msg(w,"longurl","Long URL cannot contain space characters",&form)
        return
    }
    if strings.ContainsAny(form.CatchyUrl," \t\r\n") {
        homepage_with_error_msg(w,"catchyurl","Catchy URL cannot contain space characters",&form)
        return
    }
    if strings.ContainsAny(form.CatchyUrl,"+%") {
        homepage_with_error_msg(w,"catchyurl","Catchy URL cannot contain characters \"+\" or \"%\"",&form)
        return
    }
    if 250 < len(form.LongUrl) {
        homepage_with_error_msg(w,"longurl","Long URL is too long (keep it under 250)",&form)
        return
    }
    if 250 < len(lowerCatchyUrl) {
        homepage_with_error_msg(w,"catchyurl","Catchy URL is too long (keep it under 250)",&form)
        return
    }
    if 250 < len(form.YourEmail) {
        homepage_with_error_msg(w,"youremail","Your Email is too long (keep it under 250)",&form)
        return
    }

    // check that it's not one of our few disallowed files
    for _, each := range disallowed_roots {
        if strings.HasPrefix(lowerCatchyUrl,each) {
            homepage_with_error_msg(w,"catchyurl","Cathy URL cannot begin with \"" + each + "\"",&form)
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


    homepage(w)
}

func homepage(w http.ResponseWriter) {
    fmt.Fprint(w,index_html)
}

func homepage_with_error_msg(w http.ResponseWriter,fieldname string,errormsg string,form *FormInput) {
    var page string
    page = strings.Replace(index_html,"{{"+fieldname+"-style}}","display:inline;",1)
    page = strings.Replace(page,"{{"+fieldname+"-errormsg}}",errormsg,1)

    if form != nil {
        page = strings.Replace(page,"{{longurl-value}}","value=\"" + html.EscapeString(form.LongUrl) + "\"",1)
        page = strings.Replace(page,"{{catchyurl-value}}","value=\"" + html.EscapeString(form.CatchyUrl) + "\"",1)
        page = strings.Replace(page,"{{youremail-value}}","value=\"" + html.EscapeString(form.YourEmail) + "\"",1)
    }

    fmt.Fprint(w,page)
}


func handler(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)
    log.Infof(ctx,"%s","Catchylink3, world!<br/>Path:\"" + r.URL.Path + "\"  RawPath:\"" + r.URL.RawPath + "\"  RawQuery:\"" + r.URL.RawQuery + "\"")
    if r.URL.Path == "/" {
        if r.Method == "POST" {
            post_new_catchy_link(w,r)
        } else {
            homepage(w)
        }
    } else {
        fmt.Fprint(w, "Catchylink3, world!<br/>Path:" + r.URL.Path + "<br/>RawPath:" + r.URL.RawPath + "<br/>RawQuery:" + r.URL.RawQuery)
    }
}

func robots_txt_handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "user-agent: *\r\nAllow: /$\r\nDisallow: /\r\n")
}

func email_response_handler(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)
    parts := strings.Split(r.URL.Path,"/")
    if len(parts) < 5 {
        log.Errorf(ctx,"email_reponse_handler weird URL \"%s\"",r.URL.Path)
        homepage_with_error_msg(w,"globalerror","Unrecognized URL",nil)
    } else {
        command := parts[2]
        dbid, err := strconv.Atoi(parts[3])
        uniqueKey := parts[4]
        if err != nil {
            log.Errorf(ctx,"email_reponse_handler weird URL \"%s\"\nerror: %v",r.URL.Path,err)
            homepage_with_error_msg(w,"globalerror","Unrecognized URL",nil)
        } else {
            if command == "doit"  ||  command == "cancel" {
                log.Errorf(ctx,"dbid = %d, uniqueKey = %s",dbid,uniqueKey)
                homepage(w)
            } else {
                log.Errorf(ctx,"email_reponse_handler weird URL \"%s\"",r.URL.Path)
                homepage_with_error_msg(w,"globalerror","Unrecognized URL",nil)
            }
        }
    }
}

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