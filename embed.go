package catchylink

import (
    "fmt"
    "html"
    "strings"
    "net/http"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
)

func embed_url(w http.ResponseWriter,url string,title string) {
    var page string
    page = strings.Replace(embedded_iframe_html(),"{{title-goes-here}}",html.EscapeString(title),1)
    page = strings.Replace(page,"{{url-goes-here}}",html.EscapeString(url),1)
    fmt.Fprint(w,page)
}

func test_embed_handler(w http.ResponseWriter, r *http.Request) {
    var url, cl string
    ctx := appengine.NewContext(r)

    url = r.URL.Query().Get("url")
    cl = r.URL.Query().Get("cl")

    log.Infof(ctx,"vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")

    log.Infof(ctx,"url = \"%s\"",url)
    log.Infof(ctx,"cl = \"%s\"",cl)

    log.Infof(ctx,"^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")

    embed_url(w,url,cl)
}
