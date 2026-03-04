package main

import (
	"context"
	"log"
	"time"

	"mocksmith/internal/appconfig"
	"mocksmith/internal/db"
	"mocksmith/internal/server"
	"mocksmith/internal/snapshot"
)

func main() {
	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	initial := &snapshot.Snapshot{
		Project:   "demo",
		Env:       map[string]string{"market": "IN"},
		APIKeys:   map[string]struct{}{"DEMO_KEY": {}},
		RateLimit: 60, // rpm
	}

	var repo *db.Repo
	if cfg.DatabaseURL == "" {
		log.Println("DATABASE_URL is empty; /admin/projects/:id will return 503")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pool, err := db.NewPool(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("database connection failed: %v", err)
		}
		defer pool.Close()

		repo = db.NewRepo(pool)
		log.Println("Database connected")
	}

	s := server.New(initial, cfg.AdminKey, repo)
	log.Printf("MockSmith Gin POC listening on %s", cfg.ListenAddr)
	if err := s.Run(cfg.ListenAddr); err != nil {
		log.Fatal(err)
	}
}
