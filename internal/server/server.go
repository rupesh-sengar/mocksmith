package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"fmt"
	"github.com/gin-contrib/cors"
	"mocksmith/internal/compile"
	"mocksmith/internal/config"
	"mocksmith/internal/ratelimit"
	"mocksmith/internal/rules"
	"mocksmith/internal/snapshot"
	"mocksmith/internal/templating"
)

type Server struct {
	engine   *gin.Engine
	adminKey string
	snap     atomic.Value // *snapshot.Snapshot
	limits   *ratelimit.Registry
}

func New(initial *snapshot.Snapshot, adminKey string) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), requestID())
	
	config := cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "x-admin-key", "x-mocksmith-key"},
		ExposeHeaders:    []string{"Content-Length", "x-request-id", "x-request-ts"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	r.Use(cors.New(config))

	s := &Server{engine: r, adminKey: adminKey, limits: ratelimit.New()}
	s.snap.Store(initial)

	adm := r.Group("/admin")
	adm.Use(s.requireAdmin())
	{
		adm.POST("/import", s.handleImport)
		adm.GET("/routes", s.handleRoutes)
		adm.GET("/openapi", s.handleOpenAPI) // stub
		adm.GET("/logs", s.handleLogs)       // stdout for POC
	}

	rt := r.Group("/mock/:project/:env")
	rt.Use(s.requireRuntimeKey())
	{
		rt.Any("/*path", s.handleRuntime)
	}

	return s
}

func (s *Server) Run(addr string) error { return s.engine.Run(addr) }

