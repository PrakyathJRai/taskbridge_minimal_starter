package store

import (
	"testing"

	"taskbridge/internal/model"
)

func TestCreateAgent(t *testing.T) {
	s := NewMemoryStore()

	agent := &model.Agent{
		ID: "agent1",
	}

	s.CreateAgent(agent)

	_, ok := s.GetAgent("agent1")

	if !ok {
		t.Fatal("agent was not stored")
	}
}