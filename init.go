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
        fmt.Fprintf(os.Stderr,"YIKES!!!! Cannot read web/index.html");
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
    index_html = read_min_web_file("index.html")

    http.HandleFunc("/robots.txt", robots_txt_handler)
    http.HandleFunc("/-/", admin_handler)
    http.HandleFunc("/~/", email_response_handler)
    http.HandleFunc("/", redirect_handler)
}
