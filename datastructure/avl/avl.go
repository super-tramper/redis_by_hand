package avl

import "redis_by_hand/tools"

type Node struct {
	depth  uint32
	cnt    uint32
	left   *Node
	right  *Node
	parent *Node
}

func (n *Node) Init() {
	n.depth = 1
	n.cnt = 1
}

func Depth(n *Node) uint32 {
	if n != nil {
		return n.depth
	}
	return 0
}

func Cnt(n *Node) uint32 {
	if n != nil {
		return n.cnt
	}
	return 0
}

func (n *Node) Update() {
	n.depth = 1 + tools.Max(Depth(n.left), Depth(n.right))
	n.cnt = 1 + Cnt(n.left) + Cnt(n.right)
}

func RotLeft(n *Node) *Node {
	right := n.right
	if right.left != nil {
		right.left.parent = n
	}
	n.right = right.left
	right.left = n
	right.parent = n.parent
	n.parent = right
	n.Update()
	right.Update()
	return right
}

func RotRight(n *Node) *Node {
	left := n.left
	if left.right != nil {
		left.right.parent = n
	}
	n.left = left.right
	left.right = n
	left.parent = n.parent
	n.parent = left
	n.Update()
	left.Update()
	return left
}

// AVLFixLeft 左子树太深时调整左子树
func AVLFixLeft(root *Node) *Node {
	if root.left != nil && Depth(root.left.left) < Depth(root.left.right) {
		root.left = RotLeft(root.left)
	}
	return RotRight(root)
}

func AVLFixRight(root *Node) *Node {
	if root.right != nil && Depth(root.right.right) < Depth(root.right.left) {
		root.right = RotRight(root.right)
	}
	return RotLeft(root)
}

func AVLFix(n *Node) *Node {
	for {
		n.Update()
		l := Depth(n.left)
		r := Depth(n.right)
		var from **Node
		if n.parent != nil {
			if n.parent.left == n {
				from = &n.parent.left
			} else {
				from = &n.parent.right
			}
		}
		if l == r+2 {
			n = AVLFixLeft(n)
		} else if r == l+2 {
			n = AVLFixRight(n)
		}
		if from == nil {
			return n
		}
		*from = n
		n = n.parent
	}
}

func AVLDel(n *Node) *Node {
	if n.right == nil {
		// 无右子树，用左子树代替节点
		// 将左子树挂接到父节点
		parent := n.parent
		if n.left != nil {
			n.left.parent = parent
		}
		if parent != nil {
			if parent.left == n {
				parent.left = n.left
			} else {
				parent.right = n.left
			}
			return AVLFix(parent)
		} else {
			return n.left
		}
	} else {
		// 将该节点与它的下一个兄弟进行交换
		victim := n.right
		for victim.left != nil {
			victim = victim.left
		}
		root := AVLDel(victim)
		*victim = *n
		if victim.left != nil {
			victim.left.parent = victim
		}
		if victim.right != nil {
			victim.right.parent = victim
		}
		parent := n.parent
		if parent != nil {
			if parent.left == n {
				parent.left = victim
			} else {
				parent.right = victim
			}
			return root
		} else {
			return victim
		}
	}
}
