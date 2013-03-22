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
16:31:41 srom.go:92:              Azure Data Market        435  3030   0.143564
16:31:41 srom.go:92:           Google Custom Search         92  5540   0.016606

The term 'johnny cash' has a sucks/rules ratio of 0.0800854273152947
The internet thinks 'johnny cash' ROCKS HARD.

```


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

[![Build Status](https://drone.io/github.com/jmcvetta/srom/status.png)](https://drone.io/github.com/jmcvetta/srom/latest)


## Credits

[Don Marti](https://github.com/dmarti) first introduced me to - and perhaps
invented? - the concept of a Sucks-Rules-O-Meter.  For the historically curious,
his site hosts a defunct [Operating System SROM](http://srom.zgp.org/).