func (s *Server) requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("x-admin-key") != s.adminKey {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Server) requireRuntimeKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("x-mocksmith-key")
		snap := s.snap.Load().(*snapshot.Snapshot)
		if _, ok := snap.APIKeys[key]; !ok {
			c.JSON(401, gin.H{"error": "invalid api key"})
			c.Abort()
			return
		}
		if snap.RateLimit > 0 {
			lim := s.limits.Get(key, snap.RateLimit)
			if !lim.Allow() {
				c.Header("Retry-After", "60")
				c.JSON(429, gin.H{"error": "rate limit"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

/* ===== Admin Handlers ===== */

func (s *Server) handleImport(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "read body", "details": err.Error()})
		return
	}
	var cfg config.Config
	ct := strings.ToLower(c.GetHeader("content-type"))
	switch {
	case strings.Contains(ct, "yaml"), strings.Contains(ct, "yml"):
		if err := yaml.Unmarshal(body, &cfg); err != nil {
			c.JSON(400, gin.H{"error": "parse yaml", "details": err.Error()})
			return
		}
	default:
		if err := json.Unmarshal(body, &cfg); err != nil {
			if err := yaml.Unmarshal(body, &cfg); err != nil {
				c.JSON(400, gin.H{"error": "parse json/yaml", "details": err.Error()})
				return
			}
		}
	}

	snap, err := compile.Build(&cfg)
	if err != nil {
		c.JSON(400, gin.H{"error": "compile", "details": err.Error()})
		return
	}
	s.snap.Store(snap)

	// summary
	var scenCount int
	for _, r := range snap.Routes {
		scenCount += len(r.Scenarios)
	}
	c.JSON(200, gin.H{"project": snap.Project, "routes": len(snap.Routes), "scenarios": scenCount})
}

func (s *Server) handleRoutes(c *gin.Context) {
	snap := s.snap.Load().(*snapshot.Snapshot)
	type item struct {
		Method    string   `json:"method"`
		Path      string   `json:"path"`
		Strict    bool     `json:"strict"`
		Scenarios []string `json:"scenarios"`
	}
	out := []item{}
	for _, r := range snap.Routes {
		names := make([]string, 0, len(r.Scenarios))
		for _, sc := range r.Scenarios {
			names = append(names, sc.Name)
		}
		out = append(out, item{Method: r.Method, Path: r.PathTemplate, Strict: r.Strict, Scenarios: names})
	}
	c.JSON(200, gin.H{"project": snap.Project, "routes": out})
}

func (s *Server) handleOpenAPI(c *gin.Context) {
	snap := s.snap.Load().(*snapshot.Snapshot)
	c.JSON(200, gin.H{
		"openapi": "3.0.3",
		"info":    gin.H{"title": "MockSmith POC", "version": "0.0.1"},
		"x-routes": len(snap.Routes),
	})
}

func (s *Server) handleLogs(c *gin.Context) {
	c.JSON(200, gin.H{"logs": "stdout (POC)"})
}

/* ===== Runtime Handler ===== */

func (s *Server) handleRuntime(c *gin.Context) {
	start := time.Now()
	snap := s.snap.Load().(*snapshot.Snapshot)

	requestPath := "/" + strings.TrimPrefix(c.Param("path"), "/")
	method := strings.ToUpper(c.Request.Method)

	// match route
	var rt *snapshot.CompiledRoute
	var params map[string]string
	for i := range snap.Routes {
		cr := &snap.Routes[i]
		if cr.Method != method {
			continue
		}
		if m := cr.PathRE.FindStringSubmatch(requestPath); m != nil {
			params = map[string]string{}
			for i, name := range cr.ParamNames {
				params[name] = m[i+1]
			}
			rt = cr
			break
		}
	}
	if rt == nil {
		c.JSON(404, gin.H{"error": "route not found", "path": requestPath})
		return
	}

	// read body JSON (once)
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
	}
	var bodyMap map[string]any
	_ = json.Unmarshal(bodyBytes, &bodyMap)

	// validate (if strict & schema present)
	if rt.Strict && rt.ReqSchema != nil {
		if err := rt.ReqSchema.Validate(bytes.NewReader(bodyBytes)); err != nil {
			c.JSON(400, gin.H{"error": "validation failed", "details": err.Error()})
			return
		}
	}

	// scenario selection
	ctx := snapshot.MatchContext{Query: c.Request.URL.Query(), Headers: c.Request.Header, Body: bodyMap}
	scen := rules.PickScenario(rt, ctx)

	// latency / chaos
	if scen.LatencyMS > 0 {
		time.Sleep(time.Duration(scen.LatencyMS) * time.Millisecond)
	}
	if scen.ErrorRatePct > 0 && roll(scen.ErrorRatePct) {
		c.JSON(500, gin.H{"error": "injected error"})
		return
	}

	// headers
	for k, v := range scen.Headers {
		c.Header(k, templating.RenderTokens(v, params, c.Request, snap.Env, bodyMap))
	}

	// status + body
	status := scen.Status
	if status == 0 {
		status = 200
	}
	rendered := templating.RenderTokens(string(scen.BodyBytes), params, c.Request, snap.Env, bodyMap)
	c.Data(status, "application/json", []byte(rendered))

	dur := time.Since(start).Milliseconds()
	log.Printf("%s %s -> %d [%s] %dms\n", method, requestPath, status, scen.Name, dur)
}

func roll(pct int) bool {
	if pct <= 0 {
		return false
	}
	if pct >= 100 {
		return true
	}
	// simple LCG
	var seed = time.Now().UnixNano()
	seed = (1103515245*seed + 12345) & 0x7fffffff
	return int(seed%100) < pct
}

/* ===== tiny middleware ===== */

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ts := time.Now().UnixNano()
		id := regexp.MustCompile(`\D`).ReplaceAllString(time.Now().Format("150405.000000000"), "") + "-" + strings.TrimSpace(strings.ReplaceAll(strings.TrimLeft(strings.TrimPrefix(c.ClientIP(), "::ffff:"), ":"), ":", ""))
		c.Writer.Header().Set("x-request-id", id)
		c.Writer.Header().Set("x-request-ts", fmt.Sprintf("%d", ts))
		c.Next()
	}
}
