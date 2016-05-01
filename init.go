package catchylink

import (
    "os"
    "fmt"
    "time"
    "strings"
    "io/ioutil"
    "net/http"
    "math/rand"
)

func replace_all_repeatedly(s, old, new string) string {
    oldlen := len(s)
    for {
        s = strings.Replace(s,old,new,-1)
        strlen := len(s)
        if oldlen == strlen {
            break
        }
        oldlen = len(s)
    }
    return s
}

func read_min_web_file(filespec string) string {
    var ret string
    bytes, err := ioutil.ReadFile("web/" + filespec)
    if err != nil {
        fmt.Fprintf(os.Stderr,"YIKES!!!! Cannot read web/" + filespec);
        ret = ""
    } else {
        ret = string(bytes)
        ret = strings.Replace(ret,"\t"," ",-1)
        ret = replace_all_repeatedly(ret,"  "," ")
        ret = strings.Replace(ret,"\n ","\n",-1)
        ret = strings.Replace(ret," \n","\n",-1)
    }
    return ret
}

func init() {
    rand.Seed(time.Now().UnixNano())

    // read index.html only once, so we don't read it again and again and again
    input_form_html = read_min_web_file("input_form.html")
    input_form_success_html = read_min_web_file("input_form_success.html")
    email_doit_success_html = read_min_web_file("email_doit_success.html")

    // if Mailgun parameters are in the environment variables, read them now. Getting
    // those paramaters is an annoying kludge seen in run.py or deploy.py and writing
    // of some temp files from the /secret directory
    Mailgun = &MailgunParams{
        from: os.Getenv("MAILGUN_FROM"),
        domain_name: os.Getenv("MAILGUN_DOMAIN_NAME"),
        secret_key: os.Getenv("MAILGUN_SECRET_KEY"),
        public_key: os.Getenv("MAILGUN_PUBLIC_KEY"),
    }
    if ( Mailgun.from=="" ||Mailgun.domain_name=="" || Mailgun.secret_key=="" || Mailgun.public_key=="" ) {
        Mailgun = nil
    }

    http.HandleFunc("/robots.txt", robots_txt_handler)
    http.HandleFunc("/favicon.ico", favicon_ico_handler)
    http.HandleFunc("/-/", admin_handler)
    http.HandleFunc("/~/", email_response_handler)
    http.HandleFunc("/", redirect_handler)
}
