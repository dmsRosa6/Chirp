package data_structs

import (
	"strings"
)

type Node[T any] struct {
	children map[string]*Node[T]
	value    *T
}

// Stupid simple and opinionated Trie cause we dont need something complex in our use case.
//
// It stores values under dotted paths. In the broker the paths are
// subscription patterns ("foo.bar", "foo.b*") and Match answers the
// publish-time question: which stored patterns match this concrete subject?
//
// The trie is not safe for concurrent use; the caller (the broker) locks.
type Trie[T any] struct {
	root                *Node[T]
	delimiter, wildcard string
}

func NewTrie[T any](delimiter, wildcard string) *Trie[T] {
	return &Trie[T]{
		root:      newNode[T](),
		delimiter: delimiter,
		wildcard:  wildcard,
	}
}

func newNode[T any]() *Node[T] {
	return &Node[T]{children: make(map[string]*Node[T])}
}

// Add stores value under path, replacing any existing value.
func (t *Trie[T]) Add(path string, value T) {
	node := t.root
	for _, part := range strings.Split(path, t.delimiter) {
		child, ok := node.children[part]
		if !ok {
			child = newNode[T]()
			node.children[part] = child
		}
		node = child
	}
	node.value = &value
}

// FindOne returns the value stored under exactly this path (no wildcard
// interpretation).
func (t *Trie[T]) FindOne(path string) (*T, bool) {
	node, found := t.walk(strings.Split(path, t.delimiter))
	if !found || node.value == nil {
		return nil, false
	}
	return node.value, true
}

// Remove deletes the value stored under path and prunes any nodes that are
// left with neither a value nor children. It reports whether a value existed.
func (t *Trie[T]) Remove(path string) bool {
	parts := strings.Split(path, t.delimiter)

	// nodes[i] is the node reached after consuming i parts.
	nodes := make([]*Node[T], 0, len(parts)+1)
	node := t.root
	nodes = append(nodes, node)
	for _, part := range parts {
		child, ok := node.children[part]
		if !ok {
			return false
		}
		node = child
		nodes = append(nodes, node)
	}
	if node.value == nil {
		return false
	}
	node.value = nil

	for i := len(parts) - 1; i >= 0; i-- {
		n := nodes[i+1]
		if n.value != nil || len(n.children) > 0 {
			break
		}
		delete(nodes[i].children, parts[i])
	}
	return true
}

// Match returns the values of every stored pattern that matches the concrete
// subject.
//
// All tokens but the last must match exactly. At the last level a stored key
// matches if it equals the subject's last token, or if it ends in the
// wildcard and the last token starts with the key minus the wildcard
// ("b*" matches "b", "bar", "baz"). Matching is one level only: "a.b*" does
// not match "a.bar.baz".
func (t *Trie[T]) Match(subject string) []T {
	parts := strings.Split(subject, t.delimiter)
	last := parts[len(parts)-1]

	node, found := t.walk(parts[:len(parts)-1])
	if !found {
		return nil
	}

	var out []T
	for key, child := range node.children {
		if child.value == nil {
			continue
		}
		if key == last || (strings.HasSuffix(key, t.wildcard) &&
			strings.HasPrefix(last, strings.TrimSuffix(key, t.wildcard))) {
			out = append(out, *child.value)
		}
	}
	return out
}

// walk resolves an exact (non-wildcard) sequence of path segments to a node.
func (t *Trie[T]) walk(parts []string) (*Node[T], bool) {
	node := t.root
	for _, part := range parts {
		child, ok := node.children[part]
		if !ok {
			return nil, false
		}
		node = child
	}
	return node, true
}
