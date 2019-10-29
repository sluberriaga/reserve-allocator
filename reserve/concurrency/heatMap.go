package concurrency

import (
	"errors"
	"sync"
)

type heatMap struct {
	// map[uin64]uint64
	hm sync.Map
	// map[uin64]syncRWMutex
	mu sync.Map
}

func newHeatMap() heatMap {
	return heatMap{
		hm: sync.Map{},
		mu: sync.Map{},
	}
}

var (
	LoadKeyMapError    = errors.New("could not load heat map key")
	ParseValueMapError = errors.New("could not load parse map value")
)

func (h *heatMap) LoadAndStore(key, initialValue uint64, fn func(entry uint64) uint64) error {

	entry, _ := h.mu.LoadOrStore(key, sync.RWMutex{})
	mu, ok := entry.(sync.RWMutex)
	if !ok {
		return ParseValueMapError
	}
	mu.Lock()

	entry, _ = h.hm.LoadOrStore(key, initialValue)
	heat, ok := entry.(uint64)
	if !ok {
		mu.Unlock()
		return ParseValueMapError
	}

	nextVal := fn(heat)
	if nextVal == 0 {
		h.hm.Delete(key)
		defer h.mu.Delete(key)
	} else {
		h.hm.Store(key, nextVal)
	}

	mu.Unlock()
	return nil
}

func (h *heatMap) Load(key uint64) (uint64, error) {

	entry, ok := h.mu.Load(key)
	if !ok {
		return 0, LoadKeyMapError
	}

	mu, ok := entry.(sync.RWMutex)
	if !ok {
		return 0, ParseValueMapError
	}
	mu.RLock()
	defer mu.RUnlock()

	entry, ok = h.hm.Load(key)
	if !ok {
		return 0, LoadKeyMapError
	}

	heat, ok := entry.(uint64)
	if !ok {
		return 0, ParseValueMapError
	}

	return heat, nil
}
