package api

import (
	"encoding/json"
	"net/http"
	"time"
    "strings"



	"github.com/google/uuid"
	"taskbridge/internal/model"
	"taskbridge/internal/store"
)

type AgentHandler struct {
	Store *store.MemoryStore
}

func (h *AgentHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterAgentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	if req.Hostname == "" {
	writeError(w, http.StatusBadRequest, "hostname required")
	return
}

	agent := &model.Agent{
		ID:           uuid.New().String(),
		Hostname:     req.Hostname,
		OS:           req.OS,
		Arch:         req.Arch,
		Version:      req.Version,
		Capabilities: req.Capabilities,
		LastSeen:     time.Now(),
		Status:       "ONLINE",
	}

	h.Store.CreateAgent(agent)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

func (h *AgentHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) < 4 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	agentID := parts[2]

	if !h.Store.UpdateHeartbeat(agentID) {
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "heartbeat received",
	})
}

func (h *AgentHandler) NextJob(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) < 4 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	agentID := parts[2]

	job := h.Store.GetNextPendingJob()
	if job == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(model.NextJobResponse{})
		return
	}

	job.Status = model.JobRunning
	job.AssignedAgentID = agentID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.NextJobResponse{
		Job: job,
	})
}