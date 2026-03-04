package controllers

import (
	"io"

	"github.com/gin-gonic/gin"

	"mocksmith/internal/db"
	"mocksmith/internal/server/services"
	"mocksmith/internal/server/store"
)

type AdminController struct {
	service         *services.AdminService
	projectsService *services.ProjectsService
}

func NewAdmin(state *store.State, repo *db.Repo) *AdminController {
	return &AdminController{service: services.NewAdminService(state), projectsService: services.NewProjectsService(repo)}
}

func (a *AdminController) Import(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "read body", "details": err.Error()})
		return
	}

	summary, err := a.service.Import(body, c.GetHeader("content-type"))
	if err != nil {
		if impErr, ok := err.(*services.ImportError); ok {
			c.JSON(400, gin.H{"error": impErr.Code, "details": impErr.Details})
			return
		}
		c.JSON(400, gin.H{"error": "import failed", "details": err.Error()})
		return
	}

	c.JSON(200, summary)
}

func (a *AdminController) Routes(c *gin.Context) {
	c.JSON(200, a.service.Routes())
}

func (a *AdminController) Config(c *gin.Context) {
	c.JSON(200, a.service.Config())
}

func (a *AdminController) OpenAPI(c *gin.Context) {
	c.JSON(200, a.service.OpenAPI())
}

func (a *AdminController) Logs(c *gin.Context) {
	c.JSON(200, gin.H{"logs": "stdout (POC)"})
}
