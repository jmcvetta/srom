// Copyright (C) 2013 Jason McVetta, all rights reserved.

package srom

import (
	"github.com/bmizerany/assert"
	"github.com/jmcvetta/randutil"
	"strings"
	"testing"
)

func TestBuildQuery(t *testing.T) {
	templates := []string{
		"%v rules",
		"%v rocks",
	}
	term := "foobar"
	s := buildQuery(term, templates)
	expected := "\"foobar rules\" OR \"foobar rocks\""
	assert.Equal(t, expected, s)
}

type dummySearchEngine struct {
	Sucks int
	Rules int
}

func (d *dummySearchEngine) ServiceName() string {
	return "Dummy search engine"
}

func (d *dummySearchEngine) Query(s string) (hits int, err error) {
	if strings.Contains(s, "rules") {
		return d.Rules, nil
	}
	return d.Sucks, nil
}

func TestQuery(t *testing.T) {
	rules0, _ := randutil.IntRange(0, 999999)
	sucks0, _ := randutil.IntRange(0, 999999)
	rules1, _ := randutil.IntRange(0, 999999)
	sucks1, _ := randutil.IntRange(0, 999999)
	dse0 := dummySearchEngine{
		Rules: int(rules0),
		Sucks: int(sucks0),
	}
	dse1 := dummySearchEngine{
		Rules: int(rules1),
		Sucks: int(sucks1),
	}
	engines := []SearchEngine{
		&dse0,
		&dse1,
	}
	sr := New(engines, &NilOutput{})
	sr.Run()
	ratio, err := sr.Query("foobar")
	if err != nil {
		t.Fatal(err)
	}
	r0 := float64(sucks0) / float64(rules0)
	r1 := float64(sucks1) / float64(rules1)
	expected := (r0 + r1) / 2.0
	assert.Equal(t, expected, ratio)
}
