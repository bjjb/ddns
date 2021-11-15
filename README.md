# CloudFlare Dynamic Domain Name Servicer

*cfddns* is a little client which allows you to update your CloudFlare zones
with a record for the current machine. It's also a little server that lets you
do the same thing, if that's your bag.

## Installation

    go install github.com/bjjb/cfddns

### Usage

    cfddns -h

will give you information on how to use it.

It's probably most useful to run it with cron.

You may also run the server:

    cfddns -d

That starts the server on [::]:8080 and 0.0.0.0:8080; you can control the
server's operation with the `ADDR` environment variable (to set the listen
address). The server expects (basic) authorization header which is used to
authenticate with CloudFlare, and a path consisting of the domain name that
the client wants.  `cfddns` then checks for a zone matching the higher-level
domain(s) of the requested domain (returning 404 if it can't find one), and
creates or updates an A or AAAA record therein with the client's IP address.
Clients may specify parameters to override the default behaviour:

| Parameter | Meaning                                               |
| --------- | ----------------------------------------------------- |
| zone      | The CloudFlare zone ID containing the record          |
| record    | The CloudFlare record ID to update                    |
| type      | A, AAAA or CNAME                                      |
| service   | Use something besides `api.cloudflare.com/client/v4`  |
