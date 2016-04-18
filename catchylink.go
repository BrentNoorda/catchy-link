package catchylink

import (
    "os"
    "fmt"
    "strings"
    "io/ioutil"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
)

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

    r.ParseForm()
    var form FormInput
    form.longurl = r.PostFormValue("longurl")
    form.catchyurl = r.PostFormValue("catchyurl")
    form.youremail = r.PostFormValue("youremail")

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

    homepage(w)
}

func homepage(w http.ResponseWriter) {
    fmt.Fprint(w,index_html)
}

func homepage_with_error_msg(w http.ResponseWriter,fieldname string,errormsg string,form FormInput) {
    var page string
    page = strings.Replace(index_html,"{{"+fieldname+"-style}}","display:inline;",1)
    page = strings.Replace(page,"{{"+fieldname+"-errormsg}}",errormsg,1)

    page = strings.Replace(page,"{{longurl-value}}","value=\"" + form.longurl + "\"",1)
    page = strings.Replace(page,"{{catchyurl-value}}","value=\"" + form.catchyurl + "\"",1)
    page = strings.Replace(page,"{{youremail-value}}","value=\"" + form.youremail + "\"",1)

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