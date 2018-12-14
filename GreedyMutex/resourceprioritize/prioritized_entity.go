package resourceprioritize

import (
	"context"
	"sync"
	"time"
)

// PrioritizedEntity interface
type PrioritizedEntity interface {
	GetPriority() int           // HL
	AccessResource(interface{}) // HL
}

// GreedyMutex a mutex that handles greedy entities with higher priority
type GreedyMutex struct {
	sync.Mutex
	TimeToWait time.Duration          // HL
	queue      chan PrioritizedEntity // HL
}

// Compete compete for resource
func (mux *GreedyMutex) Compete(ctx context.Context, resource interface{}) { // HL
	mux.Lock()
	greedyEntities := []PrioritizedEntity{}
	defer func() { go mux.Compete(ctx, resource) }() // HL
	defer mux.Unlock()
	defer func() { // HL
		var winner PrioritizedEntity
		for _, entity := range greedyEntities { // HL
			if winner == nil || winner.GetPriority() < entity.GetPriority() { // HL
				winner = entity // HL
			}
		}
		if winner != nil {
			winner.AccessResource(resource) // HL
		}
	}()
	// wait until have a request for resource
	entity := <-mux.queue // HL
	greedyEntities = append(greedyEntities, entity) // HL
	c := time.NewTicker(mux.TimeToWait) // HL
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.C: // HL
			c.Stop()
			return
		case entity := <-mux.queue: // HL
			greedyEntities = append(greedyEntities, entity)
		}
	}
}

// END
