package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"taskbridge/internal/executor"
	"taskbridge/internal/model"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8082", "TaskBridge server URL")
	pollInterval := flag.Duration("poll-interval", 3*time.Second, "job polling interval")
	flag.Parse()

	// Register agent
	agentReq := model.RegisterAgentRequest{
		Hostname: "local-agent",
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Version:  "1.0",
		Capabilities: []model.JobType{
			model.JobWait,
			model.JobHTTPCheck,
			model.JobTCPCheck,
			model.JobFileExists,
			model.JobChecksum,
			model.JobWriteFile,
		},
	}

	body, _ := json.Marshal(agentReq)

	resp, err := http.Post(
		*serverURL+"/agents/register",
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	var agent model.Agent

	if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
		panic(err)
	}

	fmt.Println("Registered agent:", agent.ID)

	registry := executor.NewRegistry()
    executor.RegisterDefaults(registry)


	// Heartbeat loop
	go func() {
		for {
			http.Post(
				*serverURL+"/agents/"+agent.ID+"/heartbeat",
				"application/json",
				nil,
			)

			time.Sleep(10 * time.Second)
		}
	}()

	// Poll loop
	for {

		resp, err := http.Post(
			*serverURL+"/agents/"+agent.ID+"/next-job",
			"application/json",
			nil,
		)

		if err != nil {
			fmt.Println("poll error:", err)
			time.Sleep(*pollInterval)
			continue
		}

		var next model.NextJobResponse

		json.NewDecoder(resp.Body).Decode(&next)
		resp.Body.Close()

		if next.Job == nil {
			time.Sleep(*pollInterval)
			continue
		}

	ex, ok := registry.Get(next.Job.Type)

var resultReq model.JobResultRequest

if !ok {

	resultReq = model.JobResultRequest{
		AgentID: agent.ID,
		Status:  model.JobFailed,
		Error:   "unsupported job type",
	}

} else {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(next.Job.TimeoutSeconds)*time.Second,
	)

	res := ex.Execute(ctx, *next.Job)

	cancel()

	resultReq = model.JobResultRequest{
		AgentID: agent.ID,
		Status:  res.Status,
		Result:  res.Result,
		Logs:    res.Logs,
		Error:   res.Error,
	}
}

		payload, _ := json.Marshal(resultReq)

		http.Post(
			*serverURL+"/jobs/"+next.Job.ID+"/result",
			"application/json",
			bytes.NewBuffer(payload),
		)

		fmt.Println("Submitted result for:", next.Job.ID)

		time.Sleep(*pollInterval)
	}
}