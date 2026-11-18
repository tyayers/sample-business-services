package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tyayers/sample-business-services/internal/domain"
	"github.com/tyayers/sample-business-services/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := domain.NewStore()
	handler := server.New(store)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Sample Business Service listening on http://localhost:%s", port)
		log.Printf("  Explorer UI     : http://localhost:%s/", port)
		log.Printf("  Swagger UI      : http://localhost:%s/docs", port)
		log.Printf("  OpenAPI 3.2.0   : http://localhost:%s/openapi.json", port)
		log.Printf("  MCP Endpoint    : http://localhost:%s/mcp", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
