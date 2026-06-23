package store

import (
	"sync"
	"time"

	"taskbridge/internal/model"
)

type MemoryStore struct {
	Jobs   map[string]*model.Job
	Agents map[string]*model.Agent
	Mu     sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Jobs:   make(map[string]*model.Job),
		Agents: make(map[string]*model.Agent),
	}
}

func (s *MemoryStore) CreateJob(job *model.Job) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.Jobs[job.ID] = job
}

func (s *MemoryStore) GetJob(id string) (*model.Job, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	job, ok := s.Jobs[id]
	return job, ok
}

func (s *MemoryStore) ListJobs() []*model.Job {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	var jobs []*model.Job

	for _, job := range s.Jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

func (s *MemoryStore) CreateAgent(agent *model.Agent) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.Agents[agent.ID] = agent
}

func (s *MemoryStore) GetAgent(id string) (*model.Agent, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	agent, ok := s.Agents[id]
	return agent, ok
}

func (s *MemoryStore) ListAgents() []*model.Agent {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	var agents []*model.Agent

	for _, agent := range s.Agents {
		agents = append(agents, agent)
	}

	return agents
}

func (s *MemoryStore) UpdateHeartbeat(id string) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	agent, ok := s.Agents[id]
	if !ok {
		return false
	}

	agent.LastSeen = time.Now()
	agent.Status = "ONLINE"

	return true
}

func (s *MemoryStore) GetNextPendingJob() *model.Job {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for _, job := range s.Jobs {
		if job.Status == model.JobPending {
			return job
		}
	}

	return nil
}

func (s *MemoryStore) UpdateJobResult(
	jobID string,
	status model.JobStatus,
	result map[string]any,
	logs []string,
	errMsg string,
) bool {

	s.Mu.Lock()
	defer s.Mu.Unlock()

	job, ok := s.Jobs[jobID]
	if !ok {
		return false
	}

	now := time.Now()

	job.Status = status
	job.Result = result
	job.Logs = logs
	job.Error = errMsg
	job.FinishedAt = &now

	return true
}

func (s *MemoryStore) RetryJob(jobID string) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	job, ok := s.Jobs[jobID]
	if !ok {
		return false
	}

	job.AttemptCount++

	if job.AttemptCount <= job.MaxRetries {
		job.Status = model.JobPending
		job.AssignedAgentID = ""
		return true
	}

	job.Status = model.JobFailed
	return true
}