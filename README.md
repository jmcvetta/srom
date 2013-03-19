# Sucks-Rules-O-Meter

`srom` is a command line tool, and accompanying library, that programmatically
queries search engines to determine wether the internet thinks a given term
sucks or rules.

## Install

`srom` requires a working [Go](http://golang.org) installation.

```
$ go install github.com/jmcvetta/srom
```

## Usage

The first time you run `srom` it will create a configuration file and prompt
you to populate it with Google and Azure API credentials.  You must provide
credentials for at least one search engine.  If you only want to use one
engine, remove the other engine's section from the config file.

```
$ srom 'johnny cash'
20:39:08 srom.go:92:           Google Custom Search         -1    -1   1.000000
20:39:08 srom.go:92:              Azure Data Market       2290   429   5.337995

The term 'johnny cash' has a rules/sucks ratio of 5.337995337995338

```

In the example shown, the call to Google has failed because we are over the API
request limit for the day.  The failure is handled gracefully, and does not
impact the reported ratio.


## Documentation

See GoDoc for [automatic documentation](http://godoc.org/github.com/jmcvetta/srom).


## Supported Search Engines

Currently `srom` supports Google's crappy "Custom Search" API and Microsoft's
likewise crappy Bing API.  Alas, both of these APIs returns hit counts orders
of magnitude fewer than searches on their respective websites.  

I would be happy to add support for other - perhaps less craptastic - search
APIs.  Please alert me if you know of one.  Alas, Duck-Duck-Go's API is even
more crippled, as it does not return hit count for a query.


## Status

Works.  No tests (yet?)
