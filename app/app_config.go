package app

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/store/cassandra"
	"k8s.io/apimachinery/pkg/api/resource"
)

type SchedulerConfig struct {
	S3Buffer            request.S3BufferConfig      `mapstructure:"s3-buffer,omitempty"`
	AstraCqlStore       cassandra.AstraBundleConfig `mapstructure:"astra-cql-store,omitempty"`
	ScyllaCqlStore      cassandra.ScyllaConfig      `mapstructure:"scylla-cql-store,omitempty"`
	KeyspacesCqlStore   cassandra.KeyspacesConfig   `mapstructure:"keyspaces-cql-store,omitempty"`
	CqlStoreType        string                      `mapstructure:"cql-store-type,omitempty"`
	DeployNamespace     string                      `mapstructure:"deploy-namespace,omitempty"`
	RuntimeNamespace    string                      `mapstructure:"runtime-namespace,omitempty"`
	KubeConfigPath      string                      `mapstructure:"kube-config-path,omitempty"`
	ShardKubeConfigPath string                      `mapstructure:"shard-kube-config-path,omitempty"`
	LogLevel            string                      `mapstructure:"log-level,omitempty"`
	MaxPayloadSize      string                      `mapstructure:"max-payload-size,omitempty"`
	ExternalHostname    string                      `mapstructure:"external-hostname,omitempty"`
}

const (
	CqlStoreAstra     = "cassandra-astra"
	CqlStoreScylla    = "cassandra-scylla"
	CqlStoreKeyspaces = "cassandra-keyspaces"
)

func (c *SchedulerConfig) MaxPayloadSizeBytes() int64 { // coverage-ignore
	var quantity = resource.MustParse(c.MaxPayloadSize)
	return quantity.Value()
}
