package app

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/payload"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/store/cassandra"
	nexusconf "github.com/SneaksAndData/nexus-core/pkg/configurations"
)

func getExpectedConfig(storagePath string) *SchedulerConfig {
	return &SchedulerConfig{
		S3Buffer: request.S3BufferConfig{
			PayloadStoragePath: storagePath,
			RequestPayloadProxyConfiguration: &payload.RequestPayloadProxyConfiguration{
				TenantId:          "test-tenant",
				ServePathTemplate: "/data/v1/payloads/%s/%s",
				SignSecret:        "test-secret",
			},
			BufferConfig: &request.BufferConfig{
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
		AstraCqlStore: cassandra.AstraBundleConfig{
			SecureConnectionBundleBase64: "base64value",
			GatewayUser:                  "user",
			GatewayPassword:              "password",
			IndexesSupported:             true,
			Keyspace:                     "nexus",
		},
		ScyllaCqlStore: cassandra.ScyllaConfig{
			Hosts:            []string{"127.0.0.1:9000"},
			Port:             "",
			User:             "",
			Password:         "",
			LocalDC:          "",
			IndexesSupported: true,
			Keyspace:         "nexus",
		},
		KeyspacesCqlStore: cassandra.KeyspacesConfig{
			Hosts:    []string{"keyspaces.aws.com"},
			Port:     "9042",
			CaPath:   "/tmp/ca",
			Region:   "us-east-1",
			Keyspace: "nexus",
		},
		CqlStoreType:        CqlStoreAstra,
		RuntimeNamespace:    "nexus",
		DeployNamespace:     "nexus-scheduler",
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
	host1 := "127.0.0.1:9042"
	host2 := "127.0.0.2:9042"
	_ = os.Setenv("NEXUS__S3_BUFFER__PAYLOAD_STORAGE_PATH", storagePath)
	_ = os.Setenv("NEXUS__S3_BUFFER__ACCESS_KEY_ID", keyId)
	_ = os.Setenv("NEXUS__SCYLLA_CQL_STORE__HOSTS", fmt.Sprintf("%s,%s", host1, host2))

	var expected = getExpectedConfig(storagePath)
	expected.S3Buffer.AccessKeyID = keyId
	expected.ScyllaCqlStore.Hosts = []string{host1, host2}

	var result = nexusconf.LoadConfig[SchedulerConfig](context.TODO())
	if !reflect.DeepEqual(*expected, result) {
		t.Errorf("LoadConfig failed, expected %v, got %v", *expected, result)
	}
}
