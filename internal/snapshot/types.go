package snapshot

import (
	"net/http"
	"net/url"
	"regexp"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"mocksmith/internal/config"
)

type Snapshot struct {
	Project   string
	Env       map[string]string
	APIKeys   map[string]struct{}
	RateLimit int // rpm
	Routes    []CompiledRoute
}

type CompiledRoute struct {
	Method       string
	PathTemplate string
	PathRE       *regexp.Regexp
	ParamNames   []string
	Strict       bool
	ReqSchema    *jsonschema.Schema
	Scenarios    []CompiledScenario
	Default      *CompiledScenario
}

type CompiledScenario struct {
	Name         string
	Priority     int
	Match        *config.MatchRules
	Status       int
	Headers      map[string]string
	BodyBytes    []byte
	BodyIsBase64 bool
	LatencyMS    int
	ErrorRatePct int
	Weight       int
}

type MatchContext struct {
	Query   url.Values
	Headers http.Header
	Body    map[string]any
}
