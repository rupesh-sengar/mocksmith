package controllers

import (
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"mocksmith/internal/server/services"
	"mocksmith/internal/server/store"
)

type RuntimeController struct {
	service *services.RuntimeService
}

func NewRuntime(state *store.State) *RuntimeController {
	return &RuntimeController{service: services.NewRuntimeService(state)}
}

func (r *RuntimeController) Handle(c *gin.Context) {
	start := time.Now()

	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
	}

	res := r.service.Handle(services.RuntimeRequest{
		Method:  c.Request.Method,
		Path:    c.Param("path"),
		Request: c.Request,
		Body:    bodyBytes,
	})

	for k, v := range res.Headers {
		c.Header(k, v)
	}

	c.Data(res.Status, "application/json", res.Body)

	if res.ShouldLog {
		dur := time.Since(start).Milliseconds()
		log.Printf("%s %s -> %d [%s] %dms\n", c.Request.Method, res.RequestPath, res.Status, res.Scenario, dur)
	}
}
