package catchylink

import (
)

const myRootUrl = "http://catchy.link"
const RequestTimeMin = 10       // requests will timeout in this many minutes
const sender_email_address = "verify@catchy-link.appspotmail.com"

var disallowed_roots = [...]string {
    "index.",
    "favicon.ico",
    "robots.txt",
    "_/",
    "~/",
    "-/",
}

type CatchyLinkRequest struct {
    UniqueKey string
    LongUrl, CatchyUrl, Email string
    Expire   int64
}

type FormInput struct {
    LongUrl, CatchyUrl, LCatchyUrl, Email string
}

type CatchyLinkRedirect struct {
    LCatchyUrl string   // this is the key for this database
    LongUrl, CatchyUrl, LEmail string
    Expire   int64
    Duration int16  // original duration in days
}

var input_form_html string
var input_form_success_html string
var email_doit_success_html string
