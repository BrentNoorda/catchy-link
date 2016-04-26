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
    LongUrl, CatchyUrl, YourEmail string
    Expire   int64
}

type FormInput struct {
    LongUrl, CatchyUrl, YourEmail string
}

var index_html string
