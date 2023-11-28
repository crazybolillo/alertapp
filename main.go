package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	defaultPage          = 0
	defaultSize          = 20
	AlertCountHttpHeader = "X-Alertapp-Count"
)

type Alert struct {
	UUID string     `json:",omitempty"`
	Time *time.Time `json:",omitempty"`
	Info string
}

type AlertStore interface {
	Retrieve(page int, size int) []Alert
	Store(alert Alert) Alert
}

type InMemoryAlertStore struct {
	Storage []Alert
}

func (store *InMemoryAlertStore) Retrieve(page int, size int) []Alert {
	offset := page * size
	if offset > len(store.Storage) {
		return []Alert{}
	}

	size = offset + size
	if size >= len(store.Storage) {
		size = len(store.Storage)
	}

	return store.Storage[offset:size]
}

func (store *InMemoryAlertStore) Store(alert Alert) Alert {
	store.Storage = append(store.Storage, alert)
	return store.Storage[len(store.Storage)-1]
}

func indexGet(store AlertStore, w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	page, err := strconv.Atoi(params.Get("page"))
	if err != nil {
		page = defaultPage
	}
	size, err := strconv.Atoi(params.Get("size"))
	if err != nil {
		size = defaultSize
	}
	if page < 0 || size < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	alerts := store.Retrieve(page, size)
	w.Header().Add(AlertCountHttpHeader, strconv.Itoa(len(alerts)))
	for _, alert := range alerts {
		_, _ = fmt.Fprintf(w, "UUID: %s --- %s --- %s\n", alert.UUID, alert.Time, alert.Info)
	}
}

func indexPost(store AlertStore, w http.ResponseWriter, r *http.Request) {
	var alert Alert
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&alert); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now()
	alert.UUID = uuid.NewString()
	alert.Time = &now

	store.Store(alert)
}

func indexHandler(store AlertStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			indexGet(store, w, r)
		case http.MethodPost:
			indexPost(store, w, r)
		}
	}
}

func main() {
	demo := flag.Bool("demo", false, "Use in-memory demo data")
	flag.Parse()

	storage := InMemoryAlertStore{}
	if *demo {
		for idx := 0; idx < (rand.Intn(20) + 10); idx++ {
			now := time.Now()
			storage.Store(Alert{
				UUID: uuid.NewString(),
				Time: &now,
				Info: fmt.Sprintf("Mighty demo alert number %d", idx),
			})
		}
	}

	http.HandleFunc("/", indexHandler(&storage))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
