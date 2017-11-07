package main

import (
	"context"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/olivere/elastic.v5"
)

var (
	host  = flag.String("host", ":5000", "server host and port")
	esURL = flag.String("es_url", "http://127.0.0.1:9200", "elasticsearch url")

	esClient *elastic.Client
	index    *template.Template
)

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", Log(Index))
	r.HandleFunc("/get/{index}/{type}/{id}", Log(Get))
	r.HandleFunc("/set/{index}/{type}/{id}", Log(Set))

	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)
	log.Print(http.Get(*esURL))

	log.Printf("Connecting to Elasticsearch on %s...", *esURL)
	c, err := elastic.NewSimpleClient(elastic.SetURL(*esURL))
	if err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}
	esClient = c

	index, err = template.ParseFiles("static/index.html")
	if err != nil {
		log.Fatalf("Failed to parse index template: %v", err)
	}

	log.Printf("Listening on %s...", *host)
	log.Fatal(http.ListenAndServe(*host, r))
}

func Log(h func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL)
		h(w, r)
	})
}

func Index(w http.ResponseWriter, r *http.Request) {
	var err error
	index, err = template.ParseFiles("static/index.html")
	if err != nil {
		internalError(w, err)
		return
	}

	mappings, err := esClient.GetMapping().Do(context.Background())
	if err != nil {
		internalError(w, err)
		return
	}

	indices := map[string][]string{}
	for k, v := range mappings {
		indexTypes := []string{}
		if indexMapping, ok := v.(map[string]interface{}); ok {
			if types, ok := indexMapping["mappings"].(map[string]interface{}); ok {
				for t, _ := range types {
					indexTypes = append(indexTypes, t)
				}
			}
		}
		indices[k] = indexTypes
	}

	err = index.Execute(w, indices)
	if err != nil {
		internalError(w, err)
		return
	}
}

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	d, err := esClient.
		Get().
		Index(vars["index"]).
		Type(vars["type"]).
		Id(vars["id"]).
		Do(context.Background())
	if err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(*d.Source)
}

func Set(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	decoder := json.NewDecoder(r.Body)
	var doc map[string]interface{}
	err := decoder.Decode(&doc)
	if err != nil {
		internalError(w, err)
		return
	}
	defer r.Body.Close()

	d, err := esClient.
		Update().
		Index(vars["index"]).
		Type(vars["type"]).
		Id(vars["id"]).
		Doc(doc).
		Do(context.Background())
	if err != nil {
		internalError(w, err)
		return
	}
	log.Println(d)
	w.WriteHeader(http.StatusOK)
}
func internalError(w http.ResponseWriter, err error) {
	log.Println("ERROR", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}
