package main

import (
	"context"
	"flag"
	"github.com/SneaksAndData/nexus-core/pkg/signals"
	"github.com/SneaksAndData/nexus-core/pkg/telemetry"
	v1 "github.com/SneaksAndData/nexus/api/v1"
	"github.com/SneaksAndData/nexus/app"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
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
	// Disable Console Color
	// gin.DisableConsoleColor()
	router := gin.Default()
	router.MaxMultipartMemory = MaxBodySize
	router.Use(gin.Logger())
	appConfig := app.LoadConfig(ctx)

	appServices := (&app.ApplicationServices{}).
		WithKubeClients(ctx, appConfig.KubeConfigPath).
		WithBuffer(ctx, &appConfig.Buffer, &appConfig.CqlStore).
		WithCache(ctx, appConfig.ResourceNamespace).
		WithRecorder(ctx, appConfig.ResourceNamespace)

	appServices.Start(ctx)

	// version 1.2
	apiV12 := router.Group("algorithm/v1.2")

	apiV12.POST("run/:algorithmName", v1.CreateRun(appServices.CheckpointBuffer(), appServices.Cache()))

	// TODO: Boxer auth middleware

	//// Ping test
	//r.GET("/ping", func(c *gin.Context) {
	//	c.String(http.StatusOK, "pong")
	//})
	//
	//// Get user value
	//r.GET("/user/:name", func(c *gin.Context) {
	//	user := c.Params.ByName("name")
	//	value, ok := db[user]
	//	if ok {
	//		c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
	//	} else {
	//		c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
	//	}
	//})
	//
	//// Authorized group (uses gin.BasicAuth() middleware)
	//// Same than:
	//// authorized := r.Group("/")
	//// authorized.Use(gin.BasicAuth(gin.Credentials{
	////	  "foo":  "bar",
	////	  "manu": "123",
	////}))
	//authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
	//	"foo":  "bar", // user:foo password:bar
	//	"manu": "123", // user:manu password:123
	//}))
	//
	///* example curl for /admin with basicauth header
	//   Zm9vOmJhcg== is base64("foo:bar")
	//
	//	curl -X POST \
	//  	http://localhost:8080/admin \
	//  	-H 'authorization: Basic Zm9vOmJhcg==' \
	//  	-H 'content-type: application/json' \
	//  	-d '{"value":"bar"}'
	//*/
	//authorized.POST("admin", func(c *gin.Context) {
	//	user := c.MustGet(gin.AuthUserKey).(string)
	//
	//	// Parse JSON
	//	var json struct {
	//		Value string `json:"value" binding:"required"`
	//	}
	//
	//	if c.Bind(&json) == nil {
	//		db[user] = json.Value
	//		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	//	}
	//})

	return router
}

func main() {
	ctx := signals.SetupSignalHandler()
	appLogger, err := telemetry.ConfigureLogger(ctx, map[string]string{}, logLevel)
	ctx = telemetry.WithStatsd(ctx, "nexus")
	logger := klog.FromContext(ctx)

	if err != nil {
		logger.Error(err, "One of the logging handlers cannot be configured")
	}

	klog.SetSlogLogger(appLogger)

	r := setupRouter(ctx)
	// Configure webhost
	_ = r.Run(":8080")
}
