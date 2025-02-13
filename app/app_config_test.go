package app

import (
	"context"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"os"
	"reflect"
	"testing"
	"time"
)

func getExpectedConfig(storagePath string) *SchedulerConfig {
	return &SchedulerConfig{
		Buffer: request.BufferConfig{
			PayloadStoragePath:         storagePath,
			PayloadValidFor:            time.Hour * 24,
			FailureRateMaxDelay:        time.Second * 1,
			FailureRateBaseDelay:       time.Millisecond * 100,
			RateLimitElementsPerSecond: 10,
			RateLimitElementsBurst:     100,
			Workers:                    10,
		},
		CqlStore: request.AstraBundleConfig{
			SecureConnectionBundleBase64: "base64value",
			GatewayUser:                  "user",
			GatewayPassword:              "password",
		},
		ResourceNamespace: "crystal",
		KubeConfigPath:    "/tmp/nexus-test",
	}
}

func Test_LoadConfig(t *testing.T) {
	var expected = getExpectedConfig("s3://bucket/nexus/payloads")

	var result = LoadConfig(context.TODO())
	if !reflect.DeepEqual(*expected, result) {
		t.Errorf("LoadConfig failed, expected %v, got %v", *expected, result)
	}
}

func Test_LoadConfigFromEnv(t *testing.T) {
	_ = os.Setenv("NEXUS__BUFFER.PAYLOAD_STORAGE_PATH", "s3://bucket-2/nexus/payloads")
	var expected = getExpectedConfig("s3://bucket-2/nexus/payloads")

	var result = LoadConfig(context.TODO())
	if !reflect.DeepEqual(*expected, result) {
		t.Errorf("LoadConfig failed, expected %v, got %v", *expected, result)
	}
}
