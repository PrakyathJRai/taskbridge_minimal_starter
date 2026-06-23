package store

import (
	"testing"

	"taskbridge/internal/model"
)

func TestCreateJob(t *testing.T) {
	s := NewMemoryStore()

	job := &model.Job{
		ID: "job1",
		Name: "test-job",
	}

	s.CreateJob(job)

	_, ok := s.GetJob("job1")

	if !ok {
		t.Fatal("job was not stored")
	}
}