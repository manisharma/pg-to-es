package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pg-to-es/internal/business"
	"pg-to-es/internal/config"
	"pg-to-es/internal/db"
	"pg-to-es/internal/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interruptStream := make(chan os.Signal)
	signal.Notify(interruptStream, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interruptStream)

	// Initialize Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalln("config.Load() failed, err:", err.Error())
	}

	// Initialize Elasticsearch Service
	esSvc, err := service.NewElastic(cfg.Es)
	if err != nil {
		log.Fatalf("elasticsearch.New() failed, err: %s", err)
	}

	// Run Migrations
	err = db.Migrate(cfg.Pg)
	if err != nil {
		log.Fatalf("b.Migrate() failed, err: %s", err)
	}

	// Initiate DB Listener Service
	dbListenerSvc, err := service.NewDbListener(cfg.Pg)
	if err != nil {
		log.Fatalf("db.NewListener() failed, err: %s", err)
	}

	// Initialize & run pipeline
	psToEsPipeline := business.NewPipeline(dbListenerSvc, esSvc, cfg.Es.Index)
	psToEsPipeline.Start(ctx)
	defer psToEsPipeline.Stop()

	log.Println("pipeline syncing")
	// Await interruptions
	<-interruptStream
	log.Println("pipeline interrupted!")
}
