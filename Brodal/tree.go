package Brodal

import (
	"container/list"
)

type tree struct {
	root *node

	id uint

	rankPointersW []*node
	childrenRank  []*node

	numOfNodesInW []uint

	upperBoundGuide *guide
	lowerBoundGuide *guide
	mainTreeGuideW  *guide
}

type reduceType byte

const (
	linkReduce reduceType = 0
	cutReduce             = 1
)

const UPPER_BOUND int = 7
const LOWER_BOUND int = -2

func (tree *tree) rootRank() uint {
	return tree.root.rank
}

func (tree *tree) addViolation(bad *node) violationSetType {
	if tree.id == 1 {
		if bad.rank > tree.rootRank() {
			bad.violatingSelf = tree.root.vList.PushFront(bad)
			return vSet
		} else {
			if tree.mainTreeGuideW.boundArray[bad.rank].fst != 0 {
				bad.violatingSelf = tree.root.wList.InsertAfter(bad, tree.rankPointersW[bad.rank].violatingSelf)
			} else {
				bad.violatingSelf = tree.root.wList.PushFront(bad)
				tree.rankPointersW[bad.rank] = bad
			}

			tree.numOfNodesInW[bad.rank]++
			return wSet
		}
	}

	return error
}

func (tree *tree) children() *list.List {
	return tree.root.children
}


func (tree *tree) addRootChild(child *node) {
	tree.root.addChild(child, tree.childrenRank[child.rank])
	if child.rank < tree.root.rank-2 {
		// tree.lowerBoundGuide.update(-int(tree.root.numOfChildren[child.rank]), child.rank)
	}
}

func (tree *tree) addChildTo(parent *node, newChild *node, rightBrother *node) {
	if parent == tree.root {
		tree.Insert(newChild)
	} else {
		parent.addChild(newChild, rightBrother)
		if newChild.value < parent.value {
			if tree.id == 1 && newChild.rank < tree.rootRank() {
				tree.addViolation(newChild)
			}
		}
	}
}

func (tree *tree) RemoveNode(child *node) {
	if child.parent == tree.root {
		tree.cut(child)
	} else {
		child.parent.removeChild(child)
	}
}

func (tree *tree) removeRootChild(child *node) {
	tree.root.removeChild(child)
	if child.rank < tree.root.rank-2 {
		// tree.upperBoundGuide.update(int(tree.root.numOfChildren[child.rank]), child.rank)
	}
}

func (tree *tree) delinkFromRoot() ([]*node, uint) {
	return tree.root.delink()
}

func (tree *tree) Insert(node *node) {
	tree.addRootChild(node)

	if node.rank < tree.root.rank-2 {
		actions := tree.upperBoundGuide.forceIncrease(int(node.rank), int(tree.root.numOfChildren[node.rank]), 3)

		for _, act := range actions {
			tree.performeAction(node, act, linkReduce)
		}

	}

	tree.handleHighRank(tree.root.rank-2)
	tree.handleHighRank(tree.root.rank-1)

	if tree.id == 2{
		// TODO
	}
}

func (tree *tree) cut(node *node) {
	tree.removeRootChild(node)

	if node.rank < tree.root.rank-2 {
		if tree.childrenRank[node.rank].leftBrother().rank < node.rank+1 {
			panic("Don't know what to do")
		}

		reduceVal := 2
		if tree.childrenRank[node.rank].leftBrother().numOfChildren[node.rank] == 3 {
			reduceVal = 3
		}

		actions := tree.lowerBoundGuide.forceIncrease(int(node.rank), int(tree.root.numOfChildren[node.rank]), int(reduceVal))
		for _, act := range actions {
			tree.performeAction(node, act, cutReduce)
		}

	}

	tree.handleHighRank(tree.root.rank-2)
	tree.handleHighRank(tree.root.rank-1)
}

func (tree *tree) incRank(node1 *node, node2 *node) {
	if tree.rootRank() > node1.rank || tree.rootRank() > node2.rank {
		panic("Tree ranks don't match")
	}

	tree.root.incRank(node1, node2)
	tree.childrenRank = append(tree.childrenRank, node2)
	tree.rankPointersW = append(tree.rankPointersW, nil)
	tree.numOfNodesInW = append(tree.numOfNodesInW, 0)
}

func (tree *tree) performeAction(node *node, action action, reduceType reduceType) {
	if reduceType == linkReduce {
		tree.link(uint(action.index))

		// tree.lowerBoundGuide.update(-int(tree.root.numOfChildren[action.index]), uint(action.index))
		// tree.lowerBoundGuide.update(-int(tree.root.numOfChildren[action.index+1]), uint(action.index+1))
	} else {

		removedTree := tree.childrenRank[action.index]
		tree.removeRootChild(removedTree)

		nodes, _ := removedTree.delink()
		for _, n := range nodes {
			tree.Insert(n)
		}

		tree.Insert(removedTree)
	}
}

