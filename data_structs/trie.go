package data_structs

import (
	"strings"
)

type Node[T any] struct {
	children map[string]*Node[T]
	value    *T
}

// Stupid simple and opinionated Trie cause we dont need something complex in our use case
type Trie[T any] struct {
	root                *Node[T]
	delimiter, wildcard string
}

func NewTrie[T any](delimiter, wildcard string) *Trie[T] {
	return &Trie[T]{
		root: &Node[T]{
			children: make(map[string]*Node[T]),
		},
		delimiter: delimiter,
		wildcard:  wildcard,
	}
}

func (t *Trie[T]) Add(path string, value T) {
	parts := strings.Split(path, t.delimiter)
	node := t.root

	for _, part := range parts {
		child, ok := node.children[part]
		if !ok {
			child = &Node[T]{children: make(map[string]*Node[T])}
			node.children[part] = child
		}
		node = child
	}

	node.value = &value
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

func (t *Trie[T]) getNode(path string) (*Node[T], bool) {
	return t.walk(strings.Split(path, t.delimiter))
}

func (t *Trie[T]) FindOne(path string) (*T, bool) {
	node, found := t.getNode(path)
	if !found || node.value == nil {
		return nil, false
	}
	return node.value, true
}

// For now we can assume the path has been sanitized before, so it's safe to use.
func (t *Trie[T]) FindNodeChildren(path string) ([]T, bool) {
	node, found := t.getNode(path)
	if !found {
		return nil, false
	}

	arr := make([]T, 0, len(node.children))
	for _, v := range node.children {
		if v.value != nil {
			arr = append(arr, *v.value)
		}
	}
	return arr, true
}

// FindByPrefix resolves a path whose last segment ends in "*" — e.g. "a.b*" —
// and returns the values of every direct child of "a" whose key starts with
// "b". Only direct children are matched, not deeper descendants. Wildcard
// syntax (trailing "*", only on the last segment) is assumed already
// validated upstream.
func (t *Trie[T]) FindByPrefix(path string) ([]T, bool) {
	parts := strings.Split(path, t.delimiter)
	last := parts[len(parts)-1]
	prefix := strings.TrimSuffix(last, t.wildcard)

	node, found := t.walk(parts[:len(parts)-1])
	if !found {
		return nil, false
	}

	arr := make([]T, 0, len(node.children))
	for key, child := range node.children {
		if child.value != nil && strings.HasPrefix(key, prefix) {
			arr = append(arr, *child.value)
		}
	}
	return arr, true
}
