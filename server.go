package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// App represents the server's internal state.
// It holds configuration about providers and content
type App struct {
	ContentClients map[Provider]Client
	Config         ContentMix
}

// Response struct with index to keep items in predetermined order
type Response struct {
	index int
	item  *ContentItem
}

func (App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.String())

	// prepare an unbuffered channel
	ch := make(chan *Response)

	// get count
	count, err := parseRequest("count", req)
	if err != nil {
		log.Printf("error parsing request for count: %v", err)
	}

	// get offset
	offset, err := parseRequest("offset", req)
	if err != nil {
		log.Printf("error parsing request for offset: %v", err)
	}

	// set the list of Providers (either main or fallback) to use
	var providerListToUse []ContentConfig
	for i := offset; i < offset+count; i++ {
		id := int(i) % len(app.Config)
		provider := app.Config[id]
		providerListToUse = append(providerListToUse, provider)
	}

	// get correct map of Items
	result := makeRequests(providerListToUse, ch)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		panic(err)
	}
}

// parse request for count and offset
func parseRequest(toParse string, r *http.Request) (int64, error) {
	s := r.URL.Query().Get(toParse)
	res, err := strconv.ParseInt(s, 10, 64)
	return res, err
}

// make requests both with main providers and fallback providers if needed
func makeRequests(providerListToUse []ContentConfig, ch chan *Response) map[int]*ContentItem {
	mapItems := make(map[int]*ContentItem)
	var fallBackProviders []*Provider

	// make main requests
	for i, providers := range providerListToUse {
		mainProvider := providers.Type
		fallBackProviders = append(fallBackProviders, providers.Fallback)
		client := app.ContentClients[mainProvider]
		mainIndex := i
		go request(mainIndex, mainProvider, client, ch)
		matchResToProvider := <-ch
		mapItems[matchResToProvider.index] = matchResToProvider.item
	}

	// make fallback requests
	for key, value := range mapItems {
		if value == nil {
			fallBackIndex := key
			if fallBackProviders[fallBackIndex] != nil {
				client := app.ContentClients[*fallBackProviders[fallBackIndex]]
				go request(fallBackIndex, *fallBackProviders[fallBackIndex], client, ch)
				matchResToProvider := <-ch
				mapItems[matchResToProvider.index] = matchResToProvider.item
			}
		}
	}

	return processedForDoubleFailure(mapItems)
}

// request content from Provider
func request(index int, provider Provider, client Client, ch chan *Response) {
	if client == nil {
		ch <- &Response{index: index, item: nil}
		return
	}
	res, err := client.GetContent(*addr, 1)
	if err != nil {
		log.Printf("getting item from provider '%s' failed", provider)
		return
	}
	ch <- &Response{index: index, item: res[0]}
}

// processedForDoubleFailure checks if both the main provider and the fallback fail
// (or if the main provider fails and there is no fallback), the API should respond with all
// the items before that point.
func processedForDoubleFailure(mapItems map[int]*ContentItem) map[int]*ContentItem {
	// find the first failure
	var firstFailure int
	for k := len(mapItems); k >= 0; k-- {
		if mapItems[k] == nil {
			firstFailure = k
		}
	}
	// keep items before the first failure
	for n := len(mapItems); n >= firstFailure; n-- {
		delete(mapItems, n)
	}
	return mapItems
}
