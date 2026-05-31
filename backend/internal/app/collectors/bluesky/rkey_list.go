package bluesky

import "sync"

type RKeyList struct {
	mu sync.RWMutex
	m  map[string]bool
}

func NewRKeyList() *RKeyList {
	return &RKeyList{
		m: make(map[string]bool),
	}
}

func (r RKeyList) Add(rkey string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.m[rkey] = true
}

func (r RKeyList) Contains(rkey string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.m[rkey]

	return exists
}

func (r RKeyList) Remove(rkey string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.m, rkey)
}

func (r *RKeyList) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.m)
}
