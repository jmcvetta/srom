// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package main

import (
	"errors"
	"fmt"
	"github.com/darkhelmet/env"
	"github.com/jmcvetta/restclient"
	"github.com/stathat/go"
	"log"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	apiBase = "https://www.googleapis.com/customsearch/v1"
)

var ()

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

// Srom is a Sucks-Rules-O-Meter.
type Srom struct {
	Terms          []string // List of terms to be evaluated
	Positive       []string // Templates for constructing positive queries
	Negative       []string // Templates for constructing negative queries
	Storage        Storage  // S/R stats are written to Storage
	MaxProcs       int      // Maximum processQuery() goroutines
	apiKey         string   // Google API key
	customSearchId string   // Google Custom Search identifier
	client         *restclient.Client
}

type Storage interface {
	Write(term string, pos, neg int) error
}

type statHatLogger struct {
	ezkey string
}

func (s *statHatLogger) Write(term string, pos, neg int) error {
	ratio := float64(pos) / float64(neg)
	name := "SROM -- " + term
	log.Println("Logging to StatHat:", term, pos, neg, ratio)
	return stathat.PostEZValue(name, s.ezkey, ratio)
	/*
		err := stathat.PostEZValue(base+"pos", s.ezkey, float64(pos))
		if err != nil {
			return err
		}
		err = stathat.PostEZValue(base+"neg", s.ezkey, float64(neg))
		if err != nil {
			return err
		}
	*/
}

type job struct {
	Term      string
	Timestamp time.Time
	Positive  int
	Negative  int
}

func (sr *Srom) google(query string) (hits int, err error) {
	u, err := url.Parse(apiBase)
	if err != nil {
		return -1, err
	}
	v := url.Values{}
	v.Set("key", sr.apiKey)
	v.Set("cx", sr.customSearchId)
	v.Set("q", query)
	u.RawQuery = v.Encode()
	resp := struct {
		Queries struct {
			Request []struct {
				TotalResults string `json:"totalResults"`
			} `json:"requests"`
		} `json:"queries"`
	}{}
	e := new(interface{})
	req := restclient.RestRequest{
		Url:    u.String(),
		Method: restclient.GET,
		Result: &resp,
		Error:  e,
	}
	status, err := sr.client.Do(&req)
	if err != nil {
		return -1, err
	}
	if status != 200 {
		err = errors.New(fmt.Sprintf("Bad response code from Google: %v", status))
		return -1, err
	}
	if len(resp.Queries.Request) < 1 {
		err = errors.New("Could not parse JSON response from Google.")
		return -1, err
	}
	count, err := strconv.Atoi(resp.Queries.Request[0].TotalResults)
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
		log.Println(posQuery)
		posCount, err := sr.google(posQuery)
		if err != nil {
			log.Println("Could not scrape:", err)
			out <- j
		}
		negCount, err := sr.google(negQuery)
		if err != nil {
			log.Println("Could not scrape:", err)
			out <- j
		}
		msg := strings.Repeat("-", 80) + "\n"
		msg += j.Term + "\n"
		msg += fmt.Sprintln("\t Positive:", posCount)
		msg += fmt.Sprintln("\t Negative:", negCount)
		log.Println(msg)
		sr.Storage.Write(j.Term, posCount, negCount)
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
	sr.Storage = &statHatLogger{
		ezkey: env.StringDefault("STATHAT_EZKEY", "jason.mcvetta@gmail.com"),
	}
	sr.MaxProcs = runtime.NumCPU() * 4
	sr.apiKey = env.String("GOOGLE_API_KEY")
	sr.customSearchId = env.String("GOOGLE_CUSTOM_SEARCH_ID")
	sr.client = restclient.New()
	sr.Run()
	stathat.WaitUntilFinished(time.Second * 60)

}
