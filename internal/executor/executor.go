package executor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"taskbridge/internal/model"
)

// Result is returned after executing a job.
type Result struct {
	Status model.JobStatus `json:"status"`
	Logs   []string        `json:"logs"`
	Result map[string]any  `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// Executor executes a single job type.
type Executor interface {
	Type() model.JobType
	Execute(ctx context.Context, job model.Job) Result
}

// Registry maps job types to executors.
type Registry struct {
	executors map[model.JobType]Executor
}

func NewRegistry() *Registry {
	return &Registry{
		executors: map[model.JobType]Executor{},
	}
}

func (r *Registry) Register(ex Executor) {
	r.executors[ex.Type()] = ex
}

func (r *Registry) Get(t model.JobType) (Executor, bool) {
	ex, ok := r.executors[t]
	return ex, ok
}

// ---------------- WAIT ----------------

type WaitExecutor struct{}

func (e WaitExecutor) Type() model.JobType {
	return model.JobWait
}

func (e WaitExecutor) Execute(ctx context.Context, job model.Job) Result {

	seconds, ok := job.Payload["seconds"].(float64)
	if !ok {
		return Result{
			Status: model.JobFailed,
			Error:  "missing seconds",
		}
	}

	select {
	case <-time.After(time.Duration(seconds) * time.Second):
		return Result{
			Status: model.JobSuccess,
			Logs:   []string{"wait completed"},
		}

	case <-ctx.Done():
		return Result{
			Status: model.JobFailed,
			Error:  ctx.Err().Error(),
		}
	}
}

// ---------------- HTTP CHECK ----------------

type HTTPCheckExecutor struct{}

func (e HTTPCheckExecutor) Type() model.JobType {
	return model.JobHTTPCheck
}

func (e HTTPCheckExecutor) Execute(ctx context.Context, job model.Job) Result {

	url, ok := job.Payload["url"].(string)
	if !ok {
		return Result{
			Status: model.JobFailed,
			Error:  "missing url",
		}
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  err.Error(),
		}
	}

	defer resp.Body.Close()

	return Result{
		Status: model.JobSuccess,
		Result: map[string]any{
			"status_code": resp.StatusCode,
		},
	}
}

// ---------------- TCP CHECK ----------------

type TCPCheckExecutor struct{}

func (e TCPCheckExecutor) Type() model.JobType {
	return model.JobTCPCheck
}

func (e TCPCheckExecutor) Execute(ctx context.Context, job model.Job) Result {

	address, ok := job.Payload["address"].(string)
	if !ok {
		return Result{
			Status: model.JobFailed,
			Error:  "missing address",
		}
	}

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  err.Error(),
		}
	}

	conn.Close()

	return Result{
		Status: model.JobSuccess,
	}
}

// ---------------- FILE EXISTS ----------------

type FileExistsExecutor struct{}

func (e FileExistsExecutor) Type() model.JobType {
	return model.JobFileExists
}

func (e FileExistsExecutor) Execute(ctx context.Context, job model.Job) Result {

	path, ok := job.Payload["path"].(string)
	if !ok {
		return Result{
			Status: model.JobFailed,
			Error:  "missing path",
		}
	}

	_, err := os.Stat(path)
	if err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  err.Error(),
		}
	}

	return Result{
		Status: model.JobSuccess,
	}
}

// ---------------- CHECKSUM ----------------

type ChecksumExecutor struct{}

func (e ChecksumExecutor) Type() model.JobType {
	return model.JobChecksum
}

func (e ChecksumExecutor) Execute(ctx context.Context, job model.Job) Result {

	path, ok := job.Payload["path"].(string)
	if !ok {
		return Result{
			Status: model.JobFailed,
			Error:  "missing path",
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  err.Error(),
		}
	}

	defer file.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  err.Error(),
		}
	}

	return Result{
		Status: model.JobSuccess,
		Result: map[string]any{
			"checksum": hex.EncodeToString(hash.Sum(nil)),
		},
	}
}

// ---------------- WRITE FILE ----------------

type WriteFileExecutor struct{}

func (e WriteFileExecutor) Type() model.JobType {
	return model.JobWriteFile
}

func (e WriteFileExecutor) Execute(ctx context.Context, job model.Job) Result {

	path, ok := job.Payload["path"].(string)
	if !ok {
		return Result{
			Status: model.JobFailed,
			Error:  "missing path",
		}
	}

	content, ok := job.Payload["content"].(string)
	if !ok {
		return Result{
			Status: model.JobFailed,
			Error:  "missing content",
		}
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return Result{
			Status: model.JobFailed,
			Error:  err.Error(),
		}
	}

	return Result{
		Status: model.JobSuccess,
	}
}

func RegisterDefaults(r *Registry) {
	r.Register(WaitExecutor{})
	r.Register(HTTPCheckExecutor{})
	r.Register(TCPCheckExecutor{})
	r.Register(FileExistsExecutor{})
	r.Register(ChecksumExecutor{})
	r.Register(WriteFileExecutor{})
}