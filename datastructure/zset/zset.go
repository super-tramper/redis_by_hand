package zset

import (
	"redis_by_hand/datastructure/avl"
	"redis_by_hand/datastructure/hashtable"
	"redis_by_hand/tools"
	"unsafe"
)

type ZSet struct {
	Tree *avl.Node
	Hmap hashtable.HMap
}

func (zs *ZSet) TreeAdd(node *ZNode) {
	if zs.Tree == nil {
		zs.Tree = &node.Tree
	}

	cur := zs.Tree
	for {
		from := &(cur.Right)
		if ZLess(&node.Tree, cur) {
			from = &(cur.Left)
		}
		if *from == nil {
			*from = &(node.Tree)
			node.Tree.Parent = cur
			zs.Tree = avl.AVLFix(&node.Tree)
			break
		}
		cur = *from
	}
}

func ZLessScore(lhs *avl.Node, score float64, name *string, length uint32) bool {
	zl := (*ZNode)(unsafe.Pointer(lhs))
	if zl.Score != score {
		return zl.Score < score
	}
	rv := tools.Min(zl.Length, length)
	if string([]rune(*zl.Name)[:rv]) != string([]rune(*name)[:rv]) {
		return string([]rune(*zl.Name)[:rv]) < string([]rune(*name)[:rv])
	}
	return zl.Length < length
}

func ZLess(lhs *avl.Node, rhs *avl.Node) bool {
	zr := (*ZNode)(unsafe.Pointer(rhs))
	return ZLessScore(lhs, zr.Score, zr.Name, zr.Length)
}

func (zs *ZSet) Update(node *ZNode, score float64) {
	if node.Score == score {
		return
	}
	zs.Tree = avl.AVLDel(&node.Tree)
	node.Score = score
	node.Tree.Init()
	zs.TreeAdd(node)
}

func (zs *ZSet) Add(name *string, length uint32, score float64) bool {
	node := zs.Lookup(name, length)
	if node != nil {
		zs.Update(node, score)
		return false
	} else {
		node := ZNodeNew(name, length, score)
		zs.Hmap.Insert(&node.Hmap)
		zs.TreeAdd(node)
		return true
	}
}

type HKey struct {
	Node   hashtable.HNode
	Name   *string
	Length uint32
}

func HCmp(node *hashtable.HNode, key *hashtable.HNode) bool {
	if node.HCode != key.HCode {
		return false
	}
	zNode := (*ZNode)(unsafe.Pointer(uintptr(unsafe.Pointer(node)) - unsafe.Sizeof(avl.Node{})))
	hkey := (*HKey)(unsafe.Pointer(node))
	if zNode.Length != hkey.Length {
		return false
	}
	return string([]byte(*zNode.Name)[:zNode.Length]) == string([]byte(*hkey.Name)[:zNode.Length])
}

func (zs *ZSet) Lookup(name *string, length uint32) *ZNode {
	if zs.Tree == nil {
		return nil
	}

	var key HKey
	key.Node.HCode = tools.StrHash([]byte(*name), uint64(length))
	key.Name = name
	key.Length = length
	found := zs.Hmap.Lookup(&key.Node, HCmp)
	if found == nil {
		return nil
	}

	return (*ZNode)(unsafe.Pointer(uintptr(unsafe.Pointer(found)) - unsafe.Sizeof(avl.Node{})))
}

// Query 找到大于或等于参数的（score, name）元组、 然后相对于它进行偏移。
func (zs *ZSet) Query(score float64, name *string, length uint32, offset int64) *ZNode {
	var found *avl.Node
	cur := zs.Tree
	for cur != nil {
		if ZLessScore(cur, score, name, length) {
			cur = cur.Right
		} else {
			found = cur
			cur = cur.Left
		}
	}

	if found != nil {
		found = avl.AVLOffset(found, offset)
	}
	if found != nil {
		return (*ZNode)(unsafe.Pointer(found))
	}
	return nil
}
