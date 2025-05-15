package main

import (
	"context"
	"flag"
	nexusconf "github.com/SneaksAndData/nexus-core/pkg/configurations"
	"github.com/SneaksAndData/nexus-core/pkg/signals"
	"github.com/SneaksAndData/nexus-core/pkg/telemetry"
	v1 "github.com/SneaksAndData/nexus/api/v1"
	"github.com/SneaksAndData/nexus/app"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
	"os"
)

const (
	MaxBodySize = 512 * 1024 * 1024
)

var (
	logLevel string
)

func init() {
	flag.StringVar(&logLevel, "log-level", "INFO", "Log level for the application.")
}

func setupRouter(ctx context.Context) *gin.Engine {
	gin.DisableConsoleColor()
	router := gin.Default()
	router.MaxMultipartMemory = MaxBodySize
	router.Use(gin.Logger())
	// disable trusted proxies check
	_ = router.SetTrustedProxies(nil)
	// set runtime mode
	gin.SetMode(os.Getenv("GIN_MODE"))

	appConfig := nexusconf.LoadConfig[app.SchedulerConfig](ctx)

	appServices := (&app.ApplicationServices{}).
		WithKubeClients(ctx, appConfig.KubeConfigPath).
		WithBuffer(ctx, &appConfig.S3Buffer, &appConfig.CqlStore).
		WithCache(ctx, appConfig.ResourceNamespace).
		WithRecorder(ctx, appConfig.ResourceNamespace).
		WithShards(ctx, appConfig.ShardKubeConfigPath, appConfig.ResourceNamespace).
		WithDefaultNamespace(appConfig.ResourceNamespace)

	// version 1.2
	apiV12 := router.Group("algorithm/v1.2")

	apiV12.POST("run/:algorithmName", v1.CreateRun(appServices.CheckpointBuffer(), appServices.Cache()))
	apiV12.GET("results/:algorithmName/requests/:requestId", v1.GetRunResult(appServices.CheckpointBuffer()))

	go func() {
		appServices.Start(ctx)
		// handle exit
		logger := klog.FromContext(ctx)
		reason := ctx.Err()
		if reason.Error() == context.Canceled.Error() {
			logger.V(0).Info("received SIGTERM, shutting down gracefully")
			klog.FlushAndExit(klog.ExitFlushTimeout, 0)
		}

		logger.V(0).Error(reason, "fatal error occurred.")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}()

	return router
}

func main() {
	ctx := signals.SetupSignalHandler()
	appLogger, err := telemetry.ConfigureLogger(ctx, map[string]string{}, logLevel)
	ctx = telemetry.WithStatsd(ctx, "nexus")
	logger := klog.FromContext(ctx)

	if err != nil {
		logger.Error(err, "one of the logging handlers cannot be configured")
	}

	klog.SetSlogLogger(appLogger)

	r := setupRouter(ctx)
	// Configure webhost
	_ = r.Run(":8080")
}
