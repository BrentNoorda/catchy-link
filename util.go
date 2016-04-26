package catchylink

import (
    "strconv"
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
