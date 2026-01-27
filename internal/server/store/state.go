package store

import (
	"sync/atomic"

	"mocksmith/internal/ratelimit"
	"mocksmith/internal/snapshot"
)

type State struct {
	snap   atomic.Value // *snapshot.Snapshot
	limits *ratelimit.Registry
}

func NewState(initial *snapshot.Snapshot) *State {
	st := &State{limits: ratelimit.New()}
	st.snap.Store(initial)
	return st
}

func (s *State) Snapshot() *snapshot.Snapshot {
	return s.snap.Load().(*snapshot.Snapshot)
}

func (s *State) StoreSnapshot(snap *snapshot.Snapshot) {
	s.snap.Store(snap)
}

func (s *State) Limits() *ratelimit.Registry {
	return s.limits
}
