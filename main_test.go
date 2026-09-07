package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	baseURL       = "http://localhost:5555/scheduler"
	algorithmName = "hello-world"

	createRunURLTemplate        = "%s/algorithm/v1/run/%s"
	getRunResultURLTemplate     = "%s/algorithm/v1/results/%s/requests/%s"
	getRunMetadataURLTemplate   = "%s/algorithm/v1/metadata/%s/requests/%s"
	getTaggedResultsURLTemplate = "%s/algorithm/v1/results/tags/%s"
	cancelRunURLTemplate        = "%s/algorithm/v1/cancel/%s/requests/%s"
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
	return createTestRunWithDryRun(client, tag, false)
}

func createTestRunWithDryRun(client *http.Client, tag string, dryRun bool) (string, error) {
	postURL := fmt.Sprintf(createRunURLTemplate, baseURL, algorithmName)
	if dryRun {
		postURL = fmt.Sprintf("%s?dryRun=true", postURL)
	}

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

	return runResp.RequestId, nil
}

func waitForCondition(condition wait.ConditionWithContextFunc) error {
	return wait.PollUntilContextTimeout(context.Background(), 100*time.Millisecond, 10*time.Second, true, condition)
}

func waitFor20x(client *http.Client, getURL string, target interface{}, ready func() bool) error {
	return waitForCondition(func(ctx context.Context) (bool, error) {
		resp, err := client.Get(getURL)
		if err != nil {
			return false, err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return false, fmt.Errorf("expected status %d (OK), got %d: %s", http.StatusOK, resp.StatusCode, string(body))
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		if err := json.Unmarshal(bodyBytes, target); err != nil {
			return false, err
		}

		if ready != nil {
			return ready(), nil
		}

		return true, nil
	})
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

func Test_Smoke_CreateRun_DryRun(t *testing.T) {
	client := getHTTPClient()
	tag := fmt.Sprintf("smoke-dryrun-%s", uuid.New().String()[:8])
	requestId, err := createTestRunWithDryRun(client, tag, true)
	if err != nil {
		t.Fatalf("failed to create dry run test run: %v", err)
	}

	if requestId == "" {
		t.Fatal("expected non-empty requestId in response, got empty string")
	}

	if _, err := uuid.Parse(requestId); err != nil {
		t.Fatalf("expected valid UUID for requestId, got %s: %v", requestId, err)
	}

	getURL := fmt.Sprintf(getRunResultURLTemplate, baseURL, algorithmName, requestId)
	var result requestResultResponse

	err = waitFor20x(client, getURL, &result, func() bool {
		return result.Status == "COMPLETED"
	})
	if err != nil {
		t.Fatalf("timed out waiting for dry run request %s to reach COMPLETED stage (status: %s): %v", requestId, result.Status, err)
	}

	if result.RequestId != requestId {
		t.Fatalf("expected requestId %s, got %s", requestId, result.RequestId)
	}
}

func Test_Smoke_GetRunResult(t *testing.T) {
	client := getHTTPClient()
	tag := fmt.Sprintf("smoke-result-%s", uuid.New().String()[:8])
	requestId, err := createTestRun(client, tag)
	if err != nil {
		t.Fatalf("failed to create test run: %v", err)
	}

	getURL := fmt.Sprintf(getRunResultURLTemplate, baseURL, algorithmName, requestId)
	var result requestResultResponse

	err = waitFor20x(client, getURL, &result, nil)
	if err != nil {
		t.Fatalf("timed out waiting for run result for %s: %v", requestId, err)
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

	getURL := fmt.Sprintf(getRunMetadataURLTemplate, baseURL, algorithmName, requestId)
	var meta runMetadataResponse

	err = waitFor20x(client, getURL, &meta, func() bool {
		return meta.PayloadUri != ""
	})
	if err != nil {
		t.Fatalf("timed out waiting for request %s to reach BUFFERED stage (payload_uri is empty, stage: %s): %v", requestId, meta.LifecycleStage, err)
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

	getURL := fmt.Sprintf(getTaggedResultsURLTemplate, baseURL, tag)
	var taggedResults []taggedResultResponse

	err = waitFor20x(client, getURL, &taggedResults, func() bool {
		for _, res := range taggedResults {
			if res.RequestId == requestId {
				return true
			}
		}
		return false
	})
	if err != nil {
		t.Fatalf("timed out waiting to find requestId %s in tagged results for tag %s (got: %v): %v", requestId, tag, taggedResults, err)
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
	if err != nil {
		t.Fatalf("failed to create test run: %v", err)
	}

	postURL := fmt.Sprintf(cancelRunURLTemplate, baseURL, algorithmName, requestId)
	payload := map[string]interface{}{
		"initiator": "smoke-test",
		"reason":    "testing run cancellation endpoint",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	err = waitForCondition(func(ctx context.Context) (bool, error) {
		req, err := http.NewRequest(http.MethodPost, postURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return false, fmt.Errorf("failed to create HTTP request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return false, fmt.Errorf("failed to execute POST request to %s: %w", postURL, err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return false, fmt.Errorf("expected status %d (OK), got %d: %s", http.StatusOK, resp.StatusCode, string(body))
		}

		return true, nil
	})
	if err != nil {
		t.Fatalf("timed out waiting to cancel run %s: %v", requestId, err)
	}
}
