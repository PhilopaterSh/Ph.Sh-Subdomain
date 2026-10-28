package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchJSONSuccess(t *testing.T) {
	var gotUA, gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotKey = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"data":[{"name":"a.example.com"}]}`)
	}))
	defer srv.Close()

	var out struct {
		Data []struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := fetchJSON(srv.Client(), srv.URL, map[string]string{"X-API-Key": "k123"}, &out); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(out.Data) != 1 || out.Data[0].Name != "a.example.com" {
		t.Fatalf("unexpected decode result: %+v", out)
	}
	if gotUA != searchEngineUserAgent {
		t.Errorf("User-Agent not attached: %q", gotUA)
	}
	if gotKey != "k123" {
		t.Errorf("custom header not attached: %q", gotKey)
	}
}

func TestFetchJSONNon200IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"error":"rate limited"}`)
	}))
	defer srv.Close()

	var out map[string]any
	if err := fetchJSON(srv.Client(), srv.URL, nil, &out); err == nil {
		t.Fatal("expected error on 403 response, got nil")
	} else if !strings.Contains(err.Error(), "status") {
		t.Errorf("error should mention the HTTP status, got: %v", err)
	}
}

func TestFetchJSONMalformedBodyIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `<html>not json</html>`)
	}))
	defer srv.Close()

	var out map[string]any
	if err := fetchJSON(srv.Client(), srv.URL, nil, &out); err == nil {
		t.Fatal("expected decode error for malformed JSON, got nil")
	}
}
