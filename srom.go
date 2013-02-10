// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"
	// "github.com/darkhelmet/env"
	"github.com/opesun/goquery"
	// "github.com/stathat/go"
	"runtime"
	"strings"
)

const (
	searchBase = "https://encrypted.google.com/search"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

// Srom is a Sucks-Rules-O-Meter.
type Srom struct {
	Terms    []string
	Positive []string
	Negative []string
	MaxProcs int
}

type job struct {
	Term      string
	Timestamp time.Time
	Positive  int
	Negative  int
}

func scrape(query string) (hits int, err error) {
	u, err := url.Parse(searchBase)
	if err != nil {
		return -1, err
	}
	v := url.Values{}
	v.Set("q", query)
	u.RawQuery = v.Encode()
	nodes, err := goquery.ParseUrl(u.String())
	if err != nil {
		return -1, err
	}
	resultStats := nodes.Find("#resultStats").Html()
	s := strings.Split(resultStats, " ")
	if len(s) != 3 {
		err = errors.New("Could not parse Google response")
		return -1, err
	}
	countString := s[1]
	countString = strings.Replace(countString, ",", "", -1)
	count, err := strconv.Atoi(countString)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (sr *Srom) processQuery(in, out chan *job) {
	for {
		j := <-in
		log.Println(j)
		j.Timestamp = time.Now()
		posQuery := buildQuery(j.Term, sr.Positive)
		negQuery := buildQuery(j.Term, sr.Negative)
		posCount, err := scrape(posQuery)
		if err != nil {
			log.Println("Could not scrape:", err)
			continue
		}
		negCount, err := scrape(negQuery)
		if err != nil {
			log.Println("Could not scrape:", err)
			continue
		}
		msg := strings.Repeat("-", 80) + "\n"
		msg += j.Term + "\n"
		msg += fmt.Sprintln("\t Positive:", posCount)
		msg += fmt.Sprintln("\t Negative:", negCount)
		log.Println(msg)
		out <- j
	}
	log.Println("Bye")
}

func buildQuery(term string, templates []string) string {
	q := fmt.Sprintf(templates[0], term)
	q = fmt.Sprintf("\"%v\"", q)
	for _, tmpl := range templates[1:] {
		s := fmt.Sprintf(tmpl, term)
		s = fmt.Sprintf(" OR \"%v\"", s)
		q += s
	}
	return q
}

func (sr *Srom) Run() {
	in := make(chan *job)
	out := make(chan *job)
	for i := 0; i < sr.MaxProcs; i++ {
		log.Println("Starting processQuery()", i)
		go sr.processQuery(in, out)
	}
	for _, t := range sr.Terms {
		j := new(job)
		j.Term = t
		log.Println("Queuing", t)
		in <- j
	}
	for _ = range sr.Terms {
		<-out
	}
}

func main() {
	sr := Srom{}
	sr.Terms = []string{
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
	sr.Positive = []string{
		"%v rules",
		"%v rocks",
		"%v is awesome",
		"%v kicks ass",
		"%v dominates",
		"love %v",
	}
	sr.Negative = []string{
		"%v sucks",
		"%v blows",
		"%v is worthless",
		"%v is crap",
		"%v doesn't work",
		"hate %v",
	}
	sr.MaxProcs = runtime.NumCPU() * 4
	sr.Run()
}
