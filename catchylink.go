package catchylink

import (
    "os"
    "fmt"
    "html"
    "time"
    "strings"
    "io/ioutil"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/datastore"
    "google.golang.org/appengine/log"
)

const RequestTimeMin = 10       // requests will timeout in this many minutes


type CatchyLinkRequest struct {
        longurl, catchyurl, youremail string
        expire   time.Time
}

type FormInput struct {
    longurl, catchyurl, youremail string
}

var index_html string

func init() {

    // read index.html only once, so we don't read it again and again and again
    bytes, err := ioutil.ReadFile("web/index.html")
    if err != nil {
        fmt.Fprintf(os.Stderr,"YIKES!!!! Cannot read web/index.html");
    } else {
        index_html = string(bytes)
    }

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
    form.longurl = strings.TrimSpace(r.PostFormValue("longurl"))
    form.catchyurl = strings.TrimSpace(r.PostFormValue("catchyurl"))
    form.youremail = strings.TrimSpace(r.PostFormValue("youremail"))

    // VALIDATE THE INPUT
    if errormsg = errormsg_if_blank(form.longurl,"Long URL"); errormsg!="" {
        homepage_with_error_msg(w,"longurl",errormsg,form)
        return
    }
    if errormsg = errormsg_if_blank(form.catchyurl,"Catchy URL"); errormsg!="" {
        homepage_with_error_msg(w,"catchyurl",errormsg,form)
        return
    }
    if errormsg = errormsg_if_blank(form.youremail,"Your Email"); errormsg!="" {
        homepage_with_error_msg(w,"youremail",errormsg,form)
        return
    }
    if strings.ContainsAny(form.longurl," \t\r\n") {
        homepage_with_error_msg(w,"longurl","Long URL cannot contain space characters",form)
        return
    }
    if strings.ContainsAny(form.catchyurl," \t\r\n") {
        homepage_with_error_msg(w,"catchyurl","Catchy URL cannot contain space characters",form)
        return
    }
    if strings.ContainsAny(form.catchyurl,"+%") {
        homepage_with_error_msg(w,"catchyurl","Catchy URL cannot contain characters \"+\" or \"%\"",form)
        return
    }
    if 250 < len(form.longurl) {
        homepage_with_error_msg(w,"longurl","Long URL is too long (keep it under 250)",form)
        return
    }
    if 250 < len(form.catchyurl) {
        homepage_with_error_msg(w,"catchyurl","Catchy URL is too long (keep it under 250)",form)
        return
    }
    if 250 < len(form.youremail) {
        homepage_with_error_msg(w,"youremail","Your Email is too long (keep it under 250)",form)
        return
    }

    // create CatchyLinkRequest and inform user about it
    linkRequest := CatchyLinkRequest {
        longurl: form.longurl,
        catchyurl: form.catchyurl,
        youremail: form.youremail,
        expire: time.Now().Add( time.Duration(RequestTimeMin*1000*1000*1000) ),
    }
    _, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "linkrequest", nil), &linkRequest)
    if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
    }

    homepage(w)
}

func homepage(w http.ResponseWriter) {
    fmt.Fprint(w,index_html)
}

func homepage_with_error_msg(w http.ResponseWriter,fieldname string,errormsg string,form FormInput) {
    var page string
    page = strings.Replace(index_html,"{{"+fieldname+"-style}}","display:inline;",1)
    page = strings.Replace(page,"{{"+fieldname+"-errormsg}}",errormsg,1)

    page = strings.Replace(page,"{{longurl-value}}","value=\"" + html.EscapeString(form.longurl) + "\"",1)
    page = strings.Replace(page,"{{catchyurl-value}}","value=\"" + html.EscapeString(form.catchyurl) + "\"",1)
    page = strings.Replace(page,"{{youremail-value}}","value=\"" + html.EscapeString(form.youremail) + "\"",1)

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