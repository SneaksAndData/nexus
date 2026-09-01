package main_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

type createRunResponse struct {
	RequestId string `json:"requestId"`
}

func Test_Smoke_CreateRun(t *testing.T) {
	baseURL := "http://localhost:5555/scheduler"
	postURL := baseURL + "/algorithm/v1/run/hello-world"

	payload := map[string]interface{}{
		"algorithmParameters": map[string]interface{}{
			"message": "hello world",
		},
		"tag": "smoke-test",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, postURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to execute POST request to %s: %v", postURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d (Accepted), got %d: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var runResp createRunResponse
	if err := json.Unmarshal(bodyBytes, &runResp); err != nil {
		t.Fatalf("failed to unmarshal response JSON: %v, body: %s", err, string(bodyBytes))
	}

	if runResp.RequestId == "" {
		t.Fatalf("expected non-empty requestId in response, got empty string")
	}

	if _, err := uuid.Parse(runResp.RequestId); err != nil {
		t.Fatalf("expected valid UUID for requestId, got %s: %v", runResp.RequestId, err)
	}
}
