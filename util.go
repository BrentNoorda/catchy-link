package catchylink

import (
    "fmt"
    "html"
    "strings"
    "strconv"
    "net/http"
    "math/rand"
)

func random_string(minLen int) string {
    ret := ""
    for len(ret) < minLen {
        ret += strconv.Itoa(int(rand.Uint32()))
    }
    return ret
}

func errormsg_if_blank(value string,fieldDescription string) string {
    if value == "" {
        return fieldDescription + " must not be blank"
    }
    return ""
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
