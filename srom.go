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

// Add puts the provided term in the queue to be evaluted as sucking or rocking.
func (s *Srom) Add(term string) {
	log.Println("Add", term)
	j := job{
		Term: term,
	}
	s.queue <- &j
	log.Println("Added", term)
}

func (sr *Srom) Run() {
	log.Println("Start Run()")
	jr := jobRunner{
		s: sr,
	}
	for i := 0; i < sr.MaxRunners; i++ {
		go jr.run()
	}
	log.Println("End Run()")
}

type jobRunner struct {
	s *Srom
	t tomb.Tomb
}

func (jr *jobRunner) run() {
	log.Println("Starting job runner")
	defer jr.t.Done()
	for {
		log.Println("For")
		var j *job
		select {
		case j = <-jr.s.queue:
			jr.processJob(j)
		case <-jr.t.Dying():
			log.Println("Exiting job runner")
			close(jr.s.queue)
			return
		}
	}
}

func (jr *jobRunner) processJob(j *job) {
	log.Println("Processing job", *j)
	j.Timestamp = time.Now()
	j.PosTemplates = jr.s.Positive
	j.NegTemplates = jr.s.Negative
	//
	// Issue search engine queries async
	//
	queryCnt := len(jr.s.engines) * 2 // 1 pos and 1 neg query per search engine
	wg := sync.WaitGroup{}
	wg.Add(queryCnt)
	for _, se := range jr.s.engines {
		var pos int
		var neg int
		r := result{
			SearchEngine: se.ServiceName(),
			PosCount:     pos,
			NegCount:     neg,
		}
		j.Results = append(j.Results, &r)
		// Positive
		q := buildQuery(j.Term, jr.s.Positive)
		go jr.runQuery(q, se, &pos, &wg)
		// Negative
		q = buildQuery(j.Term, jr.s.Negative)
		go jr.runQuery(q, se, &neg, &wg)
	}
	wg.Wait()
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
	log.Println("Output", *j)
	err := jr.s.Output.Write(j)
	if err != nil {
		close(jr.s.queue)
		jr.t.Kill(err)
		return
	}
}

func (jr *jobRunner) runQuery(q string, se SearchEngine, result *int, wg *sync.WaitGroup) {
	log.Println("runQuery", se.ServiceName(), q)
	defer wg.Done()
	count, err := se.Query(q)
	if err != nil {
		close(jr.s.queue)
		jr.t.Kill(err)
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
