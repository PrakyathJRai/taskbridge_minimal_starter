package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"taskbridge/internal/api"
	"taskbridge/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "server listen address")
	flag.Parse()

	memStore := store.NewMemoryStore()

	jobHandler := &api.JobHandler{
		Store: memStore,
	}

	agentHandler := &api.AgentHandler{
		Store: memStore,
	}

	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"taskbridge-server"}`))
	})

	// Jobs
	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			jobHandler.CreateJob(w, r)
		case http.MethodGet:
			jobHandler.ListJobs(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Single Job
mux.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) {

	if strings.HasSuffix(r.URL.Path, "/result") {
		if r.Method == http.MethodPost {
			jobHandler.SubmitResult(w, r)
			return
		}
	}

	if r.Method == http.MethodGet {
		jobHandler.GetJob(w, r)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
})

	// Agent Registration
	mux.HandleFunc("/agents/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			agentHandler.Register(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/agents/", func(w http.ResponseWriter, r *http.Request) {

	if strings.HasSuffix(r.URL.Path, "/heartbeat") {
		if r.Method == http.MethodPost {
			agentHandler.Heartbeat(w, r)
			return
		}
	}

	if strings.HasSuffix(r.URL.Path, "/next-job") {
		if r.Method == http.MethodPost {
			agentHandler.NextJob(w, r)
			return
		}
	}

	http.NotFound(w, r)
})

	fmt.Printf("TaskBridge server listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}