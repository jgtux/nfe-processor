// @title           NF-e Processor API
// @version         1.0
// @description     API for receiving, processing and classifying NF-e (Nota Fiscal Eletrônica) XML files.
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/nfe-processor/backend/docs"
	"github.com/nfe-processor/backend/internal/api/handler"
	"github.com/nfe-processor/backend/internal/api/middleware"
	"github.com/nfe-processor/backend/internal/config"
	"github.com/nfe-processor/backend/internal/parser"
	"github.com/nfe-processor/backend/internal/queue"
	"github.com/nfe-processor/backend/internal/repository"
	"github.com/nfe-processor/backend/internal/service"
)

func main() {
	cfg := config.Load()

	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("resolve executable path: %v", err)
	}
	parser.SchemaDir = filepath.Join(filepath.Dir(execPath), "schemas")
	log.Printf("[parser] XSD schema dir: %s", parser.SchemaDir)

	repo, err := repository.New(&cfg.DB)
	if err != nil {
		log.Fatalf("repository: %v", err)
	}

	mq, err := queue.New(&cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	defer mq.Close()

	svc := service.New(repo, mq, cfg.Quarantine)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	h := handler.New(svc)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", h.Health)
		v1.POST("/xml/upload",
			middleware.RateLimiter(cfg.RateLimit.Capacity, cfg.RateLimit.Rate),
			h.UploadXML,
		)
		v1.GET("/clients", h.ListClients)

		nfe := v1.Group("/nfe")
		{
			nfe.GET("", h.ListNFes)
			nfe.GET("/summary", h.ClientSummary)
			nfe.GET("/unidentified", h.ListUnidentified)
			nfe.GET("/quarantine", h.Quarantine)
		}
	}

	log.Printf("[server] listening on :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
