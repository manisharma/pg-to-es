package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pg-to-es/internal/business"
	"pg-to-es/internal/config"
	"pg-to-es/internal/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interruptStream := make(chan os.Signal, 1)
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

	// Initialize & run server
	server := business.NewServer(esSvc, cfg.Server.Port, cfg.Es.Index)
	server.InitRoutes()
	go func() {
		log.Printf("server listening on :%d", cfg.Server.Port)
		err = server.Start()
		if err != nil {
			log.Fatalf("srv.Start() failed, err: %s", err)
		}
	}()

	// Await interruptions
	<-interruptStream
	log.Println("server interrupted!")
	server.Shutdown(ctx)
}
