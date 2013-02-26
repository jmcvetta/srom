// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package srom

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

var ()

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func New(terms []string, e *SearchEngine, s *Storage) *Srom {
	sr := Srom{}
	sr.Terms = terms
	sr.Storage = s
	sr.SearchEngine = e
	sr.MaxProcs = runtime.NumCPU() * 4
	sr.Positive = positiveTemplates
	sr.Negative = negativeTemplates
	return &sr
}

// Srom is a Sucks-Rules-O-Meter.
type Srom struct {
	Terms        []string      // List of terms to be evaluated
	Positive     []string      // Templates for constructing positive queries
	Negative     []string      // Templates for constructing negative queries
	Storage      *Storage      // S/R stats are written to Storage
	SearchEngine *SearchEngine // SearchEngine is used to query the internet
	MaxProcs     int           // Maximum processQuery() goroutines
}

type Storage interface {
	Write(j *job) error
}

type SearchEngine interface {
	ServiceName() string
	Query(term string, templates []string) (hits int, err error)
}

type job struct {
	Term        string
	Timestamp   time.Time
	ServiceName string
	Positive    int
	Negative    int
	Ratio       float64
}

func (sr *Srom) processQuery(queue chan *job) {
	for {
		j := <-queue
		if j == nil {
			break
		}
		j.Timestamp = time.Now()
		se := *sr.SearchEngine
		j.ServiceName = se.ServiceName()
		var err error
		j.Positive, err = se.Query(j.Term, sr.Positive)
		if err != nil {
			log.Println("Could not scrape:", err)
			continue
		}
		j.Negative, err = se.Query(j.Term, sr.Negative)
		if err != nil {
			log.Println("Could not scrape:", err)
			continue
		}
		j.Ratio = float64(j.Positive) / float64(j.Negative)
		msg := "\n"
		msg += strings.Repeat("-", 80) + "\n"
		msg += j.Term + "\n"
		msg += fmt.Sprintln("\t Positive:", j.Positive)
		msg += fmt.Sprintln("\t Negative:", j.Negative)
		msg += fmt.Sprintln("\t Ratio:", j.Ratio)
		log.Println(msg)
		sto := *sr.Storage
		sto.Write(j)
	}
}

func (sr *Srom) Run() {
	queue := make(chan *job)
	for i := 0; i < sr.MaxProcs; i++ {
		log.Println("Starting worker", i)
		go sr.processQuery(queue)
	}
	for _, t := range sr.Terms {
		j := new(job)
		j.Term = t
		log.Println("Queuing", t)
		queue <- j
	}
	// Signal workers that there is no more work.
	for i := 0; i < sr.MaxProcs; i++ {
		log.Println("Quitting worker", i)
		queue <- nil
	}
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
