// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package srom

import (
	"labix.org/v2/mgo"
)

// MongoStorage allows SROM data to be persisted in a MongoDB collection.
type MongoStorage struct {
	Col *mgo.Collection
}

func (m *MongoStorage) Write(j *job) error {
	return m.Col.Insert(&j)
}
