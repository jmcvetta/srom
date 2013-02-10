// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package main

import (
	"log"
	"time"
	"fmt"
	// "github.com/jmcvetta/restclient"
	// "github.com/stathat/go"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

type Configuration interface {
	Terms() ([]string, error)
	Positive() ([]string, error)
	Negative() ([]string, error)
}

type staticConf struct {}

func (sc *staticConf) Terms() ([]string, error) {
	t := []string{
		"ubuntu",
		"obama",
		"linux",
		"windows",
		"apple",
		"iPhone",
		"android",
		"iOS",
		"ed lee",
	}
	return t, nil
}

func (sc *staticConf) Positive() ([]string, error) {
	t := []string{
		"%v rules",
		"%v rocks",
		"%v is awesome",
		"%v kicks ass",
		"%v dominates",
		"love %v",
	}
	return t, nil
}

func (sc *staticConf) Negative() ([]string, error) {
	t := []string{
		"%v sucks",
		"%v blows",
		"%v is worthless",
		"%v is crap",
		"%v doesn't work",
		"hate %v",
	}
	return t, nil
}

type query struct {
	Term string
	PositiveQuery string
	NegativeQuery string
	Timestamp time.Time
	PositiveCount int
	NegativeCount int
}

func main() {
	conf := new(staticConf)
	terms, err := conf.Terms()
	if err != nil {
		log.Panic(err)
	}
	positives, err := conf.Positive()
	if err != nil {
		log.Panic(err)
	}
	negatives, err := conf.Negative()
	if err != nil {
		log.Panic(err)
	}
	queries := make(chan *query, 4)
	for _, t := range terms {
		q := query{
			Term: t,
		}
		q.PositiveQuery = fmt.Sprintf(positives[0], t)
		q.PositiveQuery = fmt.Sprintf("\"%v\"", q.PositiveQuery)
		for _, tmpl := range positives[1:] {
			s := fmt.Sprintf(tmpl, t)
			s = fmt.Sprintf(" OR \"%v\"", s)
			q.PositiveQuery += s
		}
		q.NegativeQuery = fmt.Sprintf(negatives[0], t)
		q.NegativeQuery = fmt.Sprintf("\"%v\"", q.NegativeQuery)
		for _, tmpl := range negatives[1:] {
			s := fmt.Sprintf(tmpl, t)
			s = fmt.Sprintf(" OR \"%v\"", s)
			q.NegativeQuery += s
		}
		queries <- &q
	}
	log.Println("foobar")
	log.Println(queries)
}
