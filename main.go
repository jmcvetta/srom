// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

// A command line Sucks-Rules-O-Meter
package main

import (
	"flag"
	"fmt"
	"github.com/darkhelmet/env"
	"github.com/jmcvetta/srom/srom"
	"github.com/msbranco/goconfig"
	"labix.org/v2/mgo"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

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

func getString(c *goconfig.ConfigFile, section, option string) string {
	value, err := c.GetString(section, option)
	if err != nil {
		log.Println(section, ":", option)
		log.Fatal(err)
	}
	if value == "" {
		log.Fatalf("No value for option '%v' in section '%v'.", option, section)
	}
	return value
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile) // Ought to be controlled with a flag
	//
	// Config File
	//
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	confFolder := filepath.Join(u.HomeDir, ".config", "srom")
	confFile := filepath.Join(confFolder, "srom.conf")
	_, err = os.Stat(confFile)
	if os.IsNotExist(err) {
		os.MkdirAll(confFolder, 0700)
		c := goconfig.NewConfigFile()
		c.AddSection("google")
		c.AddOption("google", "custom_search_id", "")
		c.AddOption("google", "api_key", "")
		c.AddSection("azure")
		c.AddOption("azure", "customer_id", "")
		c.AddOption("azure", "key", "")
		header := "Sucks-Rules-O-Meter Configuration"
		err = c.WriteConfigFile(confFile, 0600, header)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Created new config file at", confFile)
		log.Println("You will need to populate it with your API credentials")
		return
	}
	engines := []srom.SearchEngine{}
	c, err := goconfig.ReadConfigFile(confFile)
	if err != nil {
		log.Fatal(err)
	}
	for _, section := range c.GetSections() {
		//
		// Google
		//
		if section == "google" {
			google := srom.GoogleSearch{
				ApiKey:         getString(c, "google", "api_key"),
				CustomSearchId: getString(c, "google", "custom_search_id"),
			}
			engines = append(engines, &google)
		}
		if section == "azure" {
			bing := srom.BingSearch{
				CustomerId: getString(c, "azure", "customer_id"),
				Key:        getString(c, "azure", "key"),
			}
			engines = append(engines, &bing)
		}
	}
	//
	// Parse Flags
	//
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("Must supply one and only one argument")
	}
	//
	// Run Query
	//
	sr := srom.New(engines, &srom.NilOutput{})
	sr.Run()
	term := flag.Arg(0)
	ratio, err := sr.Query(term)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nThe term '%v' has a sucks/rules ratio of %v\n", term, ratio)
	msg := fmt.Sprintf("The internet thinks '%v' ", term)
	switch {
	case ratio > 1:
		msg += "SUCKS"
	case ratio == 1:
		msg += "neither sucks nor rocks.  Weird, eh?"
	case ratio < 1:
		msg += "ROCKS"
	}
	if ratio > 2 || ratio < 0.5 {
		msg += " HARD"
	}
	msg += ".\n\n"
	fmt.Printf(msg)

}
