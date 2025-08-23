package games

import "fmt"

var providers = map[string]Provider{}

// Register adds a provider; panics on duplicate id.
func Register(p Provider) {
	id := p.ID()
	if _, exists := providers[id]; exists {
		panic("duplicate games provider: " + id)
	}
	providers[id] = p
}

// Get returns a provider by id.
func Get(id string) (Provider, error) {
	if p, ok := providers[id]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("unknown game provider: %s", id)
}

// All returns all registered providers.
func All() []Provider {
	out := make([]Provider, 0, len(providers))
	for _, p := range providers {
		out = append(out, p)
	}
	return out
}
