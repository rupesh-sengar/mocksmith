package main

import (
	"log"
	"os"

	"mocksmith/internal/server"
	"mocksmith/internal/snapshot"
)

func main() {
	adminKey := os.Getenv("ADMIN_KEY")
	if adminKey == "" {
		adminKey = "dev"
	}

	initial := &snapshot.Snapshot{
		Project:   "demo",
		Env:       map[string]string{"market": "IN"},
		APIKeys:   map[string]struct{}{"DEMO_KEY": {}},
		RateLimit: 60, // rpm
	}

	s := server.New(initial, adminKey)
	log.Println("MockSmith Gin POC listening on :8787")
	if err := s.Run(":8787"); err != nil {
		log.Fatal(err)
	}
}
