// Copyright (c) 2015, Emir Pasic. All rights reserved.
// Adapted for code-gen by reddec, 2019
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Package redblacktree implements a red-black tree.
//
// Used by TreeSet and TreeMap.
//
// Structure is not thread safe.
//
// References: http://en.wikipedia.org/wiki/Red%E2%80%93black_tree
package model

import (
	"fmt"
)

// Tree holds elements of the red-black tree
type Tree struct {
	Root *TreeNode
	size int
}

// TreeNode is a single element within the tree
type TreeNode struct {
	Key    int64
	Value  *User
	color  bool // black - true, red - false
	Left   *TreeNode
	Right  *TreeNode
	Parent *TreeNode
}

// Instantiates a red-black tree
func NewTree() *Tree {
	return &Tree{}
}

// Put inserts node into the tree.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Put(key int64, value *User) {
	var insertedNode *TreeNode
	if tree.Root == nil {
		tree.Root = &TreeNode{Key: key, Value: value, color: false}
		insertedNode = tree.Root
	} else {
		node := tree.Root
		loop := true
		for loop {

			compare := key - node.Key

			switch {
			case compare == 0:
				node.Key = key
				node.Value = value
				return
			case compare < 0:
				if node.Left == nil {
					node.Left = &TreeNode{Key: key, Value: value, color: false}
					insertedNode = node.Left
					loop = false
				} else {
					node = node.Left
				}
			case compare > 0:
				if node.Right == nil {
					node.Right = &TreeNode{Key: key, Value: value, color: false}
					insertedNode = node.Right
					loop = false
				} else {
					node = node.Right
				}
			}
		}
		insertedNode.Parent = node
	}
	tree.insertCase1(insertedNode)
	tree.size++
}

// Get searches the node in the tree by key and returns its value or nil if key is not found in tree.
// Second return parameter is true if key was found, otherwise false.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Get(key int64) (value *User, found bool) {
	node := tree.Lookup(key)
	if node != nil {
		return node.Value, true
	}
	return nil, false
}

// Remove remove the node from the tree by key.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Remove(key int64) {
	var child *TreeNode
	node := tree.Lookup(key)
	if node == nil {
		return
	}
	if node.Left != nil && node.Right != nil {
		pred := node.Left.maximumNode()
		node.Key = pred.Key
		node.Value = pred.Value
		node = pred
	}
	if node.Left == nil || node.Right == nil {
		if node.Right == nil {
			child = node.Left
		} else {
			child = node.Right
		}
		if node.color == true {
			node.color = nodeColor(child)
			tree.deleteCase1(node)
		}
		tree.replaceNode(node, child)
		if node.Parent == nil && child != nil {
			child.color = true
		}
	}
	tree.size--
}

// Empty returns true if tree does not contain any nodes
func (tree *Tree) Empty() bool {
	return tree.size == 0
}

// Size returns number of nodes in the tree.
func (tree *Tree) Size() int {
	return tree.size
}

// Keys returns all keys in-order
func (tree *Tree) Keys() []int64 {
	keys := make([]int64, tree.size)
	it := tree.Iterator()
	for i := 0; it.Next(); i++ {
		keys[i] = it.Key()
	}
	return keys
}

// Values returns all values in-order based on the key.
func (tree *Tree) Values() []*User {
	values := make([]*User, tree.size)
	it := tree.Iterator()
	for i := 0; it.Next(); i++ {
		values[i] = it.Value()
	}
	return values
}

// Left returns the left-most (min) node or nil if tree is empty.
func (tree *Tree) Left() *TreeNode {
	var parent *TreeNode
	current := tree.Root
	for current != nil {
		parent = current
		current = current.Left
	}
	return parent
}

// Right returns the right-most (max) node or nil if tree is empty.
func (tree *Tree) Right() *TreeNode {
	var parent *TreeNode
	current := tree.Root
	for current != nil {
		parent = current
		current = current.Right
	}
	return parent
}

// Floor Finds floor node of the input key, return the floor node or nil if no floor is found.
// Second return parameter is true if floor was found, otherwise false.
//
// Floor node is defined as the largest node that is smaller than or equal to the given node.
// A floor node may not be found, either because the tree is empty, or because
// all nodes in the tree are larger than the given node.
//
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Floor(key int64) (floor *TreeNode, found bool) {
	found = false
	node := tree.Root
	for node != nil {

		compare := key - node.Key

		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			node = node.Left
		case compare > 0:
			floor, found = node, true
			node = node.Right
		}
	}
	if found {
		return floor, true
	}
	return nil, false
}

// Ceiling finds ceiling node of the input key, return the ceiling node or nil if no ceiling is found.
// Second return parameter is true if ceiling was found, otherwise false.
//
// Ceiling node is defined as the smallest node that is larger than or equal to the given node.
// A ceiling node may not be found, either because the tree is empty, or because
// all nodes in the tree are smaller than the given node.
//
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Ceiling(key int64) (ceiling *TreeNode, found bool) {
	found = false
	node := tree.Root
	for node != nil {

		compare := key - node.Key

		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			ceiling, found = node, true
			node = node.Left
		case compare > 0:
			node = node.Right
		}
	}
	if found {
		return ceiling, true
	}
	return nil, false
}

