package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	SimpleContentRequest             = httptest.NewRequest("GET", "/?offset=0&count=5", nil)
	SimpleContentRequestWithFallback = httptest.NewRequest("GET", "/?offset=1&count=5", nil)
	OffsetContentRequest             = httptest.NewRequest("GET", "/?offset=5&count=5", nil)
	OffsetContentRequest2            = httptest.NewRequest("GET", "/?offset=10&count=100", nil)
)

func runRequest(t *testing.T, srv http.Handler, r *http.Request) (content map[int]*ContentItem) {
	response := httptest.NewRecorder()
	srv.ServeHTTP(response, r)

	if response.Code != 200 {
		t.Fatalf("Response code is %d, want 200", response.Code)
		return
	}

	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&content)
	if err != nil {
		t.Fatalf("couldn't decode Response json: %v", err)
	}

	return content
}

func TestResponseCount(t *testing.T) {
	content := runRequest(t, app, SimpleContentRequest)

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

}

func TestResponseOrder(t *testing.T) {
	content := runRequest(t, app, SimpleContentRequest)

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

	for i, item := range content {
		if Provider(item.Source) != DefaultConfig[i%len(DefaultConfig)].Type {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, DefaultConfig[i].Type,
			)
		}
	}
}

func TestOffsetResponseOrder(t *testing.T) {
	content := runRequest(t, app, OffsetContentRequest)

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

	for j, item := range content {
		i := j + 5
		if Provider(item.Source) != DefaultConfig[i%len(DefaultConfig)].Type {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, DefaultConfig[i].Type,
			)
		}
	}
}

func TestOffsetResponseOrder2(t *testing.T) {
	content := runRequest(t, app, OffsetContentRequest2)

	if len(content) != 100 {
		t.Fatalf("Got %d items back, want 100", len(content))
	}

	for j, item := range content {
		i := j + 10
		if Provider(item.Source) != DefaultConfig[i%len(DefaultConfig)].Type {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, DefaultConfig[i].Type,
			)
		}
	}
}

func runRequestWithFallback(t *testing.T, srv http.Handler, r *http.Request) (content map[int]*ContentItem) {
	response := httptest.NewRecorder()
	// force fallback on Provider3
	app.ContentClients[Provider3] = nil
	srv.ServeHTTP(response, r)

	if response.Code != 200 {
		t.Fatalf("Response code is %d, want 200", response.Code)
		return
	}

	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&content)
	if err != nil {
		t.Fatalf("couldn't decode Response json: %v", err)
	}

	return content
}

func TestResponseCountWithFallbackProvider3(t *testing.T) {
	content := runRequestWithFallback(t, app, SimpleContentRequestWithFallback)

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}
}

func TestResponseOrderWithFallbackProvider3(t *testing.T) {
	content := runRequestWithFallback(t, app, SimpleContentRequestWithFallback)

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}
	// original sequence is 1,2,3,1,1
	// config3: Provider3 fallbacks to Provider1
	// new expected sequence is 1,2,1,1,1

	if Provider(content[0].Source) != "1" {
		t.Errorf(
			"Position %d: Got Provider %v instead of Provider %v",
			0, content[0].Source, "1",
		)
	}

	if Provider(content[1].Source) != "2" {
		t.Errorf(
			"Position %d: Got Provider %v instead of Provider %v",
			1, content[1].Source, "2",
		)
	}

	if Provider(content[2].Source) != "1" {
		t.Errorf(
			"Position %d: Got Provider %v instead of Provider %v",
			2, content[2].Source, "1",
		)
	}

	if Provider(content[3].Source) != "1" {
		t.Errorf(
			"Position %d: Got Provider %v instead of Provider %v",
			3, content[3].Source, "1",
		)
	}

	if Provider(content[4].Source) != "1" {
		t.Errorf(
			"Position %d: Got Provider %v instead of Provider %v",
			4, content[4].Source, "1",
		)
	}
}

func runRequestMainFailsNoFallback(t *testing.T, srv http.Handler, r *http.Request) (content map[int]*ContentItem) {
	response := httptest.NewRecorder()
	// bring back Provider3
	app.ContentClients[Provider3] = SampleContentProvider{Source: Provider3}
	// force fallback on Provider1, this will leave config4 with no fallback
	app.ContentClients[Provider1] = nil
	srv.ServeHTTP(response, r)

	if response.Code != 200 {
		t.Fatalf("Response code is %d, want 200", response.Code)
		return
	}

	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&content)
	if err != nil {
		t.Fatalf("couldn't decode Response json: %v", err)
	}

	return content
}

func TestResponseCountMainFailsNoFallback(t *testing.T) {
	content := runRequestMainFailsNoFallback(t, app, SimpleContentRequestWithFallback)

	if len(content) != 3 {
		t.Fatalf("Got %d items back, want 3", len(content))
	}
}

func TestResponseOrderMainFailsNoFallback(t *testing.T) {
	content := runRequestMainFailsNoFallback(t, app, SimpleContentRequestWithFallback)

	if len(content) != 3 {
		t.Fatalf("Got %d items back, want 3", len(content))
	}
	// original sequence is 1,2,3,1,1
	// in config1: Provider1 fallbacks to Provider2
	// in config4: main provider fails and there is no fallback
	// respond with all the items before that point
	// new expected sequence is 2,2,3

	if Provider(content[0].Source) != "2" {
		t.Errorf(
			"Position %d: Got Provider %v instead of Provider %v",
			0, content[0].Source, "1",
		)
	}

	if Provider(content[1].Source) != "2" {
		t.Errorf(
			"Position %d: Got Provider %v instead of Provider %v",
			1, content[1].Source, "2",
		)
	}

	if Provider(content[2].Source) != "3" {
		t.Errorf(
			"Position %d: Got Provider %v instead of Provider %v",
			2, content[2].Source, "3",
		)
	}
}

func runRequestMainFailsFallbackFails(t *testing.T, srv http.Handler, r *http.Request) (content map[int]*ContentItem) {
	response := httptest.NewRecorder()
	// bring back Provider1
	app.ContentClients[Provider1] = SampleContentProvider{Source: Provider1}
	// force fallback on Provider2 and Provider3
	app.ContentClients[Provider2] = nil
	app.ContentClients[Provider3] = nil
	srv.ServeHTTP(response, r)

	if response.Code != 200 {
		t.Fatalf("Response code is %d, want 200", response.Code)
		return
	}

	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&content)
	if err != nil {
		t.Fatalf("couldn't decode Response json: %v", err)
	}

	return content
}

func TestResponseCountMainFailsFallbackFails(t *testing.T) {
	content := runRequestMainFailsFallbackFails(t, app, SimpleContentRequest)

	if len(content) != 2 {
		t.Fatalf("Got %d items back, want 2", len(content))
	}
}

func TestResponseOrderMainFailsFallbackFails(t *testing.T) {
	content := runRequestMainFailsFallbackFails(t, app, SimpleContentRequest)

	if len(content) != 2 {
		t.Fatalf("Got %d items back, want 2", len(content))
	}
	// original sequence is 1,1,2,3,1
	// in config2: Provider2 fallbacks to Provider3. Provider3 fails too.
	// respond with all the items before that point
	// new expected sequence is 1,1

	if Provider(content[0].Source) != "1" {
		t.Errorf(
			"Position %d: Got Provider %v instead of Provider %v",
			0, content[0].Source, "1",
		)
	}

	if Provider(content[1].Source) != "1" {
		t.Errorf(
			"Position %d: Got Provider %v instead of Provider %v",
			1, content[1].Source, "1",
		)
	}
}
