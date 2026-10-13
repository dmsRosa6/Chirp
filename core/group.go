package core

import (
	"fmt"
	"math/rand"
	"sync"
)

const MAX_GROUP_SIZE = 32

type Group struct {
	mu      sync.RWMutex
	id      string
	pattern string
	subs    map[subKey]*Subscription
	r       []*Subscription
	index   int
}

func NewGroup(id, pattern string) *Group {
	return &Group{
		id:      id,
		pattern: pattern,
		subs:    make(map[subKey]*Subscription, MAX_GROUP_SIZE),
		r:       make([]*Subscription, MAX_GROUP_SIZE),
		index:   0,
	}
}

func (g *Group) ID() string {
	return g.id
}

func (g *Group) Pattern() string {
	return g.pattern
}

func (g *Group) AddSubscription(subKey subKey, sub *Subscription) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.index >= MAX_GROUP_SIZE {
		return fmt.Errorf("No more space on group %s. pattern: %s", g.id, g.pattern)
	}

	if sub == nil {
		return fmt.Errorf("subscription cannot be nil")
	}

	if _, ok := g.subs[subKey]; ok {
		return fmt.Errorf("Subscription %s already exists on group %s. pattern: %s", subKey.id, g.id, g.pattern)
	}

	g.subs[subKey] = sub
	g.r[g.index] = sub
	g.index++

	return nil
}

func (g *Group) RemoveSubscription(key subKey) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	sub, ok := g.subs[key]
	if !ok {
		return nil
	}

	delete(g.subs, key)

	for i := 0; i < g.index; i++ {
		if g.r[i] == sub {
			g.index--
			g.r[i] = g.r[g.index]
			g.r[g.index] = nil
			break
		}
	}

	return nil
}

func (g *Group) GetRandom() *Subscription {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.index == 0 {
		return nil
	}
	randomIndex := rand.Intn(g.index)

	sub := g.r[randomIndex]

	return sub
}
