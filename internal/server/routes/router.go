package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"mocksmith/internal/db"
	"mocksmith/internal/server/controllers"
	"mocksmith/internal/server/middleware"
	"mocksmith/internal/server/store"
)

type Dependencies struct {
	State    *store.State
	AdminKey string
	Repo     *db.Repo
}

func Register(r *gin.Engine, deps Dependencies) {
	r.Use(gin.Recovery(), middleware.RequestID())

	corsCfg := cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "x-admin-key", "x-mocksmith-key"},
		ExposeHeaders:    []string{"Content-Length", "x-request-id", "x-request-ts"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	r.Use(cors.New(corsCfg))

	admin := controllers.NewAdmin(deps.State, deps.Repo)
	runtime := controllers.NewRuntime(deps.State)

	adm := r.Group("/admin")
	adm.Use(middleware.RequireAdmin(deps.AdminKey))
	{
		adm.GET("/config", admin.Config)
		adm.POST("/import", admin.Import)
		adm.GET("/routes", admin.Routes)
		adm.GET("/openapi", admin.OpenAPI)
		adm.GET("/logs", admin.Logs)
		adm.GET("/projects/:id", admin.GetProject)
		adm.POST("/projects", admin.CreateProject)
		adm.GET("/projects", admin.GetProjects)
	}

	rt := r.Group("/mock/:project/:env")
	rt.Use(middleware.RequireRuntimeKey(deps.State))
	{
		rt.Any("/*path", runtime.Handle)
	}
}
