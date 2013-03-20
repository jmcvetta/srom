// Copyright (C) 2013 Jason McVetta, all rights reserved.

package srom

import (
	"github.com/bmizerany/assert"
	"github.com/jmcvetta/randutil"
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

type DummySearchEngine struct {
	A int // Returned as hits on first call to Query()
	B int // Returned as hits on second call to Query()
	s bool
}

func (d *DummySearchEngine) ServiceName() string {
	return "Dummy search engine"
}

func (d *DummySearchEngine) Query(s string) (hits int, err error) {
	if d.s {
		d.s = false
		return d.B, nil
	}
	d.s = true
	return d.A, nil
}

func TestQuery(t *testing.T) {
	pos0, _ := randutil.IntRange(0, 999999999)
	neg0, _ := randutil.IntRange(0, 999999999)
	pos1, _ := randutil.IntRange(0, 999999999)
	neg1, _ := randutil.IntRange(0, 999999999)
	dse0 := DummySearchEngine{
		A: pos0,
		B: neg0,
	}
	dse1 := DummySearchEngine{
		A: pos1,
		B: neg1,
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
	r0 := float64(neg0) / float64(pos0)
	r1 := float64(neg1) / float64(pos1)
	expected := (r0 + r1) / 2.0
	assert.Equal(t, expected, ratio)
}
