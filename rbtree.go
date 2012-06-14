//
// Created by Yaz Saito on 06/10/12.
//

// A red-black tree with an API similar to C++ STL's.
//
// The implementation is inspired (read: stolen) from:
// http://en.literateprograms.org/Red-black_tree_(C)#chunk use:private function prototypes.
//
// A simple example is shown below. For more examples, see rbtree_test.go
//
// type MyItem struct {
// 	key   int
// 	value string
// }
//
// tree := NewTree(func(a, b Item) int { return a.(MyItem).key - b.(MyItem).key })
// tree.Insert(MyItem{10, "value10"})
// tree.Insert(MyItem{12, "value12"})
//
// fmt.Println("Get(10) ->", tree.Get(MyItem{10, ""}))
// fmt.Println("Get(11) ->", tree.Get(MyItem{11, ""}))
//
// // Find an element >= 11
// iter := tree.FindGE(MyItem{11, ""})
// fmt.Println("FindGE(11) ->", iter.Item())
//
// // Find an element >= 13
// iter = tree.FindGE(MyItem{13, ""})
// if !iter.End() { panic("There should be no element >= 13") }
//
// // Output:
// // Get(10) -> {10 value10}
// // Get(11) -> <nil>
// // FindGE(11) -> {12 value12}
//
package rbtree

//
// Public definitions
//

// Item is the object stored in each tree node.
type Item interface{}

// CompareFunc returns 0 if a==b, <0 if a<b, >0 if a>b.
type CompareFunc func(a, b Item) int

type Tree struct {
	root             *node
	minNode, maxNode *node
	count            int
	compare          CompareFunc
}

// Create a new empty tree.
func NewTree(compare CompareFunc) *Tree {
	r := new(Tree)
	r.compare = compare
	return r
}

// Return the number of elements in the tree.
func (root *Tree) Len() int {
	return root.count
}

// A convenience function for find the element equal to key. Return
// nil if not found.
func (root *Tree) Get(key Item) Item {
	n, exact := root.findGE(key)
	if exact {
		return n.item
	}
	return nil
}

// Create an iterator that points to the minimum item in the tree
func (root *Tree) Begin() Iterator {
	return Iterator{root, root.minNode}
}

// Create an iterator that points beyond the maximum item in the tree
func (root *Tree) End() Iterator {
	return Iterator{root, nil}
}

// Find the smallest element N such that N >= key, and return the
// iterator pointing to the element. If no such element is found,
// iter.End() becomes true.
func (root *Tree) FindGE(key Item) Iterator {
	n, _ := root.findGE(key)
	return Iterator{root, n}
}

// Find the largest element N such that N <= key, and return the
// iterator pointing to the element. If no such element is found,
// iter.End() becomes true.
func (root *Tree) FindLE(key Item) Iterator {
	n, exact := root.findGE(key)
	if exact {
		return Iterator{root, n}
	}
	if n != nil {
		return Iterator{root, n.prev()}
	}
	// return the max element
	n = root.root
	if n == nil {
		return Iterator{root, nil}
	}
	for n.right != nil {
		n = n.right
	}
	return Iterator{root, n}
}

// Insert an item. If the item is already in the tree, do nothing and
// return false. Else return true.
func (root *Tree) Insert(item Item) bool {
	n := new(node)
	n.item = item
	n.color = red

	// TODO: delay creating n until it is found to be inserted
	inserted := root.doInsert(n)
	if !inserted {
		return false
	}

	n.color = red

	for true {
		// Case 1: N is at the root
		if n.parent == nil {
			n.color = black
			break
		}

		// Case 2: The parent is black, so the tree already
		// satisfies the RB properties
		if n.parent.color == black {
			break
		}

		// Case 3: parent and uncle are both red.
		// Then paint both black and make grandparent red.
		grandparent := n.parent.parent
		var uncle *node
		if n.parent.isLeftChild() {
			uncle = grandparent.right
		} else {
			uncle = grandparent.left
		}
		if uncle != nil && uncle.color == red {
			n.parent.color = black
			uncle.color = black
			grandparent.color = red
			n = grandparent
			continue
		}

		// Case 4: parent is red, uncle is black (1)
		if n.isRightChild() && n.parent.isLeftChild() {
			root.rotateLeft(n.parent)
			n = n.left
			continue
		}
		if n.isLeftChild() && n.parent.isRightChild() {
			root.rotateRight(n.parent)
			n = n.right
			continue
		}

		// Case 5: parent is read, uncle is black (2)
		n.parent.color = black
		grandparent.color = red
		if n.isLeftChild() {
			root.rotateRight(grandparent)
		} else {
			root.rotateLeft(grandparent)
		}
		break
	}
	return true
}

