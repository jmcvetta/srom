# srom

Sucks-Rules-O-Meter

`srom` is a command line tool, and accompanying library, that programmatically
queries search engines to determine wether the internet thinks a given term
sucks or rules.

## Install

`srom` requires a working [Go](http://golang.org) installation.

```bash
$ go install github.com/jmcvetta/srom
```

## Usage

```bash
$ srom 'johnny cash'
20:39:08 srom.go:92:              Azure Data Market       2290   429   5.337995

The term 'johnny cash' has a positive/negative ratio of 5.337995337995338

```

## Status

Works.  No tests (yet?)
