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
	EnvPrefix = "NEXUS_" // varnames will be NEXUS__MY_ENV_VAR or NEXUS__SECTION1__SECTION2__MY_ENV_VAR
)

type SchedulerConfig struct {
	S3Buffer            request.S3BufferConfig    `mapstructure:"s3-buffer,omitempty"`
	CqlStore            request.AstraBundleConfig `mapstructure:"cql-store,omitempty"`
	ResourceNamespace   string                    `mapstructure:"resource-namespace,omitempty"`
	KubeConfigPath      string                    `mapstructure:"kube-config-path,omitempty"`
	ShardKubeConfigPath string                    `mapstructure:"shard-kube-config-path,omitempty"`
}

func configExists(configPath string) (bool, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func LoadConfig(ctx context.Context) SchedulerConfig {
	logger := klog.FromContext(ctx)
	customViper := viper.NewWithOptions(viper.KeyDelimiter("__"))
	customViper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	localConfig := fmt.Sprintf("appconfig.%s.yaml", strings.ToLower(os.Getenv("APPLICATION_ENVIRONMENT")))

	if exists, err := configExists(localConfig); exists {
		customViper.SetConfigFile(fmt.Sprintf("appconfig.%s.yaml", strings.ToLower(os.Getenv("APPLICATION_ENVIRONMENT"))))
	} else if err != nil {
		logger.Error(err, "could not locate application configuration file")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	} else {
		customViper.SetConfigFile("appconfig.yaml")
	}

	customViper.SetEnvPrefix(EnvPrefix)
	customViper.AllowEmptyEnv(true)
	customViper.AutomaticEnv()

	if err := customViper.ReadInConfig(); err != nil {
		logger.Error(err, "error loading application config from appconfig.yaml")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	var appConfig SchedulerConfig
	err := customViper.Unmarshal(&appConfig)

	if err != nil {
		logger.Error(err, "error loading application config from appconfig.yaml")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	return appConfig
}
