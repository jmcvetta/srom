// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package srom

import (
	"errors"
	"fmt"
	"github.com/jmcvetta/restclient"
	"log"
	"net/url"
	"strconv"
)

const (
	googleSearchApi = "https://www.googleapis.com/customsearch/v1"
	bingSearchApi   = "https://api.datamarket.azure.com/Bing/Search/v1/Composite"
)

var (
	BadResponse = errors.New("Bad response code from server.")
)

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

type GoogleSearch struct {
	ApiKey         string
	CustomSearchId string
}

func (gs *GoogleSearch) Query(term string, templates []string) (hits int, err error) {
	query := buildQuery(term, templates)
	u, err := url.Parse(googleSearchApi)
	if err != nil {
		return -1, err
	}
	v := url.Values{}
	v.Set("key", gs.ApiKey)
	v.Set("cx", gs.CustomSearchId)
	v.Set("q", query)
	u.RawQuery = v.Encode()
	resp := struct {
		Queries struct {
			Request []struct {
				TotalResults string `json:"totalResults"`
			} `json:"requests"`
		} `json:"queries"`
	}{}
	e := new(interface{})
	req := restclient.RestRequest{
		Url:    u.String(),
		Method: restclient.GET,
		Result: &resp,
		Error:  e,
	}
	status, err := restclient.Do(&req)
	if err != nil {
		return -1, err
	}
	if status != 200 {
		log.Printf("Bad response code from Google: %v", status)
		return -1, BadResponse
	}
	if len(resp.Queries.Request) < 1 {
		err = errors.New("Could not parse JSON response from Google.")
		return -1, BadResponse
	}
	count, err := strconv.Atoi(resp.Queries.Request[0].TotalResults)
	if err != nil {
		return -1, err
	}
	return count, nil
}

type BingSearch struct {
	CustomerId string
	Key        string
}

func (bs *BingSearch) Query(term string, templates []string) (hits int, err error) {
	query := buildQuery(term, templates)
	query = fmt.Sprintf("'%v'", query) // Enclose in single quote marks
	payload := map[string]string{
		"Sources": "'web'", // Inner single quote marks are required
		"$format": "json",  // Yes the $ prefix is correct
		"Query":   query,
	}
	resp := struct {
		D struct {
			Results []struct {
				Total string `json:"WebTotal"`
			} `json:"results"`
		} `json:"d"`
	}{}
	req := restclient.RestRequest{
		Url:      bingSearchApi,
		Method:   restclient.GET,
		Userinfo: url.UserPassword(bs.CustomerId, bs.Key),
		Params:   payload,
		Result:   resp,
	}
	// r.json['d']['results'][0]['WebTotal']
	status, err := restclient.Do(&req)
	if err != nil {
		return -1, err
	}
	if status != 200 {
		log.Printf("Bad response code from Bing: %v", status)
		return -1, BadResponse
	}
	if len(resp.D.Results) != 1 {
		log.Printf("Expected single item in results dict, but got", len(resp.D.Results))
		return -1, BadResponse
	}
	hits, err = strconv.Atoi(resp.D.Results[0].Total)
	if err != nil {
		return -1, err
	}
	return hits, nil
}
