// Copyright (C) 2013 Jason McVetta, all rights reserved.

package srom

import (
	"github.com/bmizerany/assert"
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
