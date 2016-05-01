package catchylink

import (
)

var myRootUrl = "http://catchy.link"  // this overridden if "CATCHYLINK_ROOT_URL" environment variable
const RequestTimeMin = 30       // requests will timeout in this many minutes

var disallowed_roots = [...]string {
    "index.",
    "favicon.ico",
    "robots.txt",
    "_/",
    "~/",
    "-/",
    "_ah/",
}

type CatchyLinkRequest struct {
    UniqueKey string
    LongUrl, CatchyUrl, Email string
    Expire   int64
    Duration int16  // original duration in days
}

type FormInput struct {
    LongUrl, CatchyUrl, LCatchyUrl, Email, Duration string
}

type CatchyLinkRedirect struct {  // key for this DB is lowercase-CatchyUrl
    LongUrl, CatchyUrl, Email string
    Expire   int64
    Duration int16  // original duration in days
}

var input_form_html string
var input_form_success_html string
var email_doit_success_html string

///////// EMAIL /////////
// use mailgun if Mailgun is not nil, else default to sender_email_address

const sender_email_address_if_no_mailgun = "verify@catchy-link.appspotmail.com"

type MailgunParams struct {
    from string
    domain_name string
    secret_key string
    public_key string
}

var Mailgun *MailgunParams  = nil