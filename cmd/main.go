package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/longvhv/saas-framework-go/pkg/config"
	"github.com/longvhv/saas-framework-go/pkg/jwt"
	"github.com/longvhv/saas-framework-go/pkg/logger"
	"github.com/longvhv/saas-framework-go/pkg/mongodb"
	"github.com/longvhv/saas-framework-go/pkg/redis"
	"github.com/longvhv/saas-framework-go/services/auth-service/internal/grpc"
	"github.com/longvhv/saas-framework-go/services/auth-service/internal/handler"
	"github.com/longvhv/saas-framework-go/services/auth-service/internal/repository"
	"github.com/longvhv/saas-framework-go/services/auth-service/internal/service"
	// pb "github.com/longvhv/saas-framework-go/services/auth-service/proto"
	"go.uber.org/zap"
	grpcServer "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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

	log.Info("Starting Auth Service", zap.String("environment", cfg.Environment))

	// Initialize MongoDB
	mongoClient, err := mongodb.NewClient(context.Background(), mongodb.Config{
		URI:         cfg.MongoDB.URI,
		Database:    cfg.MongoDB.Database,
		MaxPoolSize: cfg.MongoDB.MaxPoolSize,
		MinPoolSize: cfg.MongoDB.MinPoolSize,
	})
	if err != nil {
		log.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer mongoClient.Close(context.Background())

	// Initialize Redis
	redisClient, err := redis.NewClient(redis.Config{
		Addr:     cfg.Redis.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize JWT manager
	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.Expiration, cfg.JWT.RefreshExpiration)

	// Initialize repositories
	userRepo := repository.NewUserRepository(mongoClient.Database())
	refreshTokenRepo := repository.NewRefreshTokenRepository(mongoClient.Database())
	roleRepo := repository.NewRoleRepository(mongoClient.Database())

	// Initialize services
	authService := service.NewAuthService(userRepo, refreshTokenRepo, roleRepo, jwtManager, redisClient, log)

	// Start gRPC server
	grpcPort := os.Getenv("AUTH_SERVICE_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	go startGRPCServer(authService, log, grpcPort)

	// Start HTTP server
	httpPort := os.Getenv("AUTH_SERVICE_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}
	startHTTPServer(authService, log, httpPort)
}

func startGRPCServer(authService *service.AuthService, log *logger.Logger, port string) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	grpcSrv := grpcServer.NewServer()
	authGrpcServer := grpc.NewAuthServiceServer(authService, log)
	// pb.RegisterAuthServiceServer(grpcSrv, authGrpcServer)
	_ = authGrpcServer // Use the variable to avoid unused error

	// Register health check service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcSrv, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	log.Info("gRPC server listening", zap.String("port", port))
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatal("Failed to serve gRPC", zap.Error(err))
	}
}

func startHTTPServer(authService *service.AuthService, log *logger.Logger, port string) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, log)

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.RefreshToken)
		}
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Info("HTTP server listening", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}
