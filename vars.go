package catchylink

import (
)

var myRootUrl = "http://catchy.link"  // this overridden if "CATCHYLINK_ROOT_URL" environment variable
const RequestTimeMin = 30       // requests will timeout in this many minutes
const expiration_warning_days = 3 // how many days before expiration will an email be sent out

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
    LongUrl, CatchyUrl, Email, Duration string
}

type CatchyLinkRedirect struct {  // key for this DB is lowercase-CatchyUrl
    LongUrl, CatchyUrl, Email string
    Expire   int64  // when this expires, will be extended at least to expiration_warning_days when warning email is sent out
    Duration int16  // original duration in days
    Warned   bool   // has the expiration warning been sent (automatically set for timeout in 1-day because no email sent)
}

var input_form_html string
var input_form_success_html string
var email_doit_success_html string
var notfound_404_form_html string

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