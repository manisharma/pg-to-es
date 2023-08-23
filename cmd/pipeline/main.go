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
		type delta struct {
			Operation string          `json:"operation"`
			Table     string          `json:"table"`
			Payload   json.RawMessage `json:"payload"`
		}
		type document struct {
			UserID             int    `json:"user_id"`
			UserName           string `json:"user_name"`
			UserCreatedAt      string `json:"user_created_at"`
			ProjectID          int    `json:"project_id"`
			ProjectName        string `json:"project_name"`
			ProjectSlug        string `json:"project_slug"`
			ProjectDescription string `json:"project_description"`
			ProjectCreatedAt   string `json:"project_created_at"`
			HashtagID          int    `json:"hashtag_id"`
			HashtagName        string `json:"hashtag_name"`
			HashtagCreatedAt   string `json:"hashtag_created_at"`
			Operation          string `json:"operation"`
			Table              string `json:"table"`
		}
		for data := range deltaStream {
			log.Println("data", data)
			var d delta
			err := json.Unmarshal([]byte(data), &d)
			if err != nil {
				log.Printf("json.Unmarshal() failed, content: '%s', err: %s", data, err)
				continue
			}

			switch d.Operation {
			case "INSERT", "UPDATE":
				var u document
				err = json.Unmarshal(d.Payload, &u)
				if err != nil {
					log.Printf("\njson.Unmarshal(d.Payload, &user), err: %s", err)
					continue
				}
				log.Println(d.Operation)
				log.Println(u)
			case "DELETE":
				switch d.Table {
				case "users":
					var u model.User
					err = json.Unmarshal(d.Payload, &u)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
						continue
					}
					log.Println(d.Operation)
					log.Println(u)
					// TODO: sync to es
				case "projects":
					var p model.Project
					err = json.Unmarshal(d.Payload, &p)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
						continue
					}
					log.Println(d.Operation)
					log.Println(p)
					// TODO: sync to es
				case "hashtags":
					var h model.Hashtag
					err = json.Unmarshal(d.Payload, &h)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
						continue
					}
					log.Println(d.Operation)
					log.Println(h)
					// TODO: sync to es
				case "project_hashtags":
					var h model.ProjectHashtag
					err = json.Unmarshal(d.Payload, &h)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
						continue
					}
					log.Println(d.Operation)
					log.Println(h)
				case "user_projects":
					var h model.UserProject
					err = json.Unmarshal(d.Payload, &h)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
						continue
					}
					log.Println(d.Operation)
					log.Println(h)
				}
			}
		}
	}()

	<-interruptStream
	log.Println("pipeline interrupted!")

}
