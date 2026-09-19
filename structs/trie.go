package structs

import "strings"

type Node[T any] struct {
	children map[string]*Node[T]
	value    *T
}

// Stupid simple Trie cause we dont need something complex in our use case
type Trie[T any] struct {
	root *Node[T]
}

func NewTrie[T any]() *Trie[T] {
	return &Trie[T]{
		root: &Node[T]{
			children: make(map[string]*Node[T]),
		},
	}
}

func (t *Trie[T]) Add(path string, value T) {
	parts := strings.Split(path, ".")

	node := t.root

	for _, part := range parts {
		child, ok := node.children[part]
		if !ok {
			child = &Node[T]{
				children: make(map[string]*Node[T]),
			}
			node.children[part] = child
		}

		node = child
	}

	node.value = &value
}

func (t *Trie[T]) Find(path string) (*T, bool) {
	parts := strings.Split(path, ".")

	node := t.root

	for _, part := range parts {
		child, ok := node.children[part]
		if !ok {
			return nil, false
		}

		node = child
	}

	if node.value == nil {
		return nil, false
	}

	return node.value, true
}
