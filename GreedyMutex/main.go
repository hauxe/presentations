package main

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
func (mux *GreedyMutex) Compete(ctx context.Context, resource interface{}) {
	mux.Lock()
	greedyEntities := []PrioritizedEntity{}
	defer func() { go mux.Compete(ctx, resource) }()
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
	entity := <-mux.queue
	greedyEntities = append(greedyEntities, entity)
	c := time.NewTicker(mux.TimeToWait)
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

// PrioritizedEntitySample defines prioritized entity
type PrioritizedEntitySample struct {
	ID       int
	Priority int
}

// GetPriority gets identity priority
func (p *PrioritizedEntitySample) GetPriority() int {
	return p.Priority
}

// AccessResource access the resource
func (p *PrioritizedEntitySample) AccessResource(resource interface{}) {
	res, ok := resource.(map[int]int)
	if !ok {
		fmt.Println("Error resource not expected")
	}
	res[p.ID]++
}

// END Sample

func main() {
	// a shared resource keep track of number of accessed entity id
	sharedResource := make(map[int]int, 3) // HL
	gateKeeper := New(100)
	gateKeeper.Start(sharedResource, time.Microsecond)
	defer gateKeeper.Stop()
	// 3 different prioritized entity
	highEntity := &PrioritizedEntitySample{
		ID:       1,
		Priority: 3,
	}
	mediumEntity := &PrioritizedEntitySample{
		ID:       2,
		Priority: 2,
	}
	lowEntity := &PrioritizedEntitySample{
		ID:       3,
		Priority: 1,
	}

	// start greeding resource count
	for i := 0; i < 100; i++ {
		go gateKeeper.Register(lowEntity)
		go gateKeeper.Register(mediumEntity)
		go gateKeeper.Register(highEntity)
	}
	time.Sleep(time.Second)
	fmt.Printf("Resource Status %#v\n", sharedResource)
}
