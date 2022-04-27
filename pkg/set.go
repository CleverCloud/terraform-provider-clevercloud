package pkg

import "sync"

// Immutable Set of string
type Set struct {
	items map[string]struct{}
	lock  sync.RWMutex
}

func NewSet(items ...string) *Set {
	m := map[string]struct{}{}
	for _, item := range items {
		m[item] = struct{}{}
	}

	return &Set{items: m}
}

func (s *Set) Contains(item string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, ok := s.items[item]
	return ok
}
