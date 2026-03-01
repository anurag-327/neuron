package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Limit struct {
	TimeMs   int64 `json:"timeMs"`
	MemoryKB int64 `json:"memoryKB"`
}

type ExecuteRequest struct {
	Code     string `json:"code"`
	Input    string `json:"input"`
	Language string `json:"language"`
	Limit    Limit  `json:"limit"`
}

type TimeMetrics struct {
	Total   int64 `json:"total_ms"`
	Compile int64 `json:"compile_ms"`
	Run     int64 `json:"run_ms"`
}

type ExecuteResult struct {
	Stdout   string      `json:"stdout"`
	Stderr   string      `json:"stderr"`
	ErrType  string      `json:"err_type"`
	ErrMsg   string      `json:"err_msg"`
	ExitCode int64       `json:"exit_code"`
	Metrics  TimeMetrics `json:"metrics"`
}

type HealthStatus struct {
	Runner        string `json:"runner"`
	RunnerError   string `json:"runnerError"`
	RunnerVersion string `json:"runnerVersion"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(url string) *Client {
	return &Client{
		baseURL: url,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Execute calls the internal engine service to run code
func (c *Client) Execute(ctx context.Context, req ExecuteRequest) (*ExecuteResult, error) {
	url := fmt.Sprintf("%s/api/v1/execute", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("engine call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("engine error %d: %s", resp.StatusCode, string(respBytes))
	}

	var result ExecuteResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return &result, nil
}

// GetStatus checks the detailed health status of the core service
func (c *Client) GetStatus(ctx context.Context) (*HealthStatus, error) {
	url := fmt.Sprintf("%s/status", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("status call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status error %d: %s", resp.StatusCode, string(respBytes))
	}

	var status HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decode status failed: %w", err)
	}

	return &status, nil
}

// HealthCheck performs a simple liveness check
func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("request creation failed: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}