// Delete an item with the given key. Return true iff the item was
// found.
func (root *Tree) DeleteWithKey(key Item) bool {
	iter := root.FindGE(key)
	if iter.node != nil {
		root.DeleteWithIterator(iter)
		return true
	}
	return false
}

// Delete the current item.
//
// REQUIRES: !iter.End()
func (root *Tree) DeleteWithIterator(iter Iterator) {
	doAssert(!iter.End())
	root.doDelete(iter.node)
}

type Iterator struct {
	root *Tree
	node *node
}

// Check if the iterator points beyond the max element
func (iter Iterator) End() bool {
	return iter.node == nil
}

// Check if the iterator points beyond the minimum element
func (iter Iterator) Begin() bool {
	return iter.node == iter.root.minNode
}

// Return the current element.
//
// REQUIRES: !iter.End()
func (iter Iterator) Item() interface{} {
	return iter.node.item
}

// Create a new iterator that points to the successor of the current node.
// If the original iterator already points to the maximum
// element in the tree, the returned iterator becomes End.
//
// REQUIRES: !iter.End()
func (iter Iterator) Next() Iterator {
	doAssert(!iter.End())
	return Iterator{iter.root, iter.node.next()}
}

// Create a new iterator that points to the predecessor of the current
// node.  If the original iterator already points to the minimum
// element in the tree, the returned iterator becomes End.
//
// REQUIRES: !iter.Begin()
func (iter Iterator) Prev() Iterator {
	doAssert(!iter.Begin())
	if iter.End() {
		doAssert(iter.root.Len() > 0)
		return Iterator{iter.root, iter.root.maxNode}
	}
	return Iterator{iter.root, iter.node.prev()}
}

func doAssert(b bool) {
	if !b {
		panic("rbtree internal assertion failed")
	}
}

const red = iota
const black = 1 + iota

type node struct {
	item                Item
	parent, left, right *node
	color               int // black or red
}

//
// Internal node attribute accessors
//
func getColor(n *node) int {
	if n == nil {
		return black
	}
	return n.color
}

func (n *node) isLeftChild() bool {
	return n == n.parent.left
}

func (n *node) isRightChild() bool {
	return n == n.parent.right
}

func (n *node) sibling() *node {
	doAssert(n.parent != nil)
	if n.isLeftChild() {
		return n.parent.right
	}
	return n.parent.left
}

// Return the minimum node that's larger than N. Return nil if no such
// node is found.
func (n *node) next() *node {
	if n.right != nil {
		m := n.right
		for m.left != nil {
			m = m.left
		}
		return m
	}

	for n != nil {
		p := n.parent
		if p == nil {
			return nil
		}
		if n.isLeftChild() {
			return p
		}
		n = p
	}
	return nil
}

// Return the maximum node that's smaller than N. Return nil if no
// such node is found.
func (n *node) prev() *node {
	if n.left != nil {
		return maxPredecessor(n)
	}

	for n != nil {
		p := n.parent
		if p == nil {
			return nil
		}
		if n.isRightChild() {
			return p
		}
		n = p
	}
	return nil
}

