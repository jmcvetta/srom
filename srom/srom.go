// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

// Package srom is a Sucks-Rules-O-Meter
package srom

import (
	"fmt"
	"launchpad.net/tomb"
	"log"
	"runtime"
	"sync"
	"time"
)

func New(engines []SearchEngine, o Output) *Srom {
	sr := Srom{}
	sr.queries = make(chan *query)
	sr.Output = o
	sr.Positive = positiveTemplates
	sr.Negative = negativeTemplates
	sr.engines = engines
	sr.MaxRunners = runtime.NumCPU()
	// sr.MaxRunners = 1
	return &sr
}

// Srom is a Sucks-Rules-O-Meter.
type Srom struct {
	queries    chan *query
	engines    []SearchEngine
	queryCnt   int      // Number of queries that will be run per
	Positive   []string // Templates for constructing positive queries
	Negative   []string // Templates for constructing negative queries
	Output     Output   // S/R stats are written to Output
	MaxRunners int      // Maximum count of jobRunners
	t          tomb.Tomb
}

// Run starts up a pool of query runners.
func (sr *Srom) Run() {
	qr := queryRunner{
		sr: sr,
	}
	for i := 0; i < sr.MaxRunners; i++ {
		go qr.run(i)
	}
}

// Query asks the internet for its opinion of a given term, returning the ratio
// of favorable to unfavorable hits.
func (sr *Srom) Query(term string) (ratio float64, err error) {
	j := job{
		Term: term,
	}
	j.Timestamp = time.Now()
	j.PosTemplates = sr.Positive
	j.NegTemplates = sr.Negative
	for _, se := range sr.engines {
		posQ := query{
			wg:   &j.wg,
			se:   se,
			q:    buildQuery(j.Term, sr.Positive),
			hits: make(chan int, 1),
		}
		negQ := query{
			wg:   &j.wg,
			se:   se,
			q:    buildQuery(j.Term, sr.Negative),
			hits: make(chan int, 1),
		}
		r := result{
			SearchEngine: se.ServiceName(),
			Positive:     posQ,
			Negative:     negQ,
		}
		j.Results = append(j.Results, &r)
		j.wg.Add(2) // 1 pos and 1 neg query per search engine
		sr.queries <- &posQ
		sr.queries <- &negQ
	}
	j.wg.Wait()
	//
	// Calculate Ratio
	//
	var sum float64
	var cnt float64
	for _, r := range j.Results {
		pos := <-r.Positive.hits
		neg := <-r.Negative.hits
		ratio := float64(pos) / float64(neg)
		log.Printf("%30v %10v %5v %10f\n", r.SearchEngine, pos, neg, ratio)
		if pos == -1 || neg == -1 {
			continue // One or both queries failed
		}
		sum += ratio
		cnt += 1.0
	}
	j.Ratio = sum / cnt
	//
	// Write to output
	//
	err = sr.Output.Write(&j)
	if err != nil {
		return 0, err
	}
	return j.Ratio, nil
}

type query struct {
	q    string
	hits chan int
	se   SearchEngine
	wg   *sync.WaitGroup
}

type queryRunner struct {
	t  tomb.Tomb
	sr *Srom
}

func (qr *queryRunner) run(id int) {
	for {
		var q *query
		select {
		case <-qr.t.Dying():
			close(qr.sr.queries)
			return
		case q = <-qr.sr.queries:
		}
		hits, err := q.se.Query(q.q)
		if err != nil {
			// Query failed
			hits = -1
		}
		q.hits <- hits
		q.wg.Done()
	}
}

// A Result encapsulates the results of querying a term
type result struct {
	SearchEngine string
	Positive     query
	Negative     query
}

type job struct {
	Term         string
	Timestamp    time.Time
	Ratio        float64
	Results      []*result
	PosTemplates []string
	NegTemplates []string
	wg           sync.WaitGroup
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
