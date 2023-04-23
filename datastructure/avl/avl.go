package avl

import "redis_by_hand/tools"

type Node struct {
	Depth  uint32
	Cnt    uint32
	Left   *Node
	Right  *Node
	Parent *Node
}

func (n *Node) Init() {
	n.Depth = 1
	n.Cnt = 1
}

func Depth(n *Node) uint32 {
	if n != nil {
		return n.Depth
	}
	return 0
}

func Cnt(n *Node) uint32 {
	if n != nil {
		return n.Cnt
	}
	return 0
}

func (n *Node) Update() {
	n.Depth = 1 + tools.Max(Depth(n.Left), Depth(n.Right))
	n.Cnt = 1 + Cnt(n.Left) + Cnt(n.Right)
}

func RotLeft(n *Node) *Node {
	right := n.Right
	if right.Left != nil {
		right.Left.Parent = n
	}
	n.Right = right.Left
	right.Left = n
	right.Parent = n.Parent
	n.Parent = right
	n.Update()
	right.Update()
	return right
}

func RotRight(n *Node) *Node {
	left := n.Left
	if left.Right != nil {
		left.Right.Parent = n
	}
	n.Left = left.Right
	left.Right = n
	left.Parent = n.Parent
	n.Parent = left
	n.Update()
	left.Update()
	return left
}

// AVLFixLeft 左子树太深时调整左子树
func AVLFixLeft(root *Node) *Node {
	if root.Left != nil && Depth(root.Left.Left) < Depth(root.Left.Right) {
		root.Left = RotLeft(root.Left)
	}
	return RotRight(root)
}

func AVLFixRight(root *Node) *Node {
	if root.Right != nil && Depth(root.Right.Right) < Depth(root.Right.Left) {
		root.Right = RotRight(root.Right)
	}
	return RotLeft(root)
}

func AVLFix(n *Node) *Node {
	for {
		n.Update()
		l := Depth(n.Left)
		r := Depth(n.Right)
		var from **Node
		if n.Parent != nil {
			if n.Parent.Left == n {
				from = &n.Parent.Left
			} else {
				from = &n.Parent.Right
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
		n = n.Parent
	}
}

func AVLDel(n *Node) *Node {
	if n.Right == nil {
		// 无右子树，用左子树代替节点
		// 将左子树挂接到父节点
		parent := n.Parent
		if n.Left != nil {
			n.Left.Parent = parent
		}
		if parent != nil {
			if parent.Left == n {
				parent.Left = n.Left
			} else {
				parent.Right = n.Left
			}
			return AVLFix(parent)
		} else {
			return n.Left
		}
	} else {
		// 将该节点与它的下一个兄弟进行交换
		victim := n.Right
		for victim.Left != nil {
			victim = victim.Left
		}
		root := AVLDel(victim)
		*victim = *n
		if victim.Left != nil {
			victim.Left.Parent = victim
		}
		if victim.Right != nil {
			victim.Right.Parent = victim
		}
		parent := n.Parent
		if parent != nil {
			if parent.Left == n {
				parent.Left = victim
			} else {
				parent.Right = victim
			}
			return root
		} else {
			return victim
		}
	}
}

func AVLOffset(node *Node, offset int64) *Node {
	pos := int64(0)
	for offset != pos {
		if pos < offset && pos+int64(Cnt(node.Right)) >= offset {
			node = node.Right
			pos += int64(Cnt(node.Left)) + 1
		} else if pos > offset && pos-int64(Cnt(node.Left)) <= offset {
			node = node.Left
			pos -= int64(Cnt(node.Right)) + 1
		} else {
			parent := node.Parent
			if parent == nil {
				return nil
			}
			if parent.Right == node {
				pos -= int64(Cnt(node.Left)) + 1
			} else {
				pos += int64(Cnt(node.Right)) + 1
			}
			node = parent
		}
	}
	return node
}