func (tree *tree) link(rank uint) {
	nodeX := tree.childrenRank[rank]
	nodeY, nodeZ := nodeX.rightBrother(), nodeX.rightBrother().rightBrother()

	if nodeZ.rightBrother().rank == rank {
		tree.childrenRank[rank] = nodeZ.rightBrother()
	} else {
		tree.childrenRank[rank] = nil
	}

	minNode, nodeX, nodeY := getMinNode(nodeX, nodeY, nodeZ)

	minNode.link(nodeX, nodeY)
}

func (tree *tree) handleHighRank(rank uint) {
	if tree.root.numOfChildren[rank] > 7 {
		nodeSliceX, _ := tree.root.delink()
		nodeSliceY, _ := tree.root.delink()
		nodeSliceZ, _ := tree.root.delink()

		nodeSliceX[0].incRank(nodeSliceX[1], nodeSliceY[0])
		nodeSliceY[1].incRank(nodeSliceZ[0], nodeSliceZ[1])

		if rank == tree.root.rank-1 {
			tree.incRank(nodeSliceX[0], nodeSliceY[0])
		} else {
			tree.Insert(nodeSliceX[0])
			tree.Insert(nodeSliceY[1])
		}
	} else if tree.root.numOfChildren[rank] < 2 {
		//...
	}
}

func (tree *tree) reduceViolaton(x1 *node, x2 *node) {
	if x1.isGood() || x2.isGood() {
		if x1.isGood() {
			x1.removeSelfFromViolating()
			tree.numOfNodesInW[x1.rank]--
		}
		if x2.isGood() {
			x2.removeSelfFromViolating()
			tree.numOfNodesInW[x2.rank]--
		}
	} else {
		if x1.parent != x2.parent {
			if x1.parent.value <= x2.parent.value {
				x1.swapBrothers(x2)
			} else {
				x2.swapBrothers(x1)
			}
		}
		tree.removeViolatingNode(x1, x2)
	}
}


func (tree *tree) removeViolatingNode(rmNode *node, otherBrother *node) {
	if tree.id == 1 {
		parent := rmNode.parent
		replacement := tree.childrenRank[parent.rank]
		grandParent := parent.parent
		if otherBrother == nil {
			otherBrother = func () *node {
				if rmNode.leftBrother().rank != rmNode.rank {
					return rmNode.rightBrother()
				} else {
					return rmNode.leftBrother()
				}
			}()
		}

		if parent.numOfChildren[rmNode.rank] == 2 {
			if parent.rank == rmNode.rank+1 {
				if parent != tree.root {
					tree.cut(replacement)
					tree.addChildTo(grandParent, replacement, parent)
					grandParent.removeChild(parent)
				} else {
					tree.cut(parent)
				}

				parent.removeChild(rmNode)
				parent.removeChild(otherBrother)
				tree.Insert(parent)
			} else {
				parent.removeChild(rmNode)
				parent.removeChild(otherBrother)
			}
			otherBrother.removeSelfFromViolating()
			tree.Insert(otherBrother)
			tree.Insert(rmNode)
		} else {
			parent.removeChild(rmNode)
			tree.Insert(rmNode)
		}
		rmNode.removeSelfFromViolating()
	}
}

// ######################################## UTIL #######################################

func newTree(value float64, treeIndex uint) *tree {

	tree := &tree{
		root:            newNode(value),
		id:              treeIndex,
		rankPointersW:   []*node{},
		childrenRank:    []*node{},
		numOfNodesInW:   []uint{},
		upperBoundGuide: newGuide(UPPER_BOUND),
		lowerBoundGuide: newGuide(LOWER_BOUND),
		mainTreeGuideW:  nil,
	}

	if treeIndex == 1 {
		tree.mainTreeGuideW = newGuide(6)
	}

	return tree
}

func getMinTree(tree1 *tree, tree2 *tree) (*tree, *tree) {
	if tree1 == nil || tree2 == nil {
		panic("One of the trees is nil")
	}

	if tree1.root.value > tree2.root.value {
		return tree2, tree1
	} else {
		return tree1, tree2
	}
}

func getMaxTree(trees []*tree) (*tree, []*tree) {
	if len(trees) == 0 {
		panic("There are no trees")
	}

	maxTree := trees[0]
	newTrees := [](*tree){}

	for _, tree := range trees {
		if tree != nil {
			if maxTree.root.rank < tree.root.rank {
				newTrees = append(newTrees, maxTree)
				maxTree = tree
			} else {
				newTrees = append(newTrees, tree)
			}
		}
	}

	return maxTree, newTrees
}

func mbySwapTree(ptr1 *tree, ptr2 *tree, cond bool) (*tree, *tree) {
	if cond {
		temp := ptr1
		ptr1 = ptr2
		ptr2 = temp
	}
	return ptr1, ptr2
}

