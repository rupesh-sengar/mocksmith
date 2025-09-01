package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Registry struct {
	mu  sync.Mutex
	all map[string]*rate.Limiter
}

func New() *Registry { return &Registry{all: make(map[string]*rate.Limiter)} }

func (r *Registry) Get(key string, rpm int) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if lim, ok := r.all[key]; ok {
		return lim
	}
	lim := rate.NewLimiter(rate.Every(time.Minute/time.Duration(rpm)), rpm)
	r.all[key] = lim
	return lim
}
