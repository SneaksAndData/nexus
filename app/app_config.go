package app

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"k8s.io/apimachinery/pkg/api/resource"
)

type SchedulerConfig struct {
	S3Buffer            request.S3BufferConfig       `mapstructure:"s3-buffer,omitempty"`
	AstraCqlStore       request.AstraBundleConfig    `mapstructure:"astra-cql-store,omitempty"`
	ScyllaCqlStore      request.ScyllaCqlStoreConfig `mapstructure:"scylla-cql-store,omitempty"`
	CqlStoreType        string                       `mapstructure:"cql-store-type,omitempty"`
	ResourceNamespace   string                       `mapstructure:"resource-namespace,omitempty"`
	KubeConfigPath      string                       `mapstructure:"kube-config-path,omitempty"`
	ShardKubeConfigPath string                       `mapstructure:"shard-kube-config-path,omitempty"`
	LogLevel            string                       `mapstructure:"log-level,omitempty"`
	MaxPayloadSize      string                       `mapstructure:"max-payload-size,omitempty"`
}

const (
	CqlStoreAstra  = "astra"
	CqlStoreScylla = "scylla"
)

func (c *SchedulerConfig) MaxPayloadSizeBytes() int64 {
	var quantity = resource.MustParse(c.MaxPayloadSize)
	return quantity.Value()
}
