package services

import (
	"encoding/json"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"mocksmith/internal/compile"
	"mocksmith/internal/config"
	"mocksmith/internal/server/store"
)

type AdminService struct {
	state *store.State
}

type ImportSummary struct {
	Project   string `json:"project"`
	Routes    int    `json:"routes"`
	Scenarios int    `json:"scenarios"`
}

type ImportError struct {
	Code    string
	Details string
}

func (e *ImportError) Error() string { return e.Code }

type RouteItem struct {
	Method    string   `json:"method"`
	Path      string   `json:"path"`
	Strict    bool     `json:"strict"`
	Scenarios []string `json:"scenarios"`
}

type RoutesResponse struct {
	Project string      `json:"project"`
	Routes  []RouteItem `json:"routes"`
}

type AuthInfo struct {
	Header string   `json:"header"`
	Keys   []string `json:"keys,omitempty"`
}

type ConfigResponse struct {
	Project      string            `json:"project"`
	Env          map[string]string `json:"env"`
	RateLimitRPM int               `json:"rate_limit_rpm"`
	RuntimeAuth  AuthInfo          `json:"runtime_auth"`
	AdminAuth    AuthInfo          `json:"admin_auth"`
}

type OpenAPIInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type OpenAPIResponse struct {
	OpenAPI string      `json:"openapi"`
	Info    OpenAPIInfo `json:"info"`
	XRoutes int         `json:"x-routes"`
}

func NewAdminService(state *store.State) *AdminService {
	return &AdminService{state: state}
}

func (a *AdminService) Import(body []byte, contentType string) (ImportSummary, error) {
	var cfg config.Config
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "yaml"), strings.Contains(ct, "yml"):
		if err := yaml.Unmarshal(body, &cfg); err != nil {
			return ImportSummary{}, &ImportError{Code: "parse yaml", Details: err.Error()}
		}
	default:
		if err := json.Unmarshal(body, &cfg); err != nil {
			if err := yaml.Unmarshal(body, &cfg); err != nil {
				return ImportSummary{}, &ImportError{Code: "parse json/yaml", Details: err.Error()}
			}
		}
	}

	snap, err := compile.Build(&cfg)
	if err != nil {
		return ImportSummary{}, &ImportError{Code: "compile", Details: err.Error()}
	}
	a.state.StoreSnapshot(snap)

	scenCount := 0
	for _, r := range snap.Routes {
		scenCount += len(r.Scenarios)
	}

	return ImportSummary{
		Project:   snap.Project,
		Routes:    len(snap.Routes),
		Scenarios: scenCount,
	}, nil
}

func (a *AdminService) Routes() RoutesResponse {
	snap := a.state.Snapshot()
	out := make([]RouteItem, 0, len(snap.Routes))
	for _, r := range snap.Routes {
		names := make([]string, 0, len(r.Scenarios))
		for _, sc := range r.Scenarios {
			names = append(names, sc.Name)
		}
		out = append(out, RouteItem{Method: r.Method, Path: r.PathTemplate, Strict: r.Strict, Scenarios: names})
	}
	return RoutesResponse{Project: snap.Project, Routes: out}
}

func (a *AdminService) Config() ConfigResponse {
	snap := a.state.Snapshot()
	keys := make([]string, 0, len(snap.APIKeys))
	for key := range snap.APIKeys {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return ConfigResponse{
		Project:      snap.Project,
		Env:          snap.Env,
		RateLimitRPM: snap.RateLimit,
		RuntimeAuth:  AuthInfo{Header: "x-mocksmith-key", Keys: keys},
		AdminAuth:    AuthInfo{Header: "x-admin-key"},
	}
}

func (a *AdminService) OpenAPI() OpenAPIResponse {
	snap := a.state.Snapshot()
	return OpenAPIResponse{
		OpenAPI: "3.0.3",
		Info:    OpenAPIInfo{Title: "MockSmith POC", Version: "0.0.1"},
		XRoutes: len(snap.Routes),
	}
}