// Return the predecessor of "n".
func maxPredecessor(n *node) *node {
	doAssert(n.left != nil)
	m := n.left
	for m.right != nil {
		m = m.right
	}
	return m
}

//
// Tree methods
//

//
// Private methods
//

func (root *Tree) maybeSetMinNode(n *node) {
	if root.minNode == nil {
		root.minNode = n
		root.maxNode = n
	} else if root.compare(n.item, root.minNode.item) < 0 {
		root.minNode = n
	}
}

func (root *Tree) maybeSetMaxNode(n *node) {
	if root.maxNode == nil {
		root.minNode = n
		root.maxNode = n
	} else if root.compare(n.item, root.maxNode.item) > 0 {
		root.maxNode = n
	}
}

func (root *Tree) doInsert(n *node) bool {
	if root.root == nil {
		n.parent = nil
		root.root = n
		root.minNode = n
		root.maxNode = n
		root.count++
		return true
	}
	parent := root.root
	for true {
		comp := root.compare(n.item, parent.item)
		if comp == 0 {
			return false
		} else if comp < 0 {
			if parent.left == nil {
				n.parent = parent
				parent.left = n
				root.count++
				root.maybeSetMinNode(n)
				return true
			} else {
				parent = parent.left
			}
		} else {
			if parent.right == nil {
				n.parent = parent
				parent.right = n
				root.count++
				root.maybeSetMaxNode(n)
				return true
			} else {
				parent = parent.right
			}
		}
	}
	panic("should not reach here")
}

// Find a node whose item >= key. The 2nd return value is true iff the
// node.item==key. Returns (nil, false) if all nodes in the tree are <
// key.
func (root *Tree) findGE(key Item) (*node, bool) {
	n := root.root
	for true {
		if n == nil {
			return nil, false
		}
		comp := root.compare(key, n.item)
		if comp == 0 {
			return n, true
		} else if comp < 0 {
			if n.left != nil {
				n = n.left
			} else {
				return n, false
			}
		} else {
			if n.right != nil {
				n = n.right
			} else {
				succ := n.next()
				if succ == nil {
					return nil, false
				} else {
					comp = root.compare(key, succ.item)
					return succ, (comp == 0)
				}
			}
		}
	}
	panic("should not reach here")
}

// Delete N from the tree.
///The algorithm is largely stoler from
//
// http://en.literateprograms.org/red-black_tree_(C)#chunk use:private function prototypes
func (root *Tree) doDelete(n *node) {
	if root.minNode == n {
		root.minNode = nil
	}
	if root.maxNode == n {
		root.maxNode = nil
	}

	root.count--
/*
	if n.left != nil && n.right != nil {
		pred := maxPredecessor(n)
		n.item = pred.item
		n = pred
	}
*/

	if n.left != nil && n.right != nil {
		pred := maxPredecessor(n)
		doAssert(pred != n)
		isLeft := pred.isLeftChild()
		tmp := *pred
		root.replaceNode(n, pred)
		pred.color = n.color

		if (tmp.parent == n) {
			// swap the positions of n and pred
			if isLeft {
				pred.left = n
				pred.right = n.right
				if pred.right != nil {
					pred.right.parent = pred
				}
			} else {
				pred.left = n.left
				if pred.left != nil {
					pred.left.parent = pred
				}
				pred.right = n
			}
			n.item = tmp.item
			n.parent = pred

			n.left = tmp.left
			if n.left != nil {
				n.left.parent = n
			}
			n.right = tmp.right
			if n.right != nil {
				n.right.parent = n
			}
		} else {
			pred.left = n.left
			if pred.left != nil {
				pred.left.parent = pred
			}
			pred.right = n.right
			if pred.right != nil {
				pred.right.parent = pred
			}
			if isLeft {
				tmp.parent.left = n
			} else {
				tmp.parent.right = n
			}
			n.item = tmp.item
			n.parent = tmp.parent
			n.left = tmp.left
			if n.left != nil {
				n.left.parent = n
			}
			n.right = tmp.right
			if n.right != nil {
				n.right.parent = n
			}
		}
		n.color = tmp.color
	}

	doAssert(n.left == nil || n.right == nil)
	child := n.right
	if child == nil {
		child = n.left
	}
	if n.color == black {
		n.color = getColor(child)
		root.deleteCase1(n)
	}
	root.replaceNode(n, child)
	if n.parent == nil && child != nil {
		child.color = black
	}
	if root.count > 0 {
		if root.minNode == nil {
			root.minNode = root.root
			if root.minNode != nil {
				for root.minNode.left != nil {
					root.minNode = root.minNode.left
				}
			}
		}
		if root.maxNode == nil {
			root.maxNode = root.root
			if root.maxNode != nil {
				for root.maxNode.right != nil {
					root.maxNode = root.maxNode.right
				}
			}
		}
	}
}

