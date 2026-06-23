package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"taskbridge/internal/model"
	"taskbridge/internal/store"
)

type JobHandler struct {
	Store *store.MemoryStore
}

func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req model.CreateJobRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "job name required")
		return
	}

	if req.Type == "" {
		writeError(w, http.StatusBadRequest, "job type required")
		return
	}

	job := &model.Job{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Type:           req.Type,
		Payload:        req.Payload,
		Status:         model.JobPending,
		CreatedAt:      time.Now(),
		MaxRetries:     req.MaxRetries,
		TimeoutSeconds: req.TimeoutSeconds,
	}

	h.Store.CreateJob(job)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs := h.Store.ListJobs()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")

job, ok := h.Store.GetJob(id)
if !ok {
	writeError(w, http.StatusNotFound, "job not found")
	return
}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (h *JobHandler) SubmitResult(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, "/")

	if len(parts) < 4 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	jobID := parts[2]

	var req model.JobResultRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	writeError(w, http.StatusBadRequest, err.Error())
	return
}

	job, exists := h.Store.GetJob(jobID)

	if !exists {
	writeError(w, http.StatusBadRequest, "invalid request")
		
		return
	}

	// Retry logic
	if req.Status == model.JobFailed &&
		job.AttemptCount < job.MaxRetries {

		h.Store.RetryJob(jobID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "job scheduled for retry",
		})

		return
	}

	ok := h.Store.UpdateJobResult(
		jobID,
		req.Status,
		req.Result,
		req.Logs,
		req.Error,
	)

	if !ok {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "result accepted",
	})
}