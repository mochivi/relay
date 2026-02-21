package router

import (
	"errors"
	"fmt"
	"strings"
)

type tree struct {
	root node
}

type node struct {
	key      string
	children []*node
	val      *string
	wildcard bool
}

func newTreeFromPatterns(patterns []string, vals []string) (*tree, error) {
	if len(patterns) != len(vals) {
		return nil, errors.New("patterns length must match vals length")
	}

	tree := &tree{root: node{}}
	for i := range patterns {
		tree.insert(patterns[i], vals[i])
	}
	return tree, nil
}

func (t *tree) insert(pattern string, val string) error {
	if !strings.Contains(pattern, "/") && strings.Trim(pattern, " ") != "" {
		return fmt.Errorf("invalid pattern: %q", pattern)
	}

	pattern = strings.TrimSuffix(pattern, "/") // remove trailing slashes
	keys := strings.Split(pattern, "/")
	current := &t.root

	for i := 0; i < len(keys)-1; i++ {
		nextKey := keys[i+1]
		child := current.findChildExact(nextKey)
		if child == nil {
			child = current.addChild(nextKey, nil)
		}
		current = child
	}

	current.val = &val
	return nil
}

func (t *tree) search(pattern string) (string, bool) {
	keys := strings.Split(pattern, "/")
	current := &t.root
	var match *node

	// Match root node if there's an / route registered
	if current.val != nil {
		match = current
	}

	for i := 0; i < len(keys)-1; i++ {
		nextKey := keys[i+1]

		child := current.findChild(nextKey)
		if child == nil {
			if current.wildcard { // keep going if in wildcard node
				continue
			}
			break
		}

		if child.val != nil {
			match = child
		}
		current = child
	}
	if match == nil {
		return "", false
	}
	return *match.val, true
}

func (n *node) addChild(key string, val *string) *node {
	if n.children == nil {
		n.children = make([]*node, 0)
	}
	node := &node{key: key, val: val}
	if key == "*" {
		node.wildcard = true
	}
	n.children = append(n.children, node)
	return node
}

// findChildExact returns a child only when key matches exactly (no wildcard fallback).
// Used when building the tree so /api/... and /*/... stay distinct.
func (n *node) findChildExact(key string) *node {
	for _, child := range n.children {
		if child.key == key {
			return child
		}
	}
	return nil
}

// findChild returns exact match if present, otherwise the wildcard node for path matching.
func (n *node) findChild(key string) *node {
	var wildcard *node
	for _, child := range n.children {
		if child.key == key {
			return child
		}
		if child.key == "*" {
			wildcard = child
		}
	}
	return wildcard
}

func (t *tree) print() {
	printNode(&t.root, "", true)
}

func printNode(n *node, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	label := n.key
	if n.val != nil {
		label += fmt.Sprintf(" [%s]", *n.val)
	}

	if n.key == "" {
		if n.val != nil {
			fmt.Printf("(root) [%s]\n", *n.val)
		} else {
			fmt.Println("(root)")
		}
	} else {
		fmt.Println(prefix + connector + label)
	}

	childPrefix := prefix
	if n.key != "" {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for i, child := range n.children {
		printNode(child, childPrefix, i == len(n.children)-1)
	}
}