// Clear removes all nodes from the tree.
func (tree *Tree) Clear() {
	tree.Root = nil
	tree.size = 0
}

// String returns a string representation of container
func (tree *Tree) String() string {
	str := "RedBlackTree\n"
	if !tree.Empty() {
		output(tree.Root, "", true, &str)
	}
	return str
}

func (node *TreeNode) String() string {
	return fmt.Sprintf("%v", node.Key)
}

func output(node *TreeNode, prefix string, isTail bool, str *string) {
	if node.Right != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		output(node.Right, newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += node.String() + "\n"
	if node.Left != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		output(node.Left, newPrefix, true, str)
	}
}

func (tree *Tree) Lookup(key int64) *TreeNode {
	node := tree.Root
	for node != nil {

		compare := key - node.Key

		switch {
		case compare == 0:
			return node
		case compare < 0:
			node = node.Left
		case compare > 0:
			node = node.Right
		}
	}
	return nil
}

func (node *TreeNode) grandparent() *TreeNode {
	if node != nil && node.Parent != nil {
		return node.Parent.Parent
	}
	return nil
}

func (node *TreeNode) uncle() *TreeNode {
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return nil
	}
	return node.Parent.sibling()
}

func (node *TreeNode) sibling() *TreeNode {
	if node == nil || node.Parent == nil {
		return nil
	}
	if node == node.Parent.Left {
		return node.Parent.Right
	}
	return node.Parent.Left
}

func (tree *Tree) rotateLeft(node *TreeNode) {
	right := node.Right
	tree.replaceNode(node, right)
	node.Right = right.Left
	if right.Left != nil {
		right.Left.Parent = node
	}
	right.Left = node
	node.Parent = right
}

func (tree *Tree) rotateRight(node *TreeNode) {
	left := node.Left
	tree.replaceNode(node, left)
	node.Left = left.Right
	if left.Right != nil {
		left.Right.Parent = node
	}
	left.Right = node
	node.Parent = left
}

func (tree *Tree) replaceNode(old *TreeNode, new *TreeNode) {
	if old.Parent == nil {
		tree.Root = new
	} else {
		if old == old.Parent.Left {
			old.Parent.Left = new
		} else {
			old.Parent.Right = new
		}
	}
	if new != nil {
		new.Parent = old.Parent
	}
}

func (tree *Tree) insertCase1(node *TreeNode) {
	if node.Parent == nil {
		node.color = true
	} else {
		tree.insertCase2(node)
	}
}

func (tree *Tree) insertCase2(node *TreeNode) {
	if nodeColor(node.Parent) == true {
		return
	}
	tree.insertCase3(node)
}

func (tree *Tree) insertCase3(node *TreeNode) {
	uncle := node.uncle()
	if nodeColor(uncle) == false {
		node.Parent.color = true
		uncle.color = true
		node.grandparent().color = false
		tree.insertCase1(node.grandparent())
	} else {
		tree.insertCase4(node)
	}
}

func (tree *Tree) insertCase4(node *TreeNode) {
	grandparent := node.grandparent()
	if node == node.Parent.Right && node.Parent == grandparent.Left {
		tree.rotateLeft(node.Parent)
		node = node.Left
	} else if node == node.Parent.Left && node.Parent == grandparent.Right {
		tree.rotateRight(node.Parent)
		node = node.Right
	}
	tree.insertCase5(node)
}

func (tree *Tree) insertCase5(node *TreeNode) {
	node.Parent.color = true
	grandparent := node.grandparent()
	grandparent.color = false
	if node == node.Parent.Left && node.Parent == grandparent.Left {
		tree.rotateRight(grandparent)
	} else if node == node.Parent.Right && node.Parent == grandparent.Right {
		tree.rotateLeft(grandparent)
	}
}

func (node *TreeNode) maximumNode() *TreeNode {
	if node == nil {
		return nil
	}
	for node.Right != nil {
		node = node.Right
	}
	return node
}

func (tree *Tree) deleteCase1(node *TreeNode) {
	if node.Parent == nil {
		return
	}
	tree.deleteCase2(node)
}

func (tree *Tree) deleteCase2(node *TreeNode) {
	sibling := node.sibling()
	if nodeColor(sibling) == false {
		node.Parent.color = false
		sibling.color = true
		if node == node.Parent.Left {
			tree.rotateLeft(node.Parent)
		} else {
			tree.rotateRight(node.Parent)
		}
	}
	tree.deleteCase3(node)
}

func (tree *Tree) deleteCase3(node *TreeNode) {
	sibling := node.sibling()
	if nodeColor(node.Parent) == true &&
		nodeColor(sibling) == true &&
		nodeColor(sibling.Left) == true &&
		nodeColor(sibling.Right) == true {
		sibling.color = false
		tree.deleteCase1(node.Parent)
	} else {
		tree.deleteCase4(node)
	}
}

