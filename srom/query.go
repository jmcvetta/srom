// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the AGPL v3.  http://www.gnu.org/licenses/agpl-3.0.html

package srom

import (
	"errors"
	"fmt"
	"github.com/jmcvetta/restclient"
	"net/url"
	"strconv"
)

const (
	googleSearchApi = "https://www.googleapis.com/customsearch/v1"
)

type GoogleSearch struct {
	ApiKey         string
	CustomSearchId string
	Client         *restclient.Client
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
	status, err := gs.Client.Do(&req)
	if err != nil {
		return -1, err
	}
	if status != 200 {
		err = errors.New(fmt.Sprintf("Bad response code from Google: %v", status))
		return -1, err
	}
	if len(resp.Queries.Request) < 1 {
		err = errors.New("Could not parse JSON response from Google.")
		return -1, err
	}
	count, err := strconv.Atoi(resp.Queries.Request[0].TotalResults)
	if err != nil {
		return -1, err
	}
	return count, nil
}
