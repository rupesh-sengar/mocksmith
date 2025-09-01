package templating

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func RenderTokens(s string, params map[string]string, r *http.Request, env map[string]string, body map[string]any) string {
	repl := map[string]string{
		"{{now.iso}}": time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range params {
		repl["{{params."+k+"}}"] = v
	}
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			repl["{{query."+k+"}}"] = v[0]
		}
	}
	for k, v := range r.Header {
		if len(v) > 0 {
			repl["{{header."+strings.ToLower(k)+"}}"] = v[0]
		}
	}
	for k, v := range env {
		repl["{{env."+k+"}}"] = v
	}
	if body != nil {
		jb, _ := json.Marshal(body)
		repl["{{body}}"] = string(jb)
	}
	out := s
	for k, v := range repl {
		out = strings.ReplaceAll(out, k, v)
	}
	return out
}

func RenderTokensBytes(b []byte, params map[string]string, r *http.Request, env map[string]string, body map[string]any) []byte {
	return []byte(RenderTokens(string(b), params, r, env, body))
}
