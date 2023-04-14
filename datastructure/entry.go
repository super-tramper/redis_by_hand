package datastructure

import "redis_by_hand/datastructure/hashtable"

// Entry 侵入式数据结构, 将hashtable节点结构嵌入到有效载荷数据中
type Entry struct {
	Node hashtable.HNode
	Key  string
	Val  string
}
