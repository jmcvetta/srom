// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package srom

import (
	"fmt"
	"launchpad.net/tomb"
	"log"
	"runtime"
	"strings"
	"time"
)

func New(e *SearchEngine, o *Output) *Srom {
	sr := Srom{}
	sr.Output = o
	sr.MaxProcs = runtime.NumCPU()
	sr.Positive = positiveTemplates
	sr.Negative = negativeTemplates
	return &sr
}

// Srom is a Sucks-Rules-O-Meter.
type Srom struct {
	queue    chan string
	Positive []string // Templates for constructing positive queries
	Negative []string // Templates for constructing negative queries
	Output   *Output  // S/R stats are written to Output
	MaxProcs int      // Maximum processQuery() goroutines
	t        *tomb.Tomb
}

// Add puts the provided term in the queue to be evaluted as sucking or rocking.
func (s *Srom) Add(term string) {
}

// A Result summarizes the result of quering a term with a given search engine.
type result struct {
	SearchEngine string
	PosCount     int
	NegCount     int
	Ratio        float64
}

type job struct {
	Term         string
	Timestamp    time.Time
	Ratio        float64
	Results      []*result
	PosTemplates []string
	NegTemplates []string
}

type query struct {
	s     string
	e     SearchEngine
	count int
}

type queryRunner struct {
	queue *chan query
	s     *Srom
	t     *tomb.Tomb
}

func (qr *queryRunner) run() {
	defer qr.t.Done()
	for {
		var q query
		select {
		case q = <-*qr.queue:
			log.Println("Querying:\n\t", q.s)
		case <-qr.t.Dying():
			close(*qr.queue)
			return
		}
		count, err := q.e.Query(q.s)
		if err != nil {
			close(*qr.queue)
			qr.t.Kill(err)
			return
		}

	}
}

/*
	j.Ratio = float64(j.Positive) / float64(j.Negative)
	msg := "\n"
	msg += strings.Repeat("-", 80) + "\n"
	msg += j.Term + "\n"
	msg += fmt.Sprintln("\t Positive:", j.Positive)
	msg += fmt.Sprintln("\t Negative:", j.Negative)
	msg += fmt.Sprintln("\t Ratio:", j.Ratio)
	log.Println(msg)
	out := *sr.Output
	out.Write(j)
*/

func (sr *Srom) Run() {
	queue := make(chan query)
	t := tomb.Tomb{}
	for i := 0; i < sr.MaxProcs; i++ {
		log.Println("Starting worker", i)
		r := queryRunner{
			s:     sr,
			queue: &queue,
			t:     &t,
		}
	}
}

/*
// Stop signals workers that there is no more work.
func (sr *Srom) Stop() {
	for i := 0; i < sr.MaxProcs; i++ {
		log.Println("Quitting worker", i)
		queue <- nil
	}
}
*/

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
