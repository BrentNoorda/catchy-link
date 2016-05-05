package catchylink

import (
)

var myRootUrl = "http://catchy.link"  // this overridden if "CATCHYLINK_ROOT_URL" environment variable
const RequestTimeMin = 30       // requests will timeout in this many minutes
const expiration_warning_days = 3 // how many days before expiration will an email be sent out
const max_email_warning_retries = 3 // if cannot successfully email after this many days&retries, then give up

var seconds_per_day int64 = 60 * 60 * 24  // when debugging or developing locally, this number may be reduced
                                          // so we can wait minutes (for example) for stuff to time out instead of days
var local_debugging bool = false        // init() function may change this based on environment variables
var local_debugging_email string = "bad_email_address" // but good email here via environment variables

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
    UniqueKey   string      `datastore:",noindex"`
    LongUrl     string      `datastore:",noindex"`
    CatchyUrl   string      `datastore:",noindex"`
    Email       string      `datastore:",noindex"`
    Expire      int64
    Duration    int16       `datastore:",noindex"`  // duration in days
}

type FormInput struct {
    LongUrl, CatchyUrl, Email, Duration string
}

type CatchyLinkRedirect struct {  // key for this DB is lowercase-CatchyUrl
    LongUrl     string      `datastore:",noindex"`
    CatchyUrl   string      `datastore:",noindex"`
    Email       string      `datastore:",noindex"`
    Expire      int64                                   // when this expires, will be extended at least to
                                                        // expiration_warning_days when warning email is sent out
    Duration    int16       `datastore:",noindex"`      // original duration in days
    Warn        int8        `datastore:",noindex"`      // count how many times a warning email has gone out
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