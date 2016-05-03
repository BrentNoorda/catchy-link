# catchy-link
A URL shortener with a few differences:

* emphasis on memorable URLs, not short (and unpronounceable) URLs

* links can be altered after they have been created

* links are temporary, but can be renewed as many times as you want

* no accounts or passwords are needed, because verification happens through email

* everyone gets a special path for their email (e.g. catchy.link/joe@schmoe.com)

Interesting stuff for programmers:

* Runs on Google App Engine (with datastore, memcache, and cron)

* Follows the "passwords suck" philosophy by using email when verification is needed on actions

* Written in golang (although I've very little experience with Go, so don't be too critical)

* Sends email with Mailgun

Where to see more:

* Website in action: [http://catchy.link](http://catchy.link)

* Documentation: [catchy.link/CatchyLinkManual](http://catchy.link/CatchyLinkManual)

* Source Code: [github.com/BrentNoorda/catchy-link](https://github.com/BrentNoorda/catchy-link)
