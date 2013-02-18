// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package srom

import (
	"strings"
	"log"
	"time"
	"fmt"
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
	SearchEngine   SearchEngine // SearchEngine is used to query the internet
	MaxProcs       int      // Maximum processQuery() goroutines
}

type Storage interface {
	Write(j *job) error
}


type SearchEngine interface {
	Query(term string, templates []string) (hits int, err error)
}

type job struct {
	Term      string
	Timestamp time.Time
	Positive  int
	Negative  int
	Ratio     float64
}

func (sr *Srom) processQuery(in, out chan *job) {
	for {
		j := <-in
		j.Timestamp = time.Now()
		var err error
		j.Positive, err = sr.SearchEngine.Query(j.Term, sr.Positive)
		if err != nil {
			log.Println("Could not scrape:", err)
			out <- j
		}
		j.Negative, err = sr.SearchEngine.Query(j.Term, sr.Negative)
		if err != nil {
			log.Println("Could not scrape:", err)
			out <- j
		}
		j.Ratio = float64(j.Positive) / float64(j.Negative)
		msg := "\n"
		msg += strings.Repeat("-", 80) + "\n"
		msg += j.Term + "\n"
		msg += fmt.Sprintln("\t Positive:", j.Positive)
		msg += fmt.Sprintln("\t Negative:", j.Negative)
		msg += fmt.Sprintln("\t Ratio:", j.Ratio)
		log.Println(msg)
		sr.Storage.Write(j)
		out <- j
	}
	log.Println("Bye")
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
