package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	baseURL       = "http://localhost:5555/scheduler"
	algorithmName = "hello-world"
)

type createRunResponse struct {
	RequestId string `json:"requestId"`
}

type requestResultResponse struct {
	RequestId       string `json:"requestId"`
	Status          string `json:"status"`
	ResultUri       string `json:"resultUri"`
	RunErrorMessage string `json:"runErrorMessage"`
}

type runMetadataResponse struct {
	Id             string `json:"id"`
	Algorithm      string `json:"algorithm"`
	LifecycleStage string `json:"lifecycle_stage"`
	Tag            string `json:"tag"`
	PayloadUri     string `json:"payload_uri"`
}

type taggedResultResponse struct {
	RequestId       string `json:"requestId"`
	AlgorithmName   string `json:"algorithmName"`
	Status          string `json:"status"`
	ResultUri       string `json:"resultUri"`
	RunErrorMessage string `json:"runErrorMessage"`
}

func getHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

func createTestRun(client *http.Client, tag string) (string, error) {
	postURL := fmt.Sprintf("%s/algorithm/v1/run/%s", baseURL, algorithmName)

	payload := map[string]interface{}{
		"algorithmParameters": map[string]interface{}{
			"hello_text":   "hello world",
			"hello_author": "unit-test",
		},
		"tag": tag,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, postURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute POST request to %s: %w", postURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status %d (Accepted), got %d: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var runResp createRunResponse
	if err := json.Unmarshal(bodyBytes, &runResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response JSON: %w, body: %s", err, string(bodyBytes))
	}

	// Small delay to allow asynchronous job submission and checkpointing to take place
	time.Sleep(100 * time.Millisecond)

	return runResp.RequestId, nil
}

func Test_Smoke_CreateRun(t *testing.T) {
	client := getHTTPClient()
	tag := fmt.Sprintf("smoke-create-%s", uuid.New().String()[:8])
	requestId, err := createTestRun(client, tag)
	if err != nil {
		t.Fatalf("failed to create test run: %v", err)
	}

	if requestId == "" {
		t.Fatal("expected non-empty requestId in response, got empty string")
	}

	if _, err := uuid.Parse(requestId); err != nil {
		t.Fatalf("expected valid UUID for requestId, got %s: %v", requestId, err)
	}
}

func Test_Smoke_GetRunResult(t *testing.T) {
	client := getHTTPClient()
	tag := fmt.Sprintf("smoke-result-%s", uuid.New().String()[:8])
	requestId, err := createTestRun(client, tag)
	if err != nil {
		t.Fatalf("failed to create test run: %v", err)
	}

	getURL := fmt.Sprintf("%s/algorithm/v1/results/%s/requests/%s", baseURL, algorithmName, requestId)
	resp, err := client.Get(getURL)
	if err != nil {
		t.Fatalf("failed to execute GET request to %s: %v", getURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d (OK), got %d: %s", http.StatusOK, resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var result requestResultResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		t.Fatalf("failed to unmarshal result JSON: %v, body: %s", err, string(bodyBytes))
	}

	if result.RequestId != requestId {
		t.Fatalf("expected requestId %s, got %s", requestId, result.RequestId)
	}
}

func Test_Smoke_GetRunMetadata(t *testing.T) {
	client := getHTTPClient()
	tag := fmt.Sprintf("smoke-meta-%s", uuid.New().String()[:8])
	requestId, err := createTestRun(client, tag)
	if err != nil {
		t.Fatalf("failed to create test run: %v", err)
	}

	getURL := fmt.Sprintf("%s/algorithm/v1/metadata/%s/requests/%s", baseURL, algorithmName, requestId)

	var meta runMetadataResponse
	deadline := time.Now().Add(10 * time.Second)

	for {
		resp, err := client.Get(getURL)
		if err != nil {
			t.Fatalf("failed to execute GET request to %s: %v", getURL, err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			t.Fatalf("expected status %d (OK), got %d: %s", http.StatusOK, resp.StatusCode, string(body))
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if err := json.Unmarshal(bodyBytes, &meta); err != nil {
			t.Fatalf("failed to unmarshal metadata JSON: %v, body: %s", err, string(bodyBytes))
		}

		if meta.PayloadUri != "" {
			break
		}

		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for request %s to reach BUFFERED stage (payload_uri is empty, stage: %s)", requestId, meta.LifecycleStage)
		}

		time.Sleep(100 * time.Millisecond)
	}

	if meta.Id != requestId {
		t.Fatalf("expected metadata id %s, got %s", requestId, meta.Id)
	}

	if meta.Algorithm != algorithmName {
		t.Fatalf("expected metadata algorithm %s, got %s", algorithmName, meta.Algorithm)
	}

	// Verify that the payloadUri works when hostname is replaced with localhost:5555/scheduler (baseURL)
	parsedPayloadURI, err := url.Parse(meta.PayloadUri)
	if err != nil {
		t.Fatalf("failed to parse payload_uri '%s': %v", meta.PayloadUri, err)
	}

	payloadURL := fmt.Sprintf("%s%s", baseURL, parsedPayloadURI.RequestURI())
	payloadResp, err := client.Get(payloadURL)
	if err != nil {
		t.Fatalf("failed to execute GET request to payload URL %s: %v", payloadURL, err)
	}
	defer func() { _ = payloadResp.Body.Close() }()

	if payloadResp.StatusCode != http.StatusOK {
		pBody, _ := io.ReadAll(payloadResp.Body)
		t.Fatalf("expected status %d (OK) from payload URL %s, got %d: %s", http.StatusOK, payloadURL, payloadResp.StatusCode, string(pBody))
	}

	payloadBody, err := io.ReadAll(payloadResp.Body)
	if err != nil {
		t.Fatalf("failed to read payload response body: %v", err)
	}

	var payloadData map[string]interface{}
	if err := json.Unmarshal(payloadBody, &payloadData); err != nil {
		t.Fatalf("failed to unmarshal payload JSON: %v, body: %s", err, string(payloadBody))
	}
}

func Test_Smoke_GetRunResultsByTag(t *testing.T) {
	client := getHTTPClient()
	tag := fmt.Sprintf("smoke-tag-%s", uuid.New().String()[:8])
	requestId, err := createTestRun(client, tag)
	if err != nil {
		t.Fatalf("failed to create test run: %v", err)
	}

	getURL := fmt.Sprintf("%s/algorithm/v1/results/tags/%s", baseURL, tag)
	resp, err := client.Get(getURL)
	if err != nil {
		t.Fatalf("failed to execute GET request to %s: %v", getURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d (OK), got %d: %s", http.StatusOK, resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var taggedResults []taggedResultResponse
	if err := json.Unmarshal(bodyBytes, &taggedResults); err != nil {
		t.Fatalf("failed to unmarshal tagged results JSON: %v, body: %s", err, string(bodyBytes))
	}

	found := false
	for _, res := range taggedResults {
		if res.RequestId == requestId {
			found = true
			if res.AlgorithmName != algorithmName {
				t.Fatalf("expected algorithmName %s, got %s", algorithmName, res.AlgorithmName)
			}
			break
		}
	}

	if !found {
		t.Fatalf("expected to find requestId %s in tagged results for tag %s, but got: %v", requestId, tag, taggedResults)
	}
}

func Test_Smoke_CancelRun(t *testing.T) {
	client := getHTTPClient()
	tag := fmt.Sprintf("smoke-cancel-%s", uuid.New().String()[:8])
	requestId, err := createTestRun(client, tag)

	// wait for buffering
	time.Sleep(1 * time.Second)

	if err != nil {
		t.Fatalf("failed to create test run: %v", err)
	}

	postURL := fmt.Sprintf("%s/algorithm/v1/cancel/%s/requests/%s", baseURL, algorithmName, requestId)
	payload := map[string]interface{}{
		"initiator": "smoke-test",
		"reason":    "testing run cancellation endpoint",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d (OK), got %d: %s", http.StatusOK, resp.StatusCode, string(body))
	}
}
