package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vhvplatform/go-auth-service/internal/gateway"
	"github.com/vhvplatform/go-shared/config"
	"github.com/vhvplatform/go-shared/jwt"
	"github.com/vhvplatform/go-shared/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting API Gateway", zap.String("environment", cfg.Environment))

	// Initialize JWT manager for internal token generation
	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.Expiration, cfg.JWT.RefreshExpiration)

	// Initialize local cache
	// In a real scenario, these values should come from config
	localCache := gateway.NewCache(5*time.Minute, 10*time.Minute)

	// Initialize Proxy
	proxy := gateway.NewProxy()
	// Add default services (these should eventually come from service discovery or config)
	proxy.AddService("auth-service", "http://localhost:8081")
	proxy.AddService("file-service", "http://localhost:8082")

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	// Initialize real AuthClient (gRPC) - placeholder for now
	// authClient := &gateway.AuthRPCClient{ ... }

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "gateway is healthy"})
	})

	// Add AuthMiddleware to router
	// For public routes, we don't apply it.

	apiGroup := router.Group("/api")
	{
		// Use manual check for login/register
		apiGroup.Any("/*path", func(c *gin.Context) {
			path := c.Param("path")
			if strings.Contains(path, "/auth/login") || strings.Contains(path, "/auth/register") {
				proxy.ServeHTTP(c.Writer, c.Request, "", "")
				return
			}

			// Apply AuthMiddleware inline (simplified)
			gateway.AuthMiddleware(nil, localCache, jwtManager, log)(c)
			if c.IsAborted() {
				return
			}

			tenantID, _ := c.Get("tenant_id")
			internalToken, _ := c.Get("internal_token")

			proxy.ServeHTTP(c.Writer, c.Request, tenantID.(string), internalToken.(string))
		})
	}

	// Other groups for /page and /upload
	router.Any("/page/*path", func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request, "", "")
	})
	router.Any("/upload/*path", func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request, "", "")
	})

	// Start server
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start Gateway", zap.Error(err))
		}
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Gateway forced to shutdown", zap.Error(err))
	}
}

// Note: Added missing import for strings
