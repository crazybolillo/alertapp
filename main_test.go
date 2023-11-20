package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIndexHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	indexHandler(w, req)
	res := w.Result()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := "Hello World"
	if string(data) != expected {
		t.Errorf("expected %s but got %s", expected, string(data))
	}
}
