package executor

import (
	"context"
	"fmt"
	"sync"
	"errors"
)

type Registry struct {
	mu      sync.Mutex
	factory Factory
	pool    map[string]Executor
}

func NewRegistry(factory Factory) *Registry {
	return &Registry{factory: factory, pool: make(map[string]Executor)}
}

func (r *Registry) Get(
	ctx context.Context,
	key string,
	cfg *Config,
) (Executor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ex, ok := r.pool[key]; ok {
		return ex, nil
	}

	ex, err := r.factory(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating executor for %q: %w", key, err)
	}
	r.pool[key] = ex

	return ex, nil
}

func (r *Registry) Evict(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ex, ok := r.pool[key]
	if !ok {
		return nil
	}
	delete(r.pool, key)

	if err := ex.Close(); err != nil {
		return fmt.Errorf("closing executor %q: %w", err)
	}

	return nil
}

func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for key, ex := range r.pool {
		if err := ex.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing executor %q: %w", key, err))
		}
	}
	clear(r.pool)

	return errors.Join(errs...)
}
