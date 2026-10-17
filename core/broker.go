package core

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"log"

	"github.com/dmsRosa6/Chirp/data_structs"
	"github.com/dmsRosa6/Chirp/grammar"
	"github.com/dmsRosa6/Chirp/model"
)

// DefaultOutBuffer is how many frames may wait to be written to one session
// before it counts as a slow consumer.
const DefaultOutBuffer = 256

// Config tunes the broker. Zero values mean defaults.
type Config struct {
	OutBuffer int
}

// Broker owns every subscription and routes publishes to them.
//
// Patterns live in a trie, so matching happens at publish time: a
// subscription applies to subjects that have never been seen before.
type Broker struct {
	cfg    Config
	nextID atomic.Uint64

	mu    sync.RWMutex // guards trie and count
	trie  *data_structs.Trie[*patternEntry]
	count int
}

func NewBroker(cfg Config) *Broker {
	if cfg.OutBuffer <= 0 {
		cfg.OutBuffer = DefaultOutBuffer
	}
	return &Broker{
		cfg:  cfg,
		trie: data_structs.NewTrie[*patternEntry](grammar.SubjectDelimiter, grammar.SubjectWildcard),
	}
}

func (b *Broker) Get(subject string) ([]model.Subject, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	entries := b.trie.Match(subject)
	if len(entries) == 0 {
		return nil, fmt.Errorf("there is no subject matching '%s'", subject)
	}

	out := make([]model.Subject, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.snapshot())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Pattern < out[j].Pattern })
	return out, nil
}

// Subscribe registers pattern for the session under subID, this subscritions can be grouped by groupID. The caller is
// expected to have validated the pattern and the id.
func (b *Broker) Subscribe(s *Session, pattern, subID, groupID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := &Subscription{ID: subID, Pattern: pattern, Session: s}
	if err := s.addSub(sub); err != nil {
		return err
	}

	var entry *patternEntry
	if e, ok := b.trie.FindOne(pattern); ok {
		entry = *e
	} else {
		entry = &patternEntry{
			pattern: pattern,
			subs:    make(map[subKey]*Subscription),
			groups:  make(map[string]*Group),
		}
		b.trie.Add(pattern, entry)
	}
	if groupID != "" {
		if v, ok := entry.groups[groupID]; ok {
			v.AddSubscription(sub.key(), sub)
		} else {
			g := NewGroup(groupID, pattern)
			g.AddSubscription(sub.key(), sub)
			entry.groups[groupID] = g
		}
	} else {
		entry.subs[sub.key()] = sub
	}

	b.count++
	return nil
}

// Unsubscribe removes the session's subscription with this id.
func (b *Broker) Unsubscribe(s *Session, subID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub, ok := s.takeSub(subID)
	if !ok {
		return ErrNoSuchSubscription
	}
	b.detachLocked(sub)
	return nil
}

// Publish delivers the payload to every subscription whose pattern matches
// the subject and returns how many that was. Delivery never blocks.
func (b *Broker) Publish(subject, payload string) int {
	msg := Message{Subject: subject, Body: payload, Time: time.Now().UnixNano()}

	// Snapshot under the lock, deliver outside it: Push may close a slow
	// session, which takes the lock again.
	b.mu.RLock()
	var targets []*Subscription
	for _, entry := range b.trie.Match(subject) {
		for _, sub := range entry.subs {
			targets = append(targets, sub)
		}
		for _, g := range entry.groups {
			targets = append(targets, g.GetRandom())
		}
	}
	b.mu.RUnlock()

	for _, sub := range targets {
		log.Default().Printf("Send message to %s...", sub.Session.addr)
		sub.deliver(msg)
	}
	return len(targets)
}

// SubscriptionCount is the number of live subscriptions across all sessions.
func (b *Broker) SubscriptionCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.count
}

// detachAll removes subscriptions of a closing session.
func (b *Broker) detachAll(subs []*Subscription) {
	if len(subs) == 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, sub := range subs {
		b.detachLocked(sub)
	}
}

// detachLocked must be called with b.mu held for writing.
func (b *Broker) detachLocked(sub *Subscription) {
	e, ok := b.trie.FindOne(sub.Pattern)
	if !ok {
		return
	}
	entry := *e
	if _, present := entry.subs[sub.key()]; present {
		delete(entry.subs, sub.key())
		b.count--
	}
	if len(entry.subs) == 0 {
		b.trie.Remove(sub.Pattern)
	}
}
