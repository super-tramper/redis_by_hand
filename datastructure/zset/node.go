package zset

import (
	"redis_by_hand/datastructure/avl"
	"redis_by_hand/datastructure/hashtable"
	"redis_by_hand/tools"
)

type ZNode struct {
	Tree   avl.Node
	Hmap   hashtable.HNode
	Score  float64
	Length uint32
	Name   *string
}

func ZNodeNew(name *string, length uint32, score float64) *ZNode {
	node := &ZNode{Score: score, Length: length, Name: name}
	node.Tree.Init()
	node.Hmap.Next = nil
	node.Hmap.HCode = tools.StrHash([]byte(*name), uint64(length))
	return node
}
