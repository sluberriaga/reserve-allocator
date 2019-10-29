package allocator

import (
	"errors"
	"github.com/emirpasic/gods/maps/treebidimap"
	"github.com/emirpasic/gods/utils"
	"reserve/reserve"
	"sync"
)

type registry struct {
	// map[uin64]syncRWMutex
	mu sync.Map
	// map[uint64]map[time.Time]reserve.Reserve
	rm map[uint64]*treebidimap.Map
}

func newRegistry() registry {
	return registry{
		mu: sync.Map{},
		rm: map[uint64]*treebidimap.Map{},
	}
}

var (
	LoadKeyMapError    = errors.New("could not load reserves key")
	ParseValueMapError = errors.New("could not load parse reserves value")
)

func (r *registry) LoadAndStore(key uint64, fn func(reserves treebidimap.Map) treebidimap.Map) error {
	entry, _ := r.mu.LoadOrStore(key, sync.RWMutex{})
	mu, ok := entry.(sync.RWMutex)
	if !ok {
		return ParseValueMapError
	}
	mu.Lock()

	reserves, ok := r.rm[key]
	if !ok {
		reserves = treebidimap.NewWith(utils.TimeComparator, reserve.ByAmountComparator)
		r.rm[key] = reserves
	}

	nextVal := fn(*reserves)
	if nextVal.Size() == 0 {
		r.rm[key] = nil
		delete(r.rm, key)
		defer r.mu.Delete(key)
	} else {
		*r.rm[key] = nextVal
	}

	mu.Unlock()
	return nil
}

func (r *registry) Load(key uint64) (treebidimap.Map, bool, error) {

	entry, ok := r.mu.Load(key)
	if !ok {
		return treebidimap.Map{}, false, nil
	}

	mu, ok := entry.(sync.RWMutex)
	if !ok {
		return treebidimap.Map{}, true, ParseValueMapError
	}
	mu.RLock()
	defer mu.RUnlock()

	reserves, ok := r.rm[key]
	if !ok {
		return treebidimap.Map{}, false, nil
	}

	return *reserves, true, nil
}
