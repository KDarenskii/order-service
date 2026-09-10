package main

import (
	"context"
	"log"

	"github.com/KDarenskii/order-service/internal/app/config"
	rhealth "github.com/KDarenskii/order-service/internal/app/handler/http/health"
	rprocessor "github.com/KDarenskii/order-service/internal/app/processor/http"
	rcpostgres "github.com/KDarenskii/order-service/internal/app/repository/conn/postgres"
)

func main() {
	config.Load()

	cfg := config.Root

	ctx := context.Background()

	pgClient, err := rcpostgres.NewClient(ctx, cfg.Repository.Postgres)
	if err != nil {
		log.Fatalf("Failed to create PostgreSQL client: %v", err)
	}

	defer func() {
		if err := pgClient.Close(); err != nil {
			log.Printf("%s", err.Error())
		}
	}()

	log.Println("Connection to PostgreSQL client has been established")

	hHealth := rhealth.NewHandler()

	httpServer := rprocessor.NewHTTP(hHealth, cfg.Processor.WebServer)

	if err := httpServer.Serve(); err != nil {
		log.Printf("HTTP server failed: %v", err)
	}
}
