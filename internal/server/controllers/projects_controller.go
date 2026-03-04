package controllers

import (
	"errors"

	"github.com/gin-gonic/gin"

	"mocksmith/internal/server/services"
)

func (a *AdminController) GetProject(c *gin.Context) {
	id := c.Param("id")
	project, err := a.projectsService.GetProjectById(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrProjectNotFound):
			c.JSON(404, gin.H{"error": "project not found"})
		case errors.Is(err, services.ErrProjectsStoreUnavailable):
			c.JSON(503, gin.H{"error": "projects store unavailable"})
		default:
			c.JSON(500, gin.H{"error": "get project failed"})
		}
		return
	}
	c.JSON(200, project)
}

func (a *AdminController) CreateProject(c *gin.Context) {
	var payload services.CreateProjectPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": "invalid payload"})
		return
	}
	project, err := a.projectsService.CreateProject(c.Request.Context(), payload)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrProjectsStoreUnavailable):
			c.JSON(503, gin.H{"error": "projects store unavailable"})
		default:
			c.JSON(500, gin.H{"error": "create project failed"})
		}
		return
	}
	c.JSON(200, project)
}

func (a *AdminController) GetProjects(c *gin.Context) {
	projects, err := a.projectsService.GetProjects(c.Request.Context())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrProjectsStoreUnavailable):
			c.JSON(503, gin.H{"error": "projects store unavailable"})
		default:
			c.JSON(500, gin.H{"error": "get projects failed"})
		}
		return
	}
	c.JSON(200, projects)
}
