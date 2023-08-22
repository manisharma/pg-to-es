package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pg-to-es/internal/config"
	"pg-to-es/internal/db"
	"pg-to-es/internal/model"
)

type Delta struct {
	Operation string
	Table     string
	Payload   json.RawMessage
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interruptStream := make(chan os.Signal)
	signal.Notify(interruptStream, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interruptStream)

	// Initialize Config
	cfg, err := config.Load("pipeline")
	if err != nil {
		log.Fatalln("config.Load() failed, err:", err.Error())
	}

	// Run Migrations (if any)
	err = db.Migrate(cfg.Pg)
	if err != nil {
		log.Fatalf("db.Migrate() failed, err: %s", err)
	}

	// Initiate Listening
	listener, err := db.NewListener(cfg.Pg)
	if err != nil {
		log.Fatalf("db.NewListener() failed, err: %s", err)
	}

	deltaStream, err := listener.Start(ctx)
	if err != nil {
		log.Fatalf("listener.StartListening() failed, err: %s", err)
	}
	defer listener.Stop()

	// Process delta
	go func() {
		for data := range deltaStream {
			var d Delta
			err := json.Unmarshal([]byte(data), &d)
			if err != nil {
				log.Printf("json.Unmarshal() failed, content: '%s', err: %s", data, err)
			} else {
				switch d.Table {
				case "users":
					var u model.User
					err = json.Unmarshal(d.Payload, &u)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
					}
					// TODO: sync to es
				case "projects":
					var p model.Project
					err = json.Unmarshal(d.Payload, &p)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
					}
					// TODO: sync to es
				case "hashtags":
					var h model.Hashtag
					err = json.Unmarshal(d.Payload, &h)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
					}
					// TODO: sync to es
				}
			}
		}
	}()

	<-interruptStream
	log.Println("pipeline interrupted!")

}
