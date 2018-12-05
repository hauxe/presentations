package resourceprioritize

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// GateKeeper defines gate keeper for resource
type GateKeeper struct {
	ctx    context.Context
	cancel context.CancelFunc
	mux    sync.Mutex             // HL
	queue  chan PrioritizedEntity // HL
}

// Start starts monitoring for competition entity for resource
func (k *GateKeeper) Start(resource interface{}, waitTime time.Duration) {
	k.mux.Lock()
	defer k.mux.Unlock()
	if k.cancel != nil {
		k.cancel()
	}
	k.ctx, k.cancel = context.WithCancel(context.Background())
	mux := &GreedyMutex{
		TimeToWait: waitTime,
		queue:      k.queue,
	}
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		wg.Done()
		mux.Compete(k.ctx, resource) // HL
	}()
	wg.Wait()
	fmt.Println("successfully start monitoring for resource")
}

// Stop stops monitoring resource
func (k *GateKeeper) Stop() {
	if k.cancel != nil {
		k.cancel()
	}
}

// Register register for accessing resource
func (k *GateKeeper) Register(entity PrioritizedEntity) {
	k.queue <- entity // HL
}

// New new gate keeper
func New(queuedElements int) *GateKeeper {
	return &GateKeeper{
		queue: make(chan PrioritizedEntity, queuedElements),
	}
}
