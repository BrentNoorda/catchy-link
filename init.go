package catchylink

import (
    "os"
    "fmt"
    "time"
    "io/ioutil"
    "net/http"
    "math/rand"
)

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
    http.HandleFunc("/", redirect_handler)
}
