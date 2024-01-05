package main

import (
	"encoding/json"
	"flag"
	"fmt"
	template2 "html/template"
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

var indexTemplate *template2.Template

type Alert struct {
	UUID string     `json:",omitempty"`
	Time *time.Time `json:",omitempty"`
	Info string
}

type indexData struct {
	IsEmpty bool
	Alerts  []Alert
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
	_ = indexTemplate.Execute(w, indexData{
		IsEmpty: len(alerts) == 0,
		Alerts:  alerts,
	})
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

func init() {
	indexTemplate = template2.Must(template2.New("index.gohtml").ParseFiles("web/index.gohtml"))
}

func populate(min, max int, storage AlertStore) {
	for idx := 0; idx < (rand.Intn(max-min) + min); idx++ {
		now := time.Now()
		storage.Store(Alert{
			UUID: uuid.NewString(),
			Time: &now,
			Info: fmt.Sprintf("Mighty demo alert number %d", idx),
		})
	}
}

func main() {
	log.Println("ELF WATCH IS THE BEST")
	demo := flag.Bool("demo", false, "Use in-memory demo data")
	flag.Parse()

	storage := InMemoryAlertStore{}
	if *demo {
		populate(10, 20, &storage)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	http.HandleFunc("/", indexHandler(&storage))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
