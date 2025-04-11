package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"process-payments/internal/config"
	"process-payments/internal/middlewares"
	"process-payments/internal/routes"
	"syscall"
	"time"

	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

func StartServer(cfg *config.Config) {
	log.Println("Starting server...")
	gin.SetMode(gin.ReleaseMode)
	if !cfg.Production {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Security middleware to handle X-Forwarded headers
	secureMiddleware := secure.New(secure.Config{
		SSLRedirect:           false,
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:            31536000,
		STSIncludeSubdomains:  true,
		STSPreload:            true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
	})
	router.Use(secureMiddleware)

	router.Use(gin.Recovery())

	// Logging, security, metrics and utils middleware
	router.Use(middlewares.CustomLogger())
	router.Use(middlewares.CORSMiddleware(cfg.ClientURL, cfg.Production))

	api := router.Group("/api")
	{
		routes.StripeRoutes(api.Group("/stripe"))
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
