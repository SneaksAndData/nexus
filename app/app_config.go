package app

import (
	"context"
	"fmt"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"os"
	"strings"
)

const (
	EnvPrefix = "NEXUS_" // varnames will be NEXUS__MY_ENV_VAR
)

type SchedulerConfig struct {
	Buffer              request.BufferConfig      `mapstructure:"buffer"`
	CqlStore            request.AstraBundleConfig `mapstructure:"cql-store"`
	ResourceNamespace   string                    `mapstructure:"resource-namespace"`
	KubeConfigPath      string                    `mapstructure:"kube-config-path"`
	ShardKubeConfigPath string                    `mapstructure:"shard-kube-config-path,omitempty"`
}

func LoadConfig(ctx context.Context) SchedulerConfig {
	logger := klog.FromContext(ctx)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetConfigFile(fmt.Sprintf("appconfig.%s.yaml", strings.ToLower(os.Getenv("APPLICATION_ENVIRONMENT"))))
	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Error(err, "Error loading application config from appconfig.yaml")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	var appConfig SchedulerConfig
	err := viper.Unmarshal(&appConfig)

	if err != nil {
		logger.Error(err, "Error loading application config from appconfig.yaml")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	return appConfig
}
