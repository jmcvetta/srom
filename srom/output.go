// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package srom

import (
	"labix.org/v2/mgo"
	"log"
)

type Output interface {
	Write(j *job) error
}

// MongoStorage allows SROM data to be persisted in a MongoDB collection.
type MongoStorage struct {
	Col *mgo.Collection
}

func (m *MongoStorage) Write(j *job) error {
	return m.Col.Insert(&j)
}

type LoggerOutput struct{}

func (l *LoggerOutput) Write(j *job) error {
	log.Println(j.Term, j.Ratio)
	return nil
}

type NilOutput struct{}

// No-op
func (n *NilOutput) Write(j *job) error {
	return nil
}
