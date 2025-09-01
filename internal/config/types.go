package config

type Config struct {
	Project   string            `yaml:"project" json:"project"`
	BaseURL   string            `yaml:"base_url" json:"base_url"`
	APIKey    string            `yaml:"api_key" json:"api_key"`
	RateLimit *RateLimitCfg     `yaml:"rate_limit" json:"rate_limit"`
	Env       map[string]string `yaml:"env" json:"env"`
	Routes    []RouteCfg        `yaml:"routes" json:"routes"`
}

type RateLimitCfg struct {
	RequestsPerMinute int `yaml:"requests_per_minute" json:"requests_per_minute"`
}

type RouteCfg struct {
	Method       string        `yaml:"method" json:"method"`
	PathTemplate string        `yaml:"path_template" json:"path_template"`
	Strict       bool          `yaml:"strict" json:"strict"`
	Schemas      *SchemasCfg   `yaml:"schemas" json:"schemas"`
	Scenarios    []ScenarioCfg `yaml:"scenarios" json:"scenarios"`
	Default      *ScenarioCfg  `yaml:"default" json:"default"`
}

type SchemasCfg struct {
	RequestJSONSchema  string `yaml:"request_jsonschema"  json:"request_jsonschema"`
	ResponseJSONSchema string `yaml:"response_jsonschema" json:"response_jsonschema"`
}

type ScenarioCfg struct {
	Name         string            `yaml:"name" json:"name"`
	Priority     int               `yaml:"priority" json:"priority"`
	Match        *MatchRules       `yaml:"match" json:"match"`
	Status       int               `yaml:"status" json:"status"`
	Headers      map[string]string `yaml:"headers" json:"headers"`
	Body         any               `yaml:"body" json:"body"`
	BodyBase64   string            `yaml:"body_base64" json:"body_base64"`
	LatencyMS    int               `yaml:"latency_ms" json:"latency_ms"`
	ErrorRatePct int               `yaml:"error_rate_pct" json:"error_rate_pct"`
	Weight       int               `yaml:"weight" json:"weight"`
}

type MatchRules struct {
	Query   map[string]Cond `yaml:"query"   json:"query"`
	Headers map[string]Cond `yaml:"headers" json:"headers"`
	Body    map[string]Cond `yaml:"body"    json:"body"`
}

type Cond struct {
	Eq any   `yaml:"$eq" json:"$eq"`
	In []any `yaml:"$in" json:"$in"`
}
