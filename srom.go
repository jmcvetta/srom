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
	sr.jobs = make(chan *job)
	sr.queries = make(chan *query)
	sr.Output = o
	sr.Positive = positiveTemplates
	sr.Negative = negativeTemplates
	sr.engines = engines
	sr.MaxRunners = runtime.NumCPU()
	sr.MaxRunners = 1
	return &sr
}

// Srom is a Sucks-Rules-O-Meter.
type Srom struct {
	jobs       chan *job
	queries    chan *query
	engines    []SearchEngine
	queryCnt   int      // Number of queries that will be run per
	Positive   []string // Templates for constructing positive queries
	Negative   []string // Templates for constructing negative queries
	Output     Output   // S/R stats are written to Output
	MaxRunners int      // Maximum count of jobRunners
	t          tomb.Tomb
}

func (sr *Srom) Run() {
	qr := queryRunner{
		sr: sr,
	}
	for i := 0; i < sr.MaxRunners; i++ {
		go qr.run(i)
	}
}

func (sr *Srom) queueJob(j *job) {
	log.Println("queueJob", &j)
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
}

func (sr *Srom) Query(term string) error {
	log.Println("Querying term", term)
	j := job{
		Term: term,
	}
	sr.queueJob(&j)
	j.wg.Wait()
	log.Println("done waiting")
	for _, r := range j.Results {
		log.Println(r.Positive)
		log.Println(r.Positive.hits)
		log.Println(r.Negative)
		log.Println(r.Negative.hits)
	}
	//
	// Calculate Ratio
	//
	var sum float64
	for _, r := range j.Results {
		// pos <- r.Positive.hits
		ratio := float64(<-r.Positive.hits) / float64(<-r.Negative.hits)
		sum += ratio
	}
	j.Ratio = sum / float64(len(j.Results))
	//
	// Write to output
	//
	log.Println("Output", j)
	err := sr.Output.Write(&j)
	return err
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
	log.Println("Starting queryRunner", id)
	for {
		var q *query
		select {
		case <-qr.t.Dying():
			log.Println("queryRunner", id, "dying")
			close(qr.sr.queries)
			return
		case q = <-qr.sr.queries:
			log.Println(q)
		}
		hits, err := q.se.Query(q.q)
		if err != nil {
			log.Println("Query failed:", err)
			qr.t.Kill(err)
			return
		}
		log.Println(hits)
		q.wg.Done()
		q.hits <- hits
		log.Println("query runner done")
	}
}

// A Result summarizes the result of quering a term with a given search engine.
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
