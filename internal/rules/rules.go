package rules

import (
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"fmt"
	"time"
	"mocksmith/internal/config"
	"mocksmith/internal/snapshot"
)

var idxRe = regexp.MustCompile(`^(.+)\[(\d+)\]$`)

func PickScenario(rt *snapshot.CompiledRoute, ctx snapshot.MatchContext) snapshot.CompiledScenario {
	// sort by priority asc
	scs := append([]snapshot.CompiledScenario(nil), rt.Scenarios...)
	sort.SliceStable(scs, func(i, j int) bool { return scs[i].Priority < scs[j].Priority })

	for i := 0; i < len(scs); {
		p := scs[i].Priority
		j := i + 1
		for j < len(scs) && scs[j].Priority == p {
			j++
		}
		group := scs[i:j]

		candidates := []snapshot.CompiledScenario{}
		for _, s := range group {
			if ok := rulesOK(s.Match, ctx); ok {
				candidates = append(candidates, s)
			}
		}
		if len(candidates) > 0 {
			return weightedPick(candidates)
		}
		i = j
	}

	if rt.Default != nil {
		return *rt.Default
	}
	if len(scs) > 0 {
		return scs[0]
	}
	return snapshot.CompiledScenario{Name: "fallback", Status: 200, BodyBytes: []byte(`{}`)}
}

func rulesOK(m *config.MatchRules, ctx snapshot.MatchContext) bool {
	if m == nil {
		return true
	}
	return condsQuery(m.Query, ctx.Query) &&
		condsHeader(m.Headers, ctx.Headers) &&
		condsBody(m.Body, ctx.Body)
}

func condsQuery(m map[string]config.Cond, q url.Values) bool {
	for k, c := range m {
		vals := q[k]
		if !checkCond(vals, c) {
			return false
		}
	}
	return true
}
func condsHeader(m map[string]config.Cond, h http.Header) bool {
	for k, c := range m {
		vals := h[http.CanonicalHeaderKey(k)]
		if !checkCond(vals, c) {
			return false
		}
	}
	return true
}
func condsBody(m map[string]config.Cond, b map[string]any) bool {
	for path, c := range m {
		v := getDot(b, path)
		if !checkCond([]string{toString(v)}, c) {
			return false
		}
	}
	return true
}

func checkCond(vals []string, c config.Cond) bool {
	if c.Eq != nil {
		if len(vals) == 0 || vals[0] != toString(c.Eq) {
			return false
		}
	}
	if len(c.In) > 0 {
		allowed := map[string]struct{}{}
		for _, v := range c.In {
			allowed[toString(v)] = struct{}{}
		}
		ok := false
		for _, v := range vals {
			if _, hit := allowed[v]; hit {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func getDot(m map[string]any, path string) any {
	cur := any(m)
	for _, seg := range strings.Split(path, ".") {
		if mm, ok := cur.(map[string]any); ok {
			if m := idxRe.FindStringSubmatch(seg); m != nil {
				arr, ok := mm[m[1]].([]any)
				if !ok {
					return nil
				}
				i, _ := strconv.Atoi(m[2])
				if i < 0 || i >= len(arr) {
					return nil
				}
				cur = arr[i]
			} else {
				cur = mm[seg]
			}
		} else {
			return nil
		}
	}
	return cur
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := jsonMarshalSafe(v)
		s := string(b)
		return strings.Trim(s, "\"")
	}
}

// tiny fallback marshal without importing encoding/json here
func jsonMarshalSafe(v any) ([]byte, error) {
	type jm interface{ MarshalJSON() ([]byte, error) }
	if m, ok := v.(jm); ok {
		return m.MarshalJSON()
	}
	// defer real json to caller; keep it simple
	return []byte(`"` + strings.ReplaceAll(strings.TrimSpace(fmtAny(v)), `"`, `\"`) + `"`), nil
}

func fmtAny(v any) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.Trim(fmt.Sprintf("%v", v), "\n"), "\n", " "), "\t", " "), "  ", " "))
}

func weightedPick(list []snapshot.CompiledScenario) snapshot.CompiledScenario {
	sum := 0
	for _, s := range list {
		if s.Weight <= 0 {
			s.Weight = 1
		}
		sum += s.Weight
	}
	n := randInt(sum)
	acc := 0
	for _, s := range list {
		acc += s.Weight
		if n < acc {
			return s
		}
	}
	return list[0]
}

// small rand without importing math/rand here
var seed = time.Now().UnixNano()
func randInt(n int) int { seed = (1103515245*seed + 12345) & 0x7fffffff; return int(seed % int64(n)) }
