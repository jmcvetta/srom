// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

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
	sr.queue = make(chan *job)
	sr.Output = o
	sr.Positive = positiveTemplates
	sr.Negative = negativeTemplates
	sr.engines = engines
	queryCnt := len(sr.engines) * 2 // 1 pos and 1 neg query per search engine
	mr := runtime.NumCPU() / queryCnt
	if mr == 0 {
		mr = 1 // Minimum 1 runner required
	}
	sr.MaxRunners = mr
	return &sr
}

// Srom is a Sucks-Rules-O-Meter.
type Srom struct {
	queue      chan *job
	engines    []SearchEngine
	queryCnt   int      // Number of queries that will be run per
	Positive   []string // Templates for constructing positive queries
	Negative   []string // Templates for constructing negative queries
	Output     Output   // S/R stats are written to Output
	MaxRunners int      // Maximum count of jobRunners
	t          tomb.Tomb
}

func (sr *Srom) Query(term string) error {
	log.Println("Querying term", term)
	j := job{
		Term: term,
		Timestamp: time.Now(),
		PosTemplates: sr.Positive,
		NegTemplates: sr.Negative,
	}
	//
	// Issue search engine queries async
	//
	qr := queryRunner{}
	for _, se := range sr.engines {
		var pos int
		var neg int
		r := result{
			SearchEngine: se.ServiceName(),
			PosCount:     pos,
			NegCount:     neg,
		}
		j.Results = append(j.Results, &r)
		qr.wg.Add(2) // 1 pos and 1 neg query per search engine
		// Positive
		q := buildQuery(j.Term, sr.Positive)
		go qr.runQuery(q, se, &pos)
		// Negative
		q = buildQuery(j.Term, sr.Negative)
		go qr.runQuery(q, se, &neg)
	}
	go func() {
		qr.wg.Wait()
		qr.t.Done()
	}()
	err := qr.t.Wait()
	if err != nil {
		return err
	}
	//
	// Calculate Ratio
	//
	var sum float64
	for _, r := range j.Results {
		ratio := float64(r.PosCount) / float64(r.NegCount)
		sum += ratio
	}
	j.Ratio = sum / float64(len(j.Results))
	//
	// Write to output
	//
	log.Println("Output", j)
	err = sr.Output.Write(&j)
	return err
}

type queryRunner struct {
	t tomb.Tomb
	wg sync.WaitGroup
}

func (qr *queryRunner) runQuery(q string, se SearchEngine, result *int) {
	log.Println("runQuery", se.ServiceName(), q)
	defer qr.wg.Done()
	count, err := se.Query(q)
	if err != nil {
		log.Println("Query failed:", err)
		qr.t.Kill(err)
		return
	}
	result = &count
}

// A Result summarizes the result of quering a term with a given search engine.
type result struct {
	SearchEngine string
	PosCount     int
	NegCount     int
}

type job struct {
	Term         string
	Timestamp    time.Time
	Ratio        float64
	Results      []*result
	PosTemplates []string
	NegTemplates []string
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
