package agents

import "sort"

// Registry is an immutable-after-construction lookup of adapters — never
// global mutable state (research.md).
type Registry struct {
	adapters map[string]Adapter
}

// NewRegistry builds a Registry containing exactly the given adapters,
// keyed by their own ID().
func NewRegistry(adapters ...Adapter) *Registry {
	r := &Registry{adapters: make(map[string]Adapter, len(adapters))}
	for _, a := range adapters {
		r.adapters[a.ID()] = a
	}
	return r
}

// List returns every registered adapter, sorted by ID for deterministic
// output. No filesystem access (FR-002).
func (r *Registry) List() []Adapter {
	ids := make([]string, 0, len(r.adapters))
	for id := range r.adapters {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	list := make([]Adapter, 0, len(ids))
	for _, id := range ids {
		list = append(list, r.adapters[id])
	}
	return list
}

// Get returns (adapter, true) for a registered id, or (nil, false) —
// never an error, never an arbitrary default — for one that isn't
// (FR-003).
func (r *Registry) Get(id string) (Adapter, bool) {
	a, ok := r.adapters[id]
	return a, ok
}
