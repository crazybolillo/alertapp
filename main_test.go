package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGetIndexHandler(t *testing.T) {
	store := InMemoryAlertStore{}
	now := time.Now()

	store.Storage = append(store.Storage, Alert{
		UUID: uuid.NewString(),
		Time: &now,
		Info: "Spooky",
	}, Alert{
		UUID: uuid.NewString(),
		Time: &now,
		Info: "Help!!",
	})

	tests := []struct {
		url         string
		expectHttp  int
		expectCount string
	}{
		{"/", http.StatusOK, "2"},
		{"/?page=0&size=10", http.StatusOK, "2"},
		{"/?page=5&size=20", http.StatusOK, "0"},
		{"/?page=-1&size=15", http.StatusBadRequest, ""},
		{"/?page=1&size=-1", http.StatusBadRequest, ""},
		{"/?page=0&size=0", http.StatusOK, "0"},
	}

	for _, test := range tests {
		req := httptest.NewRequest(http.MethodGet, test.url, nil)
		w := httptest.NewRecorder()

		indexHandler(&store)(w, req)
		res := w.Result()

		if res.StatusCode != test.expectHttp {
			t.Errorf("%q = HTTP %d, want HTTP %d", test.url, res.StatusCode, test.expectHttp)
		}
		if count := res.Header.Get(AlertCountHttpHeader); count != test.expectCount {
			t.Errorf("%s = %s: %s, want %s: %s", test.url, AlertCountHttpHeader, count, AlertCountHttpHeader, test.expectCount)
		}
	}
}

func TestPostIndexHandler(t *testing.T) {
	store := InMemoryAlertStore{}
	testData := []Alert{
		{Info: "Proximity sensor activated"},
		{Info: "Detected noise"},
		{Info: "May be a ghost"},
	}

	count := 1
	for _, data := range testData {
		marshal, err := json.Marshal(data)
		if err != nil {
			t.Errorf("error while marshaling: %s", err)
			continue
		}

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(marshal))
		w := httptest.NewRecorder()

		indexHandler(&store)(w, req)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Errorf("%s / = HTTP %d, wanted HTTP %d", http.MethodPost, http.StatusOK, res.StatusCode)
		}

		if len(store.Storage) != count {
			t.Errorf("Alert not stored, expected %d, got %d", count, len(store.Storage))
		}
		count++
	}
}

func TestPostIndexHandlerBadData(t *testing.T) {
	store := InMemoryAlertStore{}
	marshal, err := json.Marshal(`{"description": "highly illegal packet"}`)
	if err != nil {
		t.Fatalf("error while marshaling: %s", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(marshal))
	w := httptest.NewRecorder()

	indexHandler(&store)(w, req)
	res := w.Result()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Bad payload = %d, wanted %d", http.StatusBadRequest, res.StatusCode)
	}
}

func TestIndexHandlerNotFound(t *testing.T) {
	store := InMemoryAlertStore{}
	req := httptest.NewRequest(http.MethodGet, "/magical/path", nil)
	w := httptest.NewRecorder()

	indexHandler(&store)(w, req)
	res := w.Result()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("HTTP GET /magical/path = %d, wanted %d", http.StatusNotFound, res.StatusCode)
	}
}

func TestPopulate(t *testing.T) {
	min, max := 5, 10
	for idx := 0; idx < 5; idx++ {
		store := InMemoryAlertStore{}
		populate(min, max, &store)

		size := len(store.Storage)
		if size < min {
			t.Errorf("Wanted at least %d entries, got %d", min, size)
		}
		if size > max {
			t.Errorf("Wanted at most %d entries, got %d", max, size)
		}
	}
}
