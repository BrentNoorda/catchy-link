package catchylink

import (
    "strconv"
    "math/rand"
)

func duration_to_string(duration int16) string {
    if duration == 1 {
        return "1 day"
    } else if duration == 7 {
        return "1 week"
    } else if duration == 31 {
        return "1 month"
    } else if duration == 365 {
        return "1 year"
    } else {
        return "???"
    }
}

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
