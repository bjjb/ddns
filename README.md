# Dynamic Domain Name Servicer

*ddns* is a little client which allows you to update your DNS zones with a
record for the current machine. It's also a little server that lets you do the
same thing, if that's your bag.

## Installation

You can obtain one of the releases, or install it using Go (>= 1.14) with

    go get github.com/bjjb/ddns

### Usage

    ddns -h

will give you information on how to use the command-line application.

It's probably most useful to run it with cron.

You may also run the server - see how with:

    ddns start -h

It contains some sub-commands for controlling a running server via its API.
