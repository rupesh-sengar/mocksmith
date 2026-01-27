package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"mocksmith/internal/rules"
	"mocksmith/internal/server/store"
	"mocksmith/internal/server/utils"
	"mocksmith/internal/snapshot"
	"mocksmith/internal/templating"
)

type RuntimeService struct {
	state *store.State
}

type RuntimeRequest struct {
	Method  string
	Path    string
	Request *http.Request
	Body    []byte
}

type RuntimeResponse struct {
	Status      int
	Headers     map[string]string
	Body        []byte
	Scenario    string
	ShouldLog   bool
	RequestPath string
}

func NewRuntimeService(state *store.State) *RuntimeService {
	return &RuntimeService{state: state}
}

func (r *RuntimeService) Handle(req RuntimeRequest) RuntimeResponse {
	snap := r.state.Snapshot()
	requestPath := "/" + strings.TrimPrefix(req.Path, "/")
	method := strings.ToUpper(req.Method)

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
		return RuntimeResponse{
			Status:      404,
			Body:        jsonBody(map[string]any{"error": "route not found", "path": requestPath}),
			ShouldLog:   false,
			RequestPath: requestPath,
		}
	}

	var bodyMap map[string]any
	_ = json.Unmarshal(req.Body, &bodyMap)

	if rt.Strict && rt.ReqSchema != nil {
		if err := rt.ReqSchema.Validate(bytes.NewReader(req.Body)); err != nil {
			return RuntimeResponse{
				Status:      400,
				Body:        jsonBody(map[string]any{"error": "validation failed", "details": err.Error()}),
				ShouldLog:   false,
				RequestPath: requestPath,
			}
		}
	}

	ctx := snapshot.MatchContext{Query: req.Request.URL.Query(), Headers: req.Request.Header, Body: bodyMap}
	scen := rules.PickScenario(rt, ctx)

	if scen.LatencyMS > 0 {
		time.Sleep(time.Duration(scen.LatencyMS) * time.Millisecond)
	}
	if scen.ErrorRatePct > 0 && utils.Roll(scen.ErrorRatePct) {
		return RuntimeResponse{
			Status:      500,
			Body:        jsonBody(map[string]any{"error": "injected error"}),
			ShouldLog:   false,
			RequestPath: requestPath,
		}
	}

	headers := map[string]string{}
	for k, v := range scen.Headers {
		headers[k] = templating.RenderTokens(v, params, req.Request, snap.Env, bodyMap)
	}

	status := scen.Status
	if status == 0 {
		status = 200
	}
	rendered := templating.RenderTokens(string(scen.BodyBytes), params, req.Request, snap.Env, bodyMap)

	return RuntimeResponse{
		Status:      status,
		Headers:     headers,
		Body:        []byte(rendered),
		Scenario:    scen.Name,
		ShouldLog:   true,
		RequestPath: requestPath,
	}
}

func jsonBody(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{"error":"serialization failed"}`)
	}
	return b
}
