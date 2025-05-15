package app

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
)

type SchedulerConfig struct {
	S3Buffer            request.S3BufferConfig    `mapstructure:"s3-buffer,omitempty"`
	CqlStore            request.AstraBundleConfig `mapstructure:"cql-store,omitempty"`
	ResourceNamespace   string                    `mapstructure:"resource-namespace,omitempty"`
	KubeConfigPath      string                    `mapstructure:"kube-config-path,omitempty"`
	ShardKubeConfigPath string                    `mapstructure:"shard-kube-config-path,omitempty"`
}
