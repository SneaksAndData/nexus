package app

import (
	"context"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	nexusconf "github.com/SneaksAndData/nexus-core/pkg/configurations"
	"os"
	"reflect"
	"testing"
	"time"
)

func getExpectedConfig(storagePath string) *SchedulerConfig {
	return &SchedulerConfig{
		S3Buffer: request.S3BufferConfig{
			BufferConfig: &request.BufferConfig{
				PayloadStoragePath:         storagePath,
				PayloadValidFor:            time.Hour * 24,
				FailureRateMaxDelay:        time.Second * 1,
				FailureRateBaseDelay:       time.Millisecond * 100,
				RateLimitElementsPerSecond: 10,
				RateLimitElementsBurst:     100,
				Workers:                    10,
			},
			AccessKeyID:     "test",
			SecretAccessKey: "test",
			Endpoint:        "http://127.0.0.1:9000",
			Region:          "us-east-1",
		},
		AstraCqlStore: request.AstraBundleConfig{
			SecureConnectionBundleBase64: "base64value",
			GatewayUser:                  "user",
			GatewayPassword:              "password",
		},
		CqlStoreType:        CqlStoreAstra,
		ResourceNamespace:   "nexus",
		KubeConfigPath:      "/tmp/nexus-test",
		ShardKubeConfigPath: "/tmp/shards",
		MaxPayloadSize:      "500Mi",
		LogLevel:            "debug",
	}
}

func Test_LoadConfig(t *testing.T) {
	var expected = getExpectedConfig("s3://bucket/nexus/payloads")

	var result = nexusconf.LoadConfig[SchedulerConfig](context.TODO())
	if !reflect.DeepEqual(*expected, result) {
		t.Errorf("LoadConfig failed, expected %v, got %v", *expected, result)
	}
}

func Test_LoadConfigFromEnv(t *testing.T) {
	storagePath := "s3://bucket-2/nexus/payloads"
	keyId := "test-key-id"
	_ = os.Setenv("NEXUS__S3_BUFFER__BUFFER_CONFIG__PAYLOAD_STORAGE_PATH", storagePath)
	_ = os.Setenv("NEXUS__S3_BUFFER__ACCESS_KEY_ID", keyId)

	var expected = getExpectedConfig(storagePath)
	expected.S3Buffer.AccessKeyID = keyId

	var result = nexusconf.LoadConfig[SchedulerConfig](context.TODO())
	if !reflect.DeepEqual(*expected, result) {
		t.Errorf("LoadConfig failed, expected %v, got %v", *expected, result)
	}
}
