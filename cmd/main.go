// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package main

import (
	"flag"
	"github.com/darkhelmet/env"
	"github.com/jmcvetta/srom"
	"labix.org/v2/mgo"
	"log"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func setupMongoStorage() *srom.MongoStorage {
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
	mongo := srom.MongoStorage{
		Col: collection,
	}
	return &mongo
}

func main() {
	//
	// Parse Flags
	//
	// useBing := flag.Bool("bing", false, "Use Bing search instead of Google")
	flag.Parse()
	//
	// Search engines
	//
	google := &srom.GoogleSearch{
		ApiKey:         env.String("GOOGLE_API_KEY"),
		CustomSearchId: env.String("GOOGLE_CUSTOM_SEARCH_ID"),
	}
	bing := &srom.BingSearch{
		CustomerId: env.String("AZURE_CUST_ID"),
		Key:        env.String("AZURE_KEY"),
	}
	engines := []srom.SearchEngine{
		google,
		bing,
	}
	//
	// Instantiate Srom Object
	//
	/*
		terms := []string{
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
	*/
	sr := srom.New(engines, &srom.LoggerOutput{})
	//
	// Run
	//
	sr.Query("ubuntu")
}
