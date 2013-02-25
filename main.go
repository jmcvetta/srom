// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package main

import (
	"github.com/darkhelmet/env"
	"github.com/jmcvetta/srom/srom"
	"labix.org/v2/mgo"
	"log"
	"runtime"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func main() {
	//
	// Setup MongoDB
	//
	mongoUrl := env.StringDefault("MONGOLAB_URI", "localhost")
	log.Println("Connecting to MongoDB on " + mongoUrl + "...")
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()
	db := session.DB("")
	_, err = db.CollectionNames()
	if err != nil {
		log.Println("Setting db name to 'SROM'.")
		db = session.DB("SROM")
	}
	collection := db.C("terms")
	termIdx := mgo.Index{
		Key:        []string{"Term", "Timestamp"},
		Unique:     false,
		Background: true,
	}
	log.Println("Ensuring Term, Timestamp index")
	err = collection.EnsureIndex(termIdx)
	if err != nil {
		log.Panic(err)
	}
	ratioIdx := mgo.Index{
		Key:        []string{"Ratio", "Timestamp"},
		Unique:     false,
		Background: true,
	}
	log.Println("Ensuring Ratio, Timestamp index")
	err = collection.EnsureIndex(ratioIdx)
	if err != nil {
		log.Panic(err)
	}
	//
	// Instantiate Srom Object
	//
	sr := srom.Srom{}
	sr.Terms = []string{
		"ubuntu",
		"obama",
		"linux",
		"windows",
		"apple",
		"iPhone",
		"android",
		"iOS",
		"ed lee",
		"kate moss",
		"richard stallman",
	}
	sr.Positive = []string{
		"%v rules",
		"%v rocks",
		"%v is awesome",
		"%v kicks ass",
		"%v dominates",
		"love %v",
	}
	sr.Negative = []string{
		"%v sucks",
		"%v blows",
		"%v is worthless",
		"%v is crap",
		"%v doesn't work",
		"hate %v",
	}
	sr.Storage = &srom.MongoStorage{
		Col: collection,
	}
	ncpu := runtime.NumCPU() * 4
	runtime.GOMAXPROCS(ncpu)
	sr.MaxProcs = ncpu
	//
	// Search engines
	//
	/*
		sr.SearchEngine = &srom.GoogleSearch{
			ApiKey:         env.String("GOOGLE_API_KEY"),
			CustomSearchId: env.String("GOOGLE_CUSTOM_SEARCH_ID"),
		}
	*/
	sr.SearchEngine = &srom.BingSearch{
		CustomerId: env.String("AZURE_CUST_ID"),
		Key:        env.String("AZURE_KEY"),
	}
	// sr.ApiKey = env.String("GOOGLE_API_KEY")
	// sr.CustomSearchId = env.String("GOOGLE_CUSTOM_SEARCH_ID")
	// sr.Client = restclient.New()
	//
	// Run
	//
	log.Println("Running SROM")
	sr.Run()
}