func (root *Tree) deleteCase1(n *node) {
	for true {
		if n.parent != nil {
			if getColor(n.sibling()) == red {
				n.parent.color = red
				n.sibling().color = black
				if n == n.parent.left {
					root.rotateLeft(n.parent)
				} else {
					root.rotateRight(n.parent)
				}
			}
			if getColor(n.parent) == black &&
				getColor(n.sibling()) == black &&
				getColor(n.sibling().left) == black &&
				getColor(n.sibling().right) == black {
				n.sibling().color = red
				n = n.parent
				continue
			} else {
				// case 4
				if getColor(n.parent) == red &&
					getColor(n.sibling()) == black &&
					getColor(n.sibling().left) == black &&
					getColor(n.sibling().right) == black {
					n.sibling().color = red
					n.parent.color = black
				} else {
					root.deleteCase5(n)
				}
			}
		}
		break
	}
}

func (root *Tree) deleteCase5(n *node) {
	if n == n.parent.left &&
		getColor(n.sibling()) == black &&
		getColor(n.sibling().left) == red &&
		getColor(n.sibling().right) == black {
		n.sibling().color = red
		n.sibling().left.color = black
		root.rotateRight(n.sibling())
	} else if n == n.parent.right &&
		getColor(n.sibling()) == black &&
		getColor(n.sibling().right) == red &&
		getColor(n.sibling().left) == black {
		n.sibling().color = red
		n.sibling().right.color = black
		root.rotateLeft(n.sibling())
	}

	// case 6
	n.sibling().color = getColor(n.parent)
	n.parent.color = black
	if n == n.parent.left {
		doAssert(getColor(n.sibling().right) == red)
		n.sibling().right.color = black
		root.rotateLeft(n.parent)
	} else {
		doAssert(getColor(n.sibling().left) == red)
		n.sibling().left.color = black
		root.rotateRight(n.parent)
	}
}

func (root *Tree) replaceNode(oldn, newn *node) {
	if oldn.parent == nil {
		root.root = newn
	} else {
		if oldn == oldn.parent.left {
			oldn.parent.left = newn
		} else {
			oldn.parent.right = newn
		}
	}
	if newn != nil {
		newn.parent = oldn.parent
	}
}

/*
    X		     Y
  A   Y	    =>     X   C
     B C 	  A B
*/
func (root *Tree) rotateLeft(x *node) {
	y := x.right
	x.right = y.left
	if y.left != nil {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		root.root = y
	} else {
		if x.isLeftChild() {
			x.parent.left = y
		} else {
			x.parent.right = y
		}
	}
	y.left = x
	x.parent = y
}

/*
     Y           X
   X   C  =>   A   Y
  A B             B C
*/
func (root *Tree) rotateRight(y *node) {
	x := y.left

	// Move "B"
	y.left = x.right
	if x.right != nil {
		x.right.parent = y
	}

	x.parent = y.parent
	if y.parent == nil {
		root.root = x
	} else {
		if y.isLeftChild() {
			y.parent.left = x
		} else {
			y.parent.right = x
		}
	}
	x.right = y
	y.parent = x
}
