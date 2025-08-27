package main

import (
	"context"
	"errors"
	nexusconf "github.com/SneaksAndData/nexus-core/pkg/configurations"
	"github.com/SneaksAndData/nexus-core/pkg/signals"
	"github.com/SneaksAndData/nexus-core/pkg/telemetry"
	v1 "github.com/SneaksAndData/nexus/api/v1"
	"github.com/SneaksAndData/nexus/app"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
	"os"
)

func setupRouter(ctx context.Context, appConfig *app.SchedulerConfig) *gin.Engine {
	gin.DisableConsoleColor()
	router := gin.Default()
	router.MaxMultipartMemory = appConfig.MaxPayloadSizeBytes()
	router.Use(gin.Logger())
	// disable trusted proxies check
	_ = router.SetTrustedProxies(nil)
	// set runtime mode
	gin.SetMode(os.Getenv("GIN_MODE"))

	appServices := (&app.ApplicationServices{}).
		WithKubeClients(ctx, appConfig.KubeConfigPath)

	switch appConfig.CqlStoreType {
	case app.CqlStoreAstra:
		appServices = appServices.WithAstraS3Buffer(ctx, &appConfig.S3Buffer, &appConfig.AstraCqlStore)
	case app.CqlStoreScylla:
		appServices = appServices.WithScyllaS3Buffer(ctx, &appConfig.S3Buffer, &appConfig.ScyllaCqlStore)
	default:
		klog.FromContext(ctx).Error(errors.New("unknown store type "+appConfig.CqlStoreType), "failed to initialize a CqlStore")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	appServices = appServices.
		WithCache(ctx, appConfig.ResourceNamespace).
		WithRecorder(ctx, appConfig.ResourceNamespace).
		WithShards(ctx, appConfig.ShardKubeConfigPath, appConfig.ResourceNamespace).
		WithDefaultNamespace(appConfig.ResourceNamespace).
		BuildScheduler(ctx)

	// version 1
	apiV1 := router.Group("algorithm/v1")

	apiV1.POST("run/:algorithmName", v1.CreateRun(appServices.CheckpointBuffer(), appServices.Cache(), appServices.Scheduler(), appServices.Logger(ctx)))
	apiV1.POST("cancel/:algorithmName/requests/:requestId", v1.CancelRun(appServices.Scheduler(), appServices.Logger(ctx)))
	apiV1.GET("results/:algorithmName/requests/:requestId", v1.GetRunResult(appServices.CheckpointBuffer()))
	apiV1.GET("results/tags/:requestTag", v1.GetRunResultsByTag(appServices.CheckpointBuffer(), appServices.Logger(ctx)))
	apiV1.GET("metadata/:algorithmName/requests/:requestId", v1.GetRunMetadata(appServices.CheckpointBuffer()))
	apiV1.GET("buffer/:algorithmName/requests/:requestId", v1.GetBufferedRunMetadata(appServices.CheckpointBuffer()))
	apiV1.GET("payload/:algorithmName/requests/:requestId", v1.GetRunPayload(appServices.CheckpointBuffer()))

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

// @title           Nexus Scheduler API
// @version         1.0
// @description     Nexus Scheduler API specification. All Nexus supported clients conform to this spec.

// @contact.name   ESD Support
// @contact.email  esdsupport@ecco.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath  /algorithm/v1
func main() {
	ctx := signals.SetupSignalHandler()
	appConfig := nexusconf.LoadConfig[app.SchedulerConfig](ctx)
	appLogger, err := telemetry.ConfigureLogger(ctx, map[string]string{}, appConfig.LogLevel)
	ctx = telemetry.WithStatsd(ctx, "nexus")
	logger := klog.FromContext(ctx)

	if err != nil {
		logger.V(0).Error(err, "one of the logging handlers cannot be configured")
	}

	klog.SetSlogLogger(appLogger)

	r := setupRouter(ctx, &appConfig)

	// Configure webhost
	_ = r.Run(":8080")
}
