package server

import (
	"github.com/gin-gonic/gin"

	"mocksmith/internal/db"
	"mocksmith/internal/server/routes"
	"mocksmith/internal/server/store"
	"mocksmith/internal/snapshot"
)

type Server struct {
	engine *gin.Engine
}

func New(initial *snapshot.Snapshot, adminKey string, repo *db.Repo) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	state := store.NewState(initial)
	routes.Register(engine, routes.Dependencies{
		State:    state,
		AdminKey: adminKey,
		Repo:     repo,
	})

	return &Server{engine: engine}
}

func (s *Server) Run(addr string) error { return s.engine.Run(addr) }
