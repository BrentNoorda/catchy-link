package catchylink

import (
)

var myRootUrl = "http://catchy.link"  // this overridden if "CATCHYLINK_ROOT_URL" environment variable
const RequestTimeMin = 30       // requests will timeout in this many minutes
const expiration_warning_days = 3 // how many days before expiration will an email be sent out
const max_email_warning_retries = 3 // if cannot successfully email after this many days&retries, then give up

var seconds_per_day int64 = 60 * 60 * 24  // when debugging or developing locally, this number may be reduced
                                          // so we can wait minutes (for example) for stuff to time out instead of days

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
    Warn     int8   // count how many times a warning email has gone out
}

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