func (tree *Tree) deleteCase4(node *TreeNode) {
	sibling := node.sibling()
	if nodeColor(node.Parent) == false &&
		nodeColor(sibling) == true &&
		nodeColor(sibling.Left) == true &&
		nodeColor(sibling.Right) == true {
		sibling.color = false
		node.Parent.color = true
	} else {
		tree.deleteCase5(node)
	}
}

func (tree *Tree) deleteCase5(node *TreeNode) {
	sibling := node.sibling()
	if node == node.Parent.Left &&
		nodeColor(sibling) == true &&
		nodeColor(sibling.Left) == false &&
		nodeColor(sibling.Right) == true {
		sibling.color = false
		sibling.Left.color = true
		tree.rotateRight(sibling)
	} else if node == node.Parent.Right &&
		nodeColor(sibling) == true &&
		nodeColor(sibling.Right) == false &&
		nodeColor(sibling.Left) == true {
		sibling.color = false
		sibling.Right.color = true
		tree.rotateLeft(sibling)
	}
	tree.deleteCase6(node)
}

func (tree *Tree) deleteCase6(node *TreeNode) {
	sibling := node.sibling()
	sibling.color = nodeColor(node.Parent)
	node.Parent.color = true
	if node == node.Parent.Left && nodeColor(sibling.Right) == false {
		sibling.Right.color = true
		tree.rotateLeft(node.Parent)
	} else if nodeColor(sibling.Left) == false {
		sibling.Left.color = true
		tree.rotateRight(node.Parent)
	}
}

func nodeColor(node *TreeNode) bool {
	if node == nil {
		return true
	}
	return node.color
}

// Copyright (c) 2015, Emir Pasic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Adapted for code-gen by reddec (2019)

// Iterator holding the iterator's state
type TreeIterator struct {
	tree     *Tree
	node     *TreeNode
	position byte // begin = 0, between = 1, end = 2
}

// Iterator returns a stateful iterator whose elements are key/value pairs.
func (tree *Tree) Iterator() TreeIterator {
	return TreeIterator{tree: tree, node: nil, position: 0}
}

// Next moves the iterator to the next element and returns true if there was a next element in the container.
// If Next() returns true, then next element's key and value can be retrieved by Key() and Value().
// If Next() was called for the first time, then it will point the iterator to the first element if it exists.
// Modifies the state of the iterator.
func (iterator *TreeIterator) Next() bool {
	if iterator.position == 2 {
		goto end
	}
	if iterator.position == 0 {
		left := iterator.tree.Left()
		if left == nil {
			goto end
		}
		iterator.node = left
		goto between
	}
	if iterator.node.Right != nil {
		iterator.node = iterator.node.Right
		for iterator.node.Left != nil {
			iterator.node = iterator.node.Left
		}
		goto between
	}
	if iterator.node.Parent != nil {
		node := iterator.node
		for iterator.node.Parent != nil {
			iterator.node = iterator.node.Parent
			key := iterator.node.Key

			compare := node.Key - key

			if compare <= 0 {
				goto between
			}
		}
	}

end:
	iterator.node = nil
	iterator.position = 2
	return false

between:
	iterator.position = 1
	return true
}

// Prev moves the iterator to the previous element and returns true if there was a previous element in the container.
// If Prev() returns true, then previous element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (iterator *TreeIterator) Prev() bool {
	if iterator.position == 0 {
		goto begin
	}
	if iterator.position == 2 {
		right := iterator.tree.Right()
		if right == nil {
			goto begin
		}
		iterator.node = right
		goto between
	}
	if iterator.node.Left != nil {
		iterator.node = iterator.node.Left
		for iterator.node.Right != nil {
			iterator.node = iterator.node.Right
		}
		goto between
	}
	if iterator.node.Parent != nil {
		node := iterator.node
		for iterator.node.Parent != nil {
			iterator.node = iterator.node.Parent
			key := iterator.node.Key

			compare := node.Key - key

			if compare >= 0 {
				goto between
			}
		}
	}

begin:
	iterator.node = nil
	iterator.position = 0
	return false

between:
	iterator.position = 1
	return true
}

// Value returns the current element's value.
// Does not modify the state of the iterator.
func (iterator *TreeIterator) Value() *User {
	return iterator.node.Value
}

// Key returns the current element's key.
// Does not modify the state of the iterator.
func (iterator *TreeIterator) Key() int64 {
	return iterator.node.Key
}

// Begin resets the iterator to its initial state (one-before-first)
// Call Next() to fetch the first element if any.
func (iterator *TreeIterator) Begin() {
	iterator.node = nil
	iterator.position = 0
}

// End moves the iterator past the last element (one-past-the-end).
// Call Prev() to fetch the last element if any.
func (iterator *TreeIterator) End() {
	iterator.node = nil
	iterator.position = 2
}

// First moves the iterator to the first element and returns true if there was a first element in the container.
// If First() returns true, then first element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator
func (iterator *TreeIterator) First() bool {
	iterator.Begin()
	return iterator.Next()
}

// Last moves the iterator to the last element and returns true if there was a last element in the container.
// If Last() returns true, then last element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (iterator *TreeIterator) Last() bool {
	iterator.End()
	return iterator.Prev()
}
