package compile

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"mocksmith/internal/config"
	"mocksmith/internal/snapshot"
)

func Build(cfg *config.Config) (*snapshot.Snapshot, error) {
	if cfg.Project == "" {
		cfg.Project = "demo"
	}
	rpm := 0
	if cfg.RateLimit != nil && cfg.RateLimit.RequestsPerMinute > 0 {
		rpm = cfg.RateLimit.RequestsPerMinute
	}

	snap := &snapshot.Snapshot{
		Project:   cfg.Project,
		Env:       cfg.Env,
		APIKeys:   map[string]struct{}{cfg.APIKey: {}},
		RateLimit: rpm,
	}

	for _, rc := range cfg.Routes {
		re, names := compilePath(rc.PathTemplate)

		var reqSch *jsonschema.Schema
		if rc.Schemas != nil && strings.TrimSpace(rc.Schemas.RequestJSONSchema) != "" {
			comp := jsonschema.NewCompiler()
			_ = comp.AddResource("schema.json", strings.NewReader(rc.Schemas.RequestJSONSchema))
			s, err := comp.Compile("schema.json")
			if err != nil {
				return nil, fmt.Errorf("compile jsonschema for %s %s: %w", rc.Method, rc.PathTemplate, err)
			}
			reqSch = s
		}

		cr := snapshot.CompiledRoute{
			Method:       strings.ToUpper(rc.Method),
			PathTemplate: rc.PathTemplate,
			PathRE:       re,
			ParamNames:   names,
			Strict:       rc.Strict,
			ReqSchema:    reqSch,
		}

		for _, sc := range rc.Scenarios {
			if sc.Weight == 0 {
				sc.Weight = 1
			}
			if sc.Priority == 0 {
				sc.Priority = 100
			}
			b, _ := json.Marshal(sc.Body)
			cr.Scenarios = append(cr.Scenarios, snapshot.CompiledScenario{
				Name:         sc.Name,
				Priority:     sc.Priority,
				Match:        sc.Match,
				Status:       sc.Status,
				Headers:      sc.Headers,
				BodyBytes:    b,
				BodyIsBase64: sc.BodyBase64 != "",
				LatencyMS:    sc.LatencyMS,
				ErrorRatePct: sc.ErrorRatePct,
				Weight:       sc.Weight,
			})
		}

		if rc.Default != nil {
			b, _ := json.Marshal(rc.Default.Body)
			def := &snapshot.CompiledScenario{
				Name:      ifEmpty(rc.Default.Name, "default"),
				Priority:  999999,
				Status:    ifZero(rc.Default.Status, 200),
				Headers:   rc.Default.Headers,
				BodyBytes: b,
				LatencyMS: rc.Default.LatencyMS,
			}
			cr.Default = def
		}

		snap.Routes = append(snap.Routes, cr)
	}

	return snap, nil
}

func compilePath(tpl string) (*regexp.Regexp, []string) {
	trim := strings.TrimPrefix(tpl, "/")
	parts := strings.Split(trim, "/")
	var names []string
	var b strings.Builder
	b.WriteString("^/")
	for i, p := range parts {
		if strings.HasPrefix(p, ":") {
			names = append(names, p[1:])
			b.WriteString("([^/]+)")
		} else {
			b.WriteString(regexp.QuoteMeta(p))
		}
		if i < len(parts)-1 {
			b.WriteString("/")
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String()), names
}

func ifZero(v, alt int) int { if v == 0 { return alt }; return v }
func ifEmpty(v, alt string) string { if strings.TrimSpace(v) == "" { return alt }; return v }
