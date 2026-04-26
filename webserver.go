package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type webServerStorage struct {
	sync.RWMutex
	contentStorage  map[string]map[string]string // map of bucket->hash->actual body
	metadataStorage map[string]map[string]string // map of bucket->id->sha256 hash of body
}

func (server *webServerStorage) handlePut(w http.ResponseWriter, req *http.Request) {
	bucket := req.PathValue("bucket")
	id := req.PathValue("id")

	payload, err := getRequestBody(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error getting request body"))
		return
	}

	requestHash := fmt.Sprintf("%x", sha256.Sum256([]byte(payload)))

	server.Lock()
	if _, ok := server.contentStorage[bucket]; !ok {
		server.contentStorage[bucket] = make(map[string]string)
		server.metadataStorage[bucket] = make(map[string]string)
	}
	server.contentStorage[bucket][requestHash] = string(payload)
	server.metadataStorage[bucket][id] = requestHash
	server.Unlock()

	w.WriteHeader(http.StatusCreated)
	successString := fmt.Sprintf("Status: 201 Created {\"id\":\"%s\"}", id)
	w.Write([]byte(successString))
}

func (server *webServerStorage) handleGet(w http.ResponseWriter, req *http.Request) {
	bucket := req.PathValue("bucket")
	id := req.PathValue("id")

	server.RLock()
	defer server.RUnlock()
	bucketMap, bucketPresent := server.metadataStorage[bucket]
	hash, idPresent := bucketMap[id]

	if bucketPresent && idPresent {
		w.WriteHeader(http.StatusOK)
		successString := fmt.Sprintf("Status: 200 OK {%s}", server.contentStorage[bucket][hash])
		w.Write([]byte(successString))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Status 400 Not Found"))
	}
}

func (server *webServerStorage) handleDelete(w http.ResponseWriter, req *http.Request) {
	bucket := req.PathValue("bucket")
	id := req.PathValue("id")

	server.Lock()
	defer server.Unlock()

	bucketMap, bucketPresent := server.metadataStorage[bucket]
	contentHash, idPresent := bucketMap[id]
	if bucketPresent && idPresent {
		// we can delete the item from the metadata storage
		delete(server.metadataStorage[bucket], id)

		// for deduping, we need to see if there's any remaining use
		// by iterating over the metadata map for this bucket
		stillReferenced := false
		for _, currentHash := range server.metadataStorage[bucket] {
			if currentHash == contentHash {
				stillReferenced = true
				break
			}
		}
		if !stillReferenced {
			delete(server.contentStorage[bucket], contentHash)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Status: 200 OK"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Status 400 Not Found"))
	}
}

func getRequestBody(req *http.Request) (string, error) {
	bodyBytes, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		return "", fmt.Errorf("Unable to read body")
	}

	return string(bodyBytes), nil
}

func main() {

	webServerStorage := &webServerStorage{}
	webServerStorage.contentStorage = make(map[string]map[string]string)
	webServerStorage.metadataStorage = make(map[string]map[string]string)

	// start the server and listen on the port specified by the user
	userPort := flag.String("port", "8080", "port to listen on")
	flag.Parse()
	formattedPort := ":" + *userPort

	http.HandleFunc("GET /objects/{bucket}/{id}", webServerStorage.handleGet)
	http.HandleFunc("PUT /objects/{bucket}/{id}", webServerStorage.handlePut)
	http.HandleFunc("DELETE /objects/{bucket}/{id}", webServerStorage.handleDelete)

	server := &http.Server{
		Addr:         formattedPort,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewBuildInfoCollector(),
	)
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// since we've got multiple endpoints we have to handle them in parallel
	wg := &sync.WaitGroup{}
	wg.Go(func() {
		http.ListenAndServe(":2112", nil)
	})

	wg.Go(func() {
		server.ListenAndServe()
	})
	wg.Wait()
}
