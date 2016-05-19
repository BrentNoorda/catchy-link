package catchylink

import (
    "os"
    "fmt"
    "time"
    "strings"
    "strconv"
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

func read_min_web_file(filespec string,replace_css, replace_root_url, replace_google_analytics bool) string {
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
        if replace_css {
            ret = strings.Replace(ret,"<link rel=\"stylesheet\" media=\"screen\" type=\"text/css\" href=\"catchylink.css\">",
                                  "<style media=\"screen\" type=\"text/css\">\n"+catchylink_css()+"\n</style>",1)
        }
        if replace_root_url {
            ret = strings.Replace(ret,"{{catchylink_root_url}}",myRootUrl,-1)
        }
        if replace_google_analytics {
            ret = strings.Replace(ret,"<!--google-analytics-->",google_analytics_txt(),1)
        }
    }
    return ret
}

// read index.html only once, so we don't read it again and again and again
var _input_form_html string = ""
var _input_form_success_html string = ""
var _email_doit_success_html string = ""
var _notfound_404_form_html string = ""
var _catchylink_css string = ""
var _google_analytics_txt string = ""
var _embedded_iframe_html string = ""
var _prompt_redirect_html string = ""
var _prompt_redirect_with_email_html string = ""

func input_form_html() string {
    if _input_form_html == "" {
        text := read_min_web_file("input_form.html",true,true,true)
        if local_debugging {
            text = strings.Replace(text,"</body>","<a style=\"float:right;color:#eeeeee;\" href=\"/~/build-local-debug-db\">BUILD</a></body>",1)
        }
        _input_form_html = text
    }
    return _input_form_html
}

func input_form_success_html() string {
    if _input_form_success_html == "" {
        _input_form_success_html = read_min_web_file("input_form_success.html",true,true,true)
    }
    return _input_form_success_html
}

func email_doit_success_html() string {
    if _email_doit_success_html == "" {
        _email_doit_success_html = read_min_web_file("email_doit_success.html",true,true,true)
    }
    return _email_doit_success_html
}

func notfound_404_form_html() string {
    if _notfound_404_form_html == "" {
        _notfound_404_form_html = read_min_web_file("notfound_404_form.html",true,true,true)
    }
    return _notfound_404_form_html
}

func prompt_redirect_html() string {
    if _prompt_redirect_html == "" {
        _prompt_redirect_html = read_min_web_file("prompt_redirect.html",true,true,true)
    }
    return _prompt_redirect_html
}

func prompt_redirect_with_email_html() string {
    if _prompt_redirect_with_email_html == "" {
        _prompt_redirect_with_email_html = read_min_web_file("prompt_redirect_with_email.html",true,true,true)
    }
    return _prompt_redirect_with_email_html
}

func catchylink_css() string {
    if _catchylink_css == "" {
        css := read_min_web_file("catchylink.css",false,false,false)
        css = strings.Replace(css,": ",":",-1)
        _catchylink_css = css
    }
    return _catchylink_css
}

func google_analytics_txt() string {
    if _google_analytics_txt == "" {
        _google_analytics_txt = read_min_web_file("google_analytics.txt",false,false,false)
    }
    return _google_analytics_txt
}

func embedded_iframe_html() string {
    if _embedded_iframe_html == "" {
        _embedded_iframe_html = read_min_web_file("embedded_iframe.html",false,false,false)
    }
    return _embedded_iframe_html
}

func init() {
    rand.Seed(time.Now().UnixNano())

    if txt := os.Getenv("CATCHYLINK_ROOT_URL"); txt != "" {
        myRootUrl = txt
    }
    if txt := os.Getenv("CATCHYLINK_SECONDS_PER_DAY"); txt != "" {
        seconds_per_day,_ = strconv.ParseInt(txt,10,64)
    }
    if os.Getenv("CATCHYLINK_LOCAL_DEBUGGING") != "" {
        local_debugging = true
    }
    if txt := os.Getenv("CATCHYLINK_LOCAL_DEBUGGING_EMAIL"); txt != "" {
        local_debugging_email = txt
    }
    fmt.Fprintf(os.Stderr,"INIT\nmyRootUrl = %s\nsec/day = %d\nlocal_debugging = %t\n",myRootUrl,seconds_per_day,local_debugging)

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
    if local_debugging {
        http.HandleFunc("/~/build-local-debug-db", build_local_debug_db)
    }
    http.HandleFunc("/~/embtst", test_embed_handler)
    http.HandleFunc("/~/", email_response_handler)
    http.HandleFunc("/", redirect_handler)
    //http.HandleFunc("/~/onetime_db_fixup_pefqpqpouqwifqpfiqfhqwfqfuiqef", onetime_db_fixup)
}